package text

import (
	"fmt"

	"github.com/pkoukk/tiktoken-go"
	"github.com/tmc/langchaingo/textsplitter"
)

func NewTokenTextSplitter(tokenization Tokenization, opts ...textsplitter.Option) TextSplitter {
	options := textsplitter.DefaultOptions()

	for _, o := range opts {
		o(&options)
	}

	var sp TextSplitter
	switch tokenization.Choice.TokenizationMethod {
	case "Model":
		if modelInList(tokenization.Choice.Model, MistralModels) {
			sp = MistralSplitter{
				ChunkSize:         options.ChunkSize,
				ChunkOverlap:      options.ChunkOverlap,
				Model:             tokenization.Choice.Model,
				AllowedSpecial:    options.AllowedSpecial,
				DisallowedSpecial: options.DisallowedSpecial,
			}
		} else if modelInList(tokenization.Choice.Model, OpenAIModels) {
			sp = OpenAISplitter{
				ChunkSize:         options.ChunkSize,
				ChunkOverlap:      options.ChunkOverlap,
				Model:             tokenization.Choice.Model,
				AllowedSpecial:    options.AllowedSpecial,
				DisallowedSpecial: options.DisallowedSpecial,
			}
		} else if modelInList(tokenization.Choice.Model, CohereModels) {
			sp = CohereSplitter{
				ChunkSize:         options.ChunkSize,
				ChunkOverlap:      options.ChunkOverlap,
				Model:             tokenization.Choice.Model,
				AllowedSpecial:    options.AllowedSpecial,
				DisallowedSpecial: options.DisallowedSpecial,
			}
		}
	case "Encoding":
		sp = EncodingSplitter{
			ChunkSize:         options.ChunkSize,
			ChunkOverlap:      options.ChunkOverlap,
			Encoding:          tokenization.Choice.Encoding,
			AllowedSpecial:    options.AllowedSpecial,
			DisallowedSpecial: options.DisallowedSpecial,
		}
	}

	return sp
}

type OpenAISplitter struct {
	ChunkSize         int
	ChunkOverlap      int
	Model             string
	AllowedSpecial    []string
	DisallowedSpecial []string
}

func (sp OpenAISplitter) SplitText(text string) ([]string, error) {

	tk, err := tiktoken.EncodingForModel(sp.Model)

	if err != nil {
		return nil, fmt.Errorf("tiktoken.EncodingForModel: %w", err)
	}

	texts := splitText(text, tk, sp.AllowedSpecial, sp.DisallowedSpecial, sp.ChunkSize, sp.ChunkOverlap)

	return texts, nil
}

type EncodingSplitter struct {
	ChunkSize         int
	ChunkOverlap      int
	Encoding          string
	AllowedSpecial    []string
	DisallowedSpecial []string
}

func (sp EncodingSplitter) SplitText(text string) ([]string, error) {

	tk, err := tiktoken.GetEncoding(sp.Encoding)

	if err != nil {
		return nil, fmt.Errorf("tiktoken.GetEncoding: %w", err)
	}

	texts := splitText(text, tk, sp.AllowedSpecial, sp.DisallowedSpecial, sp.ChunkSize, sp.ChunkOverlap)

	return texts, nil
}

type CohereSplitter struct {
	ChunkSize         int
	ChunkOverlap      int
	Model             string
	AllowedSpecial    []string
	DisallowedSpecial []string
}

func (sp CohereSplitter) SplitText(text string) ([]string, error) {

	return nil, fmt.Errorf("CohereSplitter not implemented yet")
}

type MistralSplitter struct {
	ChunkSize         int
	ChunkOverlap      int
	Model             string
	AllowedSpecial    []string
	DisallowedSpecial []string
}

func (sp MistralSplitter) SplitText(text string) ([]string, error) {
	return nil, fmt.Errorf("MistralSplitter not implemented yet")
}

func splitText(text string, tk *tiktoken.Tiktoken, allowSpecial, disallowedSpecial []string, chunkSize, chunkOverlap int) []string {
	splits := make([]string, 0)
	inputIDs := tk.Encode(text, allowSpecial, disallowedSpecial)

	startIdx := 0
	curIdx := len(inputIDs)
	if startIdx+chunkSize < curIdx {
		curIdx = startIdx + chunkSize
	}
	for startIdx < len(inputIDs) {
		chunkIDs := inputIDs[startIdx:curIdx]
		splits = append(splits, tk.Decode(chunkIDs))
		startIdx += chunkSize - chunkOverlap
		curIdx = startIdx + chunkSize
		if curIdx > len(inputIDs) {
			curIdx = len(inputIDs)
		}
	}
	return splits
}
