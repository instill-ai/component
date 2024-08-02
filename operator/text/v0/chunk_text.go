package text

import (
	"fmt"

	"github.com/instill-ai/component/base"
	"github.com/tmc/langchaingo/textsplitter"
	"google.golang.org/protobuf/types/known/structpb"
)

type ChunkTextInput struct {
	Text         string       `json:"text"`
	Strategy     Strategy     `json:"strategy"`
	Tokenization Tokenization `json:"tokenization"`
}

type Strategy struct {
	Setting Setting `json:"setting"`
}

type Setting struct {
	ChunkMethod       string   `json:"chunk-method,omitempty"`
	ChunkSize         int      `json:"chunk-size,omitempty"`
	ChunkOverlap      int      `json:"chunk-overlap,omitempty"`
	ModelName         string   `json:"model-name,omitempty"`
	AllowedSpecial    []string `json:"allowed-special,omitempty"`
	DisallowedSpecial []string `json:"disallowed-special,omitempty"`
	Separators        []string `json:"separators,omitempty"`
	KeepSeparator     bool     `json:"keep-separator,omitempty"`
	CodeBlocks        bool     `json:"code-blocks,omitempty"`
	// TODO: Add SecondSplitter, which is to set the details about how to chunk the paragraphs in Markdown format.
	// https://pkg.go.dev/github.com/tmc/langchaingo@v0.1.10/textsplitter#MarkdownTextSplitter
	// secondSplitter textsplitter.TextSplitter
}

type Tokenization struct {
	Choice Choice `json:"choice"`
}

type Choice struct {
	TokenizationMethod   string `json:"tokenization-method"`
	ModelName            string `json:"model-name,omitempty"`
	EncodingName         string `json:"encoding-name,omitempty"`
	HuggingFaceModelName string `json:"hugging-face-model-name,omitempty"`
}

type ChunkTextOutput struct {
	ChunkNum         int         `json:"chunk-num"`
	TextChunks       []TextChunk `json:"text-chunks"`
	TokenCount       int         `json:"token-count"`
	ChunksTokenCount int         `json:"chunks-token-count"`
}

type TextChunk struct {
	Text          string `json:"text"`
	StartPosition int    `json:"start-position"`
	EndPosition   int    `json:"end-position"`
	TokenCount    int    `json:"token-count"`
}

func (s *Setting) SetDefault() {
	if s.ChunkSize == 0 {
		s.ChunkSize = 512
	}
	if s.ChunkOverlap == 0 {
		s.ChunkOverlap = 100
	}
	if s.ModelName == "" {
		s.ModelName = "gpt-3.5-turbo"
	}
	if s.AllowedSpecial == nil {
		s.AllowedSpecial = []string{}
	}
	if s.DisallowedSpecial == nil {
		s.DisallowedSpecial = []string{"all"}
	}
	if s.Separators == nil {
		s.Separators = []string{"\n\n", "\n", " ", ""}
	}
}

type TextSplitter interface {
	SplitText(text string) ([]string, error)
}

// ChunkText do 3 blocks of work:
// 1. Split Text
// 2. Positioning the chunks
// 3. Tokenize the chunks
// Because we will need to tokenize the chunks with Python code,
// it will be better to pass whole chunks to Python code and tokenize them there.
func chunkText(inputPb *structpb.Struct) (*structpb.Struct, error) {
	input := ChunkTextInput{}

	err := base.ConvertFromStructpb(inputPb, &input)
	if err != nil {
		return nil, err
	}

	var split TextSplitter
	setting := input.Strategy.Setting
	// TODO: Take this out when we fix the error in frontend side.
	// Bug: The default value is not set from frontend side.
	setting.SetDefault()

	var output ChunkTextOutput
	switch setting.ChunkMethod {
	case "Token":
		if setting.ChunkOverlap >= setting.ChunkSize {
			return nil, fmt.Errorf("ChunkOverlap must be less than ChunkSize when using Token method")
		}

		split = textsplitter.NewTokenSplitter(
			textsplitter.WithChunkSize(setting.ChunkSize),
			textsplitter.WithChunkOverlap(setting.ChunkOverlap),
			textsplitter.WithModelName(setting.ModelName),
			textsplitter.WithAllowedSpecial(setting.AllowedSpecial),
			textsplitter.WithDisallowedSpecial(setting.DisallowedSpecial),
		)
	case "Markdown":
		split = NewMarkdownTextSplitter(
			textsplitter.WithChunkSize(setting.ChunkSize),
			textsplitter.WithChunkOverlap(setting.ChunkOverlap),
		)
	case "Recursive":
		split = textsplitter.NewRecursiveCharacter(
			textsplitter.WithSeparators(setting.Separators),
			textsplitter.WithChunkSize(setting.ChunkSize),
			textsplitter.WithChunkOverlap(setting.ChunkOverlap),
			textsplitter.WithKeepSeparator(setting.KeepSeparator),
		)
	}

	chunks, err := split.SplitText(input.Text)
	if err != nil {
		return nil, fmt.Errorf("failed to split text: %w", err)
	}

	output.setChunksWithPosition(chunks, input.Text, setting.ChunkMethod)
	output.ChunkNum = len(output.TextChunks)

	// TODO: chung8511, implement the tokenizer in Python code.
	// originalTextToken := tkm.Encode(input.Text, setting.AllowedSpecial, setting.DisallowedSpecial)
	// output.TokenCount = len(originalTextToken)
	// output.ChunksTokenCount = totalTokenCount

	outputPb, err := base.ConvertToStructpb(output)
	if err != nil {
		return nil, fmt.Errorf("failed to convert output to structpb: %w", err)
	}

	return outputPb, nil
}
