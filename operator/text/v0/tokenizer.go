package text

import (
	"fmt"

	"github.com/pkoukk/tiktoken-go"
)

type Tokenizer interface {
	Encode(chunks []TextChunk) (map[int]int, error)
	// TODO: chuang8511 need to add encode for token chunk strategy
	// EncodeTokenChunk(chunks string) ([]string, error)
}

type OpenAITokenizer struct {
	model string
}
type MistralTokenizer struct {
	model string
}
type CohereTokenizer struct {
	model string
}
type EncodingTokenizer struct {
	encoding string
}
type HuggingFaceTokenizer struct {
	model string
}

func (choice Choice) GetTokenizer() (Tokenizer, error) {
	switch choice.TokenizationMethod {
	case "Model":
		return getModelTokenizer(choice.Model)
	case "Encoding":
		return EncodingTokenizer{
			encoding: choice.Encoding,
		}, nil
	case "HuggingFace":
		return HuggingFaceTokenizer{
			model: choice.HuggingFaceModel,
		}, nil
	}
	return nil, fmt.Errorf("Tokenization method %s not found", choice.TokenizationMethod)
}

func getModelTokenizer(model string) (Tokenizer, error) {
	if modelInList(model, MistralModels) {
		return MistralTokenizer{
			model: model,
		}, nil
	}
	if modelInList(model, OpenAIModels) {
		return OpenAITokenizer{
			model: model,
		}, nil
	}
	if modelInList(model, CohereModels) {
		return CohereTokenizer{
			model: model,
		}, nil
	}
	return nil, fmt.Errorf("Model %s not found", model)
}

func (t OpenAITokenizer) Encode(textChunks []TextChunk) (map[int]int, error) {
	tke, err := tiktoken.EncodingForModel(t.model)
	if err != nil {
		return map[int]int{}, fmt.Errorf("Failed to get encoding by model name %s: %w", t.model, err)
	}

	tokenIdxCountMap := make(map[int]int)

	for i, textChunk := range textChunks {
		tokenCount := len(tke.Encode(textChunk.Text, nil, nil))
		tokenIdxCountMap[i] = tokenCount
	}

	return tokenIdxCountMap, nil
}

func (t EncodingTokenizer) Encode(textChunks []TextChunk) (map[int]int, error) {
	tke, err := tiktoken.GetEncoding(t.encoding)
	if err != nil {
		return map[int]int{}, fmt.Errorf("Failed to get encoding by encoding name %s: %w", t.encoding, err)
	}

	tokenIdxCountMap := make(map[int]int)

	for i, textChunk := range textChunks {
		tokenCount := len(tke.Encode(textChunk.Text, nil, nil))
		tokenIdxCountMap[i] = tokenCount
	}

	return tokenIdxCountMap, nil
}

func (t MistralTokenizer) Encode(textChunks []TextChunk) (map[int]int, error) {
	return map[int]int{}, nil
}

func (t CohereTokenizer) Encode(textChunks []TextChunk) (map[int]int, error) {
	return map[int]int{}, nil
}

func (t HuggingFaceTokenizer) Encode(textChunks []TextChunk) (map[int]int, error) {
	return map[int]int{}, nil
}

func (output *ChunkTextOutput) setTokenizeChunks(choice Choice) error {
	tokenizer, err := choice.GetTokenizer()

	if err != nil {
		return fmt.Errorf("Failed to get tokenizer: %w", err)
	}

	tokenMap, err := tokenizer.Encode(output.TextChunks)

	if err != nil {
		return fmt.Errorf("Failed to encode text: %w", err)
	}

	for i, tokenCount := range tokenMap {
		output.TextChunks[i].TokenCount = tokenCount
		output.ChunksTokenCount += tokenCount
	}

	return nil
}

func (output *ChunkTextOutput) setFileTokenCount(choice Choice, rawText string) error {
	tokenizer, err := choice.GetTokenizer()

	if err != nil {
		return fmt.Errorf("Failed to get tokenizer: %w", err)
	}

	tokenMap, err := tokenizer.Encode([]TextChunk{
		{
			Text: rawText,
		},
	})

	if err != nil {
		return fmt.Errorf("Failed to encode text: %w", err)
	}

	output.TokenCount = tokenMap[0]

	return nil
}

var OpenAIModels = []string{
	"gpt-4o",
	"gpt-4",
	"gpt-3.5-turbo",
	"text-davinci-003",
	"text-davinci-002",
	"text-davinci-001",
	"text-curie-001",
	"text-babbage-001",
	"text-ada-001",
	"davinci",
	"curie",
	"babbage",
	"ada",
	"code-davinci-002",
	"code-davinci-001",
	"code-cushman-002",
	"code-cushman-001",
	"davinci-codex",
	"cushman-codex",
	"text-davinci-edit-001",
	"code-davinci-edit-001",
	"text-embedding-ada-002",
	"text-similarity-davinci-001",
	"text-similarity-curie-001",
	"text-similarity-babbage-001",
	"text-similarity-ada-001",
	"text-search-davinci-doc-001",
	"text-search-curie-doc-001",
	"text-search-babbage-doc-001",
	"text-search-ada-doc-001",
	"code-search-babbage-code-001",
	"code-search-ada-code-001",
	"gpt2",
}

var MistralModels = []string{
	"open-mixtral-8x22b",
	"open-mixtral-8x7b",
	"open-mistral-7b",
	"mistral-large-latest",
	"mistral-small-latest",
	"codestral-latest",
	"mistral-embed",
}

var CohereModels = []string{
	"command-r-plus",
	"command-r",
	"command",
	"command-nightly",
	"command-light",
	"command-light-nightly",
	"embed-english-v3.0",
	"embed-multilingual-v3.0",
	"embed-english-light-v3.0",
	"embed-multilingual-light-v3.0",
}

func modelInList(model string, list []string) bool {
	for _, m := range list {
		if m == model {
			return true
		}
	}
	return false
}
