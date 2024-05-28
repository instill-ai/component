package text

import (
	"fmt"

	"github.com/instill-ai/component/pkg/external/langchaingo/textsplitter"
	"github.com/pkoukk/tiktoken-go"
)

type ChunkTextInput struct {
	Text     string   `json:"text"`
	Strategy Strategy `json:"strategy"`
}

type Strategy struct {
	Setting Setting `json:"setting"`
}

type Setting struct {
	ChunkMethod       string   `json:"chunk_method,omitempty"`
	ChunkSize         int      `json:"chunk_size,omitempty"`
	ChunkOverlap      int      `json:"chunk_overlap,omitempty"`
	ModelName         string   `json:"model_name,omitempty"`
	EncodingName      string   `json:"encoding_name,omitempty"`
	AllowedSpecial    []string `json:"allowed_special,omitempty"`
	DisallowedSpecial []string `json:"disallowed_special,omitempty"`
	Separators        []string `json:"separators,omitempty"`
	KeepSeparator     bool     `json:"keep_separator,omitempty"`
	CodeBlocks        bool     `json:"code_blocks,omitempty"`
	ReferenceLinks    bool     `json:"reference_links,omitempty"`
	// TODO: Add SecondSplitter, which is to set the details about how to chunk the paragraphs in Markdown format.
	// https://pkg.go.dev/github.com/tmc/langchaingo@v0.1.10/textsplitter#MarkdownTextSplitter
	// secondSplitter textsplitter.TextSplitter
}

type ChunkTextOutput struct {
	ChunkNum   int         `json:"chunk_num"`
	TextChunks []TextChunk `json:"text_chunks"`
	TokenCount int         `json:"token_count,omitempty"`
}

type TextChunk struct {
	Text          string `json:"text"`
	StartPosition int    `json:"start_position,omitempty"`
	EndPosition   int    `json:"end_position,omitempty"`
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
	if s.EncodingName == "" {
		s.EncodingName = "cl100k_base"
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

func chunkText(input ChunkTextInput) (ChunkTextOutput, error) {
	var split textsplitter.TextSplitter
	setting := input.Strategy.Setting
	setting.SetDefault()
	switch setting.ChunkMethod {
	case "Token":

		if setting.ChunkOverlap >= setting.ChunkSize {
			err := fmt.Errorf("ChunkOverlap must be less than ChunkSize when using Token method.")
			return ChunkTextOutput{}, err
		}

		split = textsplitter.NewTokenSplitter(
			textsplitter.WithChunkSize(setting.ChunkSize),
			textsplitter.WithChunkOverlap(setting.ChunkOverlap),
			textsplitter.WithModelName(setting.ModelName),
			textsplitter.WithEncodingName(setting.EncodingName),
			textsplitter.WithAllowedSpecial(setting.AllowedSpecial),
			textsplitter.WithDisallowedSpecial(setting.DisallowedSpecial),
		)
	case "Markdown":
		split = textsplitter.NewMarkdownTextSplitter(
			textsplitter.WithChunkSize(setting.ChunkSize),
			textsplitter.WithChunkOverlap(setting.ChunkOverlap),
			textsplitter.WithCodeBlocks(setting.CodeBlocks),
			textsplitter.WithReferenceLinks(setting.ReferenceLinks),
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
		return ChunkTextOutput{}, err
	}

	output := ChunkTextOutput{
		ChunkNum: len(chunks),
	}

	if setting.ChunkMethod == "Token" {
		tkm, err := tiktoken.EncodingForModel(setting.ModelName)
		if err != nil {
			return ChunkTextOutput{}, err
		}
		token := tkm.Encode(input.Text, setting.AllowedSpecial, setting.DisallowedSpecial)
		output.TokenCount = len(token)
	}

	startPosition := 1
	for _, c := range chunks {
		output.TextChunks = append(output.TextChunks, TextChunk{
			Text:          c,
			StartPosition: startPosition,
			EndPosition:   startPosition + len(c) - 1,
		})
		startPosition += len(c)
	}

	return output, nil
}
