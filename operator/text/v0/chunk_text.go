package text

import (
	"fmt"
	"reflect"

	"github.com/pkoukk/tiktoken-go"
	"github.com/tmc/langchaingo/textsplitter"
)

type ChunkTextInput struct {
	Text     string   `json:"text"`
	Strategy Strategy `json:"strategy"`
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
	ReferenceLinks    bool     `json:"reference-links,omitempty"`
	ShowTokenCount    bool     `json:"show-token-count"`
	// TODO: Add SecondSplitter, which is to set the details about how to chunk the paragraphs in Markdown format.
	// https://pkg.go.dev/github.com/tmc/langchaingo@v0.1.10/textsplitter#MarkdownTextSplitter
	// secondSplitter textsplitter.TextSplitter
}

type ChunkTextOutput struct {
	ChunkNum   int         `json:"chunk-num"`
	TextChunks []TextChunk `json:"text-chunks"`
	TokenCount int         `json:"token-count,omitempty"`
}

type TextChunk struct {
	Text          string `json:"text"`
	StartPosition int    `json:"start-position,omitempty"`
	EndPosition   int    `json:"end-position,omitempty"`
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

func chunkText(input ChunkTextInput) (ChunkTextOutput, error) {
	var split textsplitter.TextSplitter
	setting := input.Strategy.Setting
	// TODO: Take this out when we fix the error in frontend side.
	// Bug: The default value is not set from frontend side.
	setting.SetDefault()

	var output ChunkTextOutput
	var positionCalculator ChunkPositionCalculator

	switch setting.ChunkMethod {
	case "Token":
		positionCalculator = PositionCalculator{}
		if setting.ChunkOverlap >= setting.ChunkSize {
			err := fmt.Errorf("ChunkOverlap must be less than ChunkSize when using Token method")
			return output, err
		}

		split = textsplitter.NewTokenSplitter(
			textsplitter.WithChunkSize(setting.ChunkSize),
			textsplitter.WithChunkOverlap(setting.ChunkOverlap),
			textsplitter.WithModelName(setting.ModelName),
			textsplitter.WithAllowedSpecial(setting.AllowedSpecial),
			textsplitter.WithDisallowedSpecial(setting.DisallowedSpecial),
		)
	case "Markdown":
		positionCalculator = MarkdownPositionCalculator{}
		split = textsplitter.NewMarkdownTextSplitter(
			textsplitter.WithChunkSize(setting.ChunkSize),
			textsplitter.WithChunkOverlap(setting.ChunkOverlap),
			textsplitter.WithCodeBlocks(setting.CodeBlocks),
			textsplitter.WithReferenceLinks(setting.ReferenceLinks),
		)
	case "Recursive":
		positionCalculator = PositionCalculator{}
		split = textsplitter.NewRecursiveCharacter(
			textsplitter.WithSeparators(setting.Separators),
			textsplitter.WithChunkSize(setting.ChunkSize),
			textsplitter.WithChunkOverlap(setting.ChunkOverlap),
			textsplitter.WithKeepSeparator(setting.KeepSeparator),
		)
	}

	if setting.ShowTokenCount {
		tkm, err := tiktoken.EncodingForModel(setting.ModelName)
		if err != nil {
			return output, err
		}
		token := tkm.Encode(input.Text, setting.AllowedSpecial, setting.DisallowedSpecial)
		output.TokenCount = len(token)
	}

	chunks, err := split.SplitText(input.Text)
	if err != nil {
		return output, err
	}
	output.ChunkNum = len(chunks)

	startScanPosition := 0
	rawRunes := []rune(input.Text)
	for _, chunk := range chunks {
		chunkRunes := []rune(chunk)

		startPosition, endPosition := positionCalculator.getChunkPositions(rawRunes, chunkRunes, startScanPosition)

		output.TextChunks = append(output.TextChunks, TextChunk{
			Text:          chunk,
			StartPosition: startPosition,
			EndPosition:   endPosition,
		})
		startScanPosition = startPosition + 1
	}

	return output, nil
}

type ChunkPositionCalculator interface {
	getChunkPositions(rawText, chunk []rune, startScanPosition int) (startPosition int, endPosition int)
}

type PositionCalculator struct{}

func (PositionCalculator) getChunkPositions(rawText, chunk []rune, startScanPosition int) (startPosition int, endPosition int) {

	for i := startScanPosition; i < len(rawText); i++ {
		if rawText[i] == chunk[0] {
			if reflect.DeepEqual(rawText[i:i+len(chunk)], chunk) {
				startPosition = i
				endPosition = len(chunk) + i - 1
				break
			}
		}
	}
	return startPosition, endPosition
}

type MarkdownPositionCalculator struct{}

func (MarkdownPositionCalculator) getChunkPositions(rawText, chunk []rune, startScanPosition int) (startPosition int, endPosition int) {

	skipHeaderIndex := getSkipHeaderIndex(chunk)

	for i := startScanPosition; i < len(rawText); i++ {

		if rawText[i] == chunk[skipHeaderIndex] {
			if reflect.DeepEqual(rawText[i:(i+len(chunk)-skipHeaderIndex)], chunk[skipHeaderIndex:]) {
				startPosition = i
				endPosition = len(chunk) + i - 1 - skipHeaderIndex
				break
			}
		}
	}
	return startPosition, endPosition
}

func getSkipHeaderIndex(chunk []rune) int {
	hashtagCount := 0
	for i := 0; i < len(chunk); i++ {
		if chunk[i] == '#' {
			hashtagCount++
		}

		if hashtagCount >= 1 && chunk[i] == '\n' {
			return i + 1
		}
	}
	return 0
}
