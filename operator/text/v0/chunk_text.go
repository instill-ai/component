package text

import (
	"github.com/instill-ai/component/pkg/external/langchaingo/textsplitter"
	"github.com/pkoukk/tiktoken-go"
)

type ChunkTextInput struct {
	Text              string   `json:"text"`
	ChunkStrategy     string   `json:"chunk_strategy"`
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
	Chunks     []Chunk `json:"chunks"`
	ChunkNum   int     `json:"chunk_num"`
	TokenCount int     `json:"token_count,omitempty"`
}

type Chunk struct {
	Text          string `json:"text"`
	StartPosition int    `json:"start_position,omitempty"`
	EndPosition   int    `json:"end_position,omitempty"`
}

func chunkText(input ChunkTextInput) (ChunkTextOutput, error) {

	var split textsplitter.TextSplitter
	switch input.ChunkStrategy {
	// TODO: default write in json file to reduce the complexity.
	case "token":
		split = textsplitter.NewTokenSplitter(
			textsplitter.WithChunkSize(input.ChunkSize),
			textsplitter.WithChunkOverlap(input.ChunkOverlap),
			textsplitter.WithModelName(input.ModelName),
			textsplitter.WithEncodingName(input.EncodingName),
			textsplitter.WithAllowedSpecial(input.AllowedSpecial),
			textsplitter.WithDisallowedSpecial(input.DisallowedSpecial),
		)
	case "markdown":
		split = textsplitter.NewMarkdownTextSplitter(
			textsplitter.WithChunkSize(input.ChunkSize),
			textsplitter.WithChunkOverlap(input.ChunkOverlap),
			textsplitter.WithCodeBlocks(input.CodeBlocks),
			textsplitter.WithReferenceLinks(input.ReferenceLinks),
		)
	case "recursive":
		split = textsplitter.NewRecursiveCharacter(
			textsplitter.WithSeparators(input.Separators),
			textsplitter.WithChunkSize(input.ChunkSize),
			textsplitter.WithChunkOverlap(input.ChunkOverlap),
			textsplitter.WithKeepSeparator(input.KeepSeparator),
		)
	}

	chunks, err := split.SplitText(input.Text)
	if err != nil {
		return ChunkTextOutput{}, err
	}

	output := ChunkTextOutput{
		ChunkNum: len(chunks),
	}

	// To remain the original output
	if input.ChunkStrategy == "token" {
		tkm, err := tiktoken.EncodingForModel(input.ModelName)
		if err != nil {
			return ChunkTextOutput{}, err
		}
		token := tkm.Encode(input.Text, input.AllowedSpecial, input.DisallowedSpecial)
		output.TokenCount = len(token)
	}

	startPosition := 0
	for _, c := range chunks {
		output.Chunks = append(output.Chunks, Chunk{
			Text:          c,
			StartPosition: startPosition,
			EndPosition:   startPosition + len(c) - 1,
		})
		startPosition += len(c)
	}

	return output, nil
}
