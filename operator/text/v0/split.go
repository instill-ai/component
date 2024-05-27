package text

import (
	"github.com/pkoukk/tiktoken-go"
)

const defaultChunkTokenSize = 500

// SplitByTokenInput defines the input for split by token task
type SplitByTokenInput struct {
	// Text: Text to split
	Text string `json:"text"`
	// Model: ID of the model to use for tokenization
	Model string `json:"model"`
	// ChunkTokenSize: Number of tokens per text chunk
	ChunkTokenSize *int `json:"chunk_token_size,omitempty"`
}

// SplitByTokenOutput defines the output for split by token task
type SplitByTokenOutput struct {
	// TokenCount: Number of tokens in the text
	TokenCount int `json:"token_count"`
	// TextChunks: List of text chunks
	TextChunks []string `json:"text_chunks"`
	// ChunkNum: Number of text chunks
	ChunkNum int `json:"chunk_num"`
}

// splitTextIntoChunks splits text into text chunks based on token size
func splitTextIntoChunks(input SplitByTokenInput) (SplitByTokenOutput, error) {
	output := SplitByTokenOutput{}

	if input.ChunkTokenSize == nil || *input.ChunkTokenSize <= 0 {
		input.ChunkTokenSize = new(int)
		*input.ChunkTokenSize = defaultChunkTokenSize
	}

	tkm, err := tiktoken.EncodingForModel(input.Model)
	if err != nil {
		return output, err
	}

	token := tkm.Encode(input.Text, nil, nil)
	output.TokenCount = len(token)
	output.TextChunks = []string{}
	for start := 0; start < len(token); start += *input.ChunkTokenSize {
		end := min(start+*input.ChunkTokenSize, len(token))
		output.TextChunks = append(output.TextChunks, tkm.Decode(token[start:end]))
	}
	output.ChunkNum = len(output.TextChunks)
	return output, nil
}
