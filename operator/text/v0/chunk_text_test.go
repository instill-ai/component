package text

import (
	"testing"

	"github.com/frankban/quicktest"
)

func TestChunkText(t *testing.T) {

	c := quicktest.New(t)

	testCases := []struct {
		name   string
		input  ChunkTextInput
		output ChunkTextOutput
	}{
		{
			name: "chunk text by token",
			input: ChunkTextInput{
				Text: "Hello world.",
				Strategy: Strategy{
					Setting: Setting{
						ChunkMethod: "Token",
						ChunkSize:   512,
						ModelName:   "gpt-3.5-turbo",
					},
				},
			},
			output: ChunkTextOutput{
				TextChunks: []TextChunk{
					{
						Text:          "Hello world.",
						StartPosition: 1,
						EndPosition:   12,
					},
				},
				ChunkNum: 1,
				// TODO: minimock failed to generate the mock for tiktoken.EncodingForModel
				// Please mock the tiktoken.EncodingForModel function to return the following value
				TokenCount: 3,
			},
		},
		{
			name: "chunk text by markdown",
			input: ChunkTextInput{
				Text: "Hello world.",
				Strategy: Strategy{
					Setting: Setting{
						ChunkMethod: "Markdown",
						ChunkSize:   5,
					},
				},
			},
			output: ChunkTextOutput{
				TextChunks: []TextChunk{
					{
						Text:          "Hello world.",
						StartPosition: 1,
						EndPosition:   12,
					},
				},
				ChunkNum: 1,
			},
		},
		{
			name: "chunk text by recursive",
			input: ChunkTextInput{
				Text: "Hello world.",
				Strategy: Strategy{
					Setting: Setting{
						ChunkMethod: "Recursive",
						ChunkSize:   5,
						Separators:  []string{" ", "."},
					},
				},
			},
			output: ChunkTextOutput{
				TextChunks: []TextChunk{
					{
						Text:          "Hello",
						StartPosition: 1,
						EndPosition:   5,
					},
					{
						Text:          "world",
						StartPosition: 6,
						EndPosition:   10,
					},
				},
				ChunkNum: 2,
			},
		},
	}

	for _, tc := range testCases {
		c.Run(tc.name, func(c *quicktest.C) {
			output, err := chunkText(tc.input)
			c.Assert(err, quicktest.IsNil)
			c.Check(output, quicktest.DeepEquals, tc.output)

		})
	}
}
