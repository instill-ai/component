package text

import (
	"context"
	"os"
	"testing"

	"github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
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
				Tokenization: Tokenization{
					Choice: Choice{
						TokenizationMethod: "Model",
						Model:              "gpt-3.5-turbo",
						Encoding:           "",
						HuggingFaceModel:   "",
					},
				},
			},
			output: ChunkTextOutput{
				TextChunks: []TextChunk{
					{
						Text:          "Hello world.",
						StartPosition: 0,
						EndPosition:   11,
						TokenCount:    3,
					},
				},
				ChunkNum:         1,
				TokenCount:       3,
				ChunksTokenCount: 3,
			},
		},
		{
			name: "chunk text by markdown",
			input: ChunkTextInput{
				Text: "Hello world.",
				Strategy: Strategy{
					Setting: Setting{
						ChunkMethod: "Markdown",
						ModelName:   "gpt-3.5-turbo",
						ChunkSize:   5,
					},
				},
				Tokenization: Tokenization{
					Choice: Choice{
						TokenizationMethod: "Model",
						Model:              "gpt-3.5-turbo",
						Encoding:           "",
						HuggingFaceModel:   "",
					},
				},
			},
			output: ChunkTextOutput{
				TextChunks: []TextChunk{
					{
						Text:          "Hello",
						StartPosition: 0,
						EndPosition:   4,
						TokenCount:    1,
					},
					{
						Text:          "world.",
						StartPosition: 6,
						EndPosition:   11,
						TokenCount:    2,
					},
				},
				ChunkNum:         2,
				TokenCount:       3,
				ChunksTokenCount: 3,
			},
		},
		{
			name: "chunk text by recursive",
			input: ChunkTextInput{
				Text: "Hello world.",
				Strategy: Strategy{
					Setting: Setting{
						ChunkMethod: "Recursive",
						ModelName:   "gpt-3.5-turbo",
						ChunkSize:   5,
						Separators:  []string{" ", "."},
					},
				},
				Tokenization: Tokenization{
					Choice: Choice{
						TokenizationMethod: "Model",
						Model:              "gpt-3.5-turbo",
						Encoding:           "",
						HuggingFaceModel:   "",
					},
				},
			},
			output: ChunkTextOutput{
				TextChunks: []TextChunk{
					{
						Text:          "Hello",
						StartPosition: 0,
						EndPosition:   4,
						TokenCount:    1,
					},
					{
						Text:          "world",
						StartPosition: 6,
						EndPosition:   10,
						TokenCount:    1,
					},
				},
				ChunkNum:         2,
				TokenCount:       3,
				ChunksTokenCount: 2,
			},
		},
	}

	for _, tc := range testCases {
		c.Run(tc.name, func(c *quicktest.C) {

			bc := base.Component{}
			component := Init(bc)
			c.Assert(component, quicktest.IsNotNil)

			execution, err := component.CreateExecution(base.ComponentExecution{
				Component: component,
				Task:      "TASK_CHUNK_TEXT",
			})

			c.Assert(err, quicktest.IsNil)
			c.Assert(execution, quicktest.IsNotNil)

			inputPd, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, quicktest.IsNil)

			outputPd, err := execution.Execute(context.TODO(), []*structpb.Struct{inputPd})
			c.Assert(err, quicktest.IsNil)
			output := ChunkTextOutput{}
			err = base.ConvertFromStructpb(outputPd[0], &output)

			c.Assert(err, quicktest.IsNil)
			c.Check(output, quicktest.DeepEquals, tc.output)
		})
	}
}

func Test_ChunkPositionCalculator(t *testing.T) {
	c := quicktest.New(t)

	testCases := []struct {
		name                   string
		positionCalculatorType string
		rawTextFilePath        string
		chunkTextFilePath      string
		expectStartPosition    int
		expectEndPosition      int
	}{
		{
			name:                   "Chinese text with NOT Markdown Chunking 1",
			positionCalculatorType: "PositionCalculator",
			rawTextFilePath:        "testdata/chinese/text1.txt",
			chunkTextFilePath:      "testdata/chinese/chunk1_1.txt",
			expectStartPosition:    0,
			expectEndPosition:      35,
		},
		{
			name:                   "Chinese text with NOT Markdown Chunking 2",
			positionCalculatorType: "PositionCalculator",
			rawTextFilePath:        "testdata/chinese/text1.txt",
			chunkTextFilePath:      "testdata/chinese/chunk1_2.txt",
			expectStartPosition:    26,
			expectEndPosition:      46,
		},
		{
			name:                   "Chinese text with NOT Markdown Chunking 3",
			positionCalculatorType: "PositionCalculator",
			rawTextFilePath:        "testdata/chinese/text1.txt",
			chunkTextFilePath:      "testdata/chinese/chunk1_3.txt",
			expectStartPosition:    49,
			expectEndPosition:      80,
		},
		{
			name:                   "Chinese text with Markdown Chunking 1",
			positionCalculatorType: "MarkdownPositionCalculator",
			rawTextFilePath:        "testdata/chinese_markdown/text1.txt",
			chunkTextFilePath:      "testdata/chinese_markdown/chunk1_1.txt",
			expectStartPosition:    4,
			expectEndPosition:      46,
		},
		{
			name:                   "Chinese text with Markdown Chunking 2",
			positionCalculatorType: "MarkdownPositionCalculator",
			rawTextFilePath:        "testdata/chinese_markdown/text1.txt",
			chunkTextFilePath:      "testdata/chinese_markdown/chunk1_2.txt",
			expectStartPosition:    49,
			expectEndPosition:      91,
		},
		{
			name:                   "Chinese text with Markdown Chunking 3",
			positionCalculatorType: "MarkdownPositionCalculator",
			rawTextFilePath:        "testdata/chinese_markdown/text1.txt",
			chunkTextFilePath:      "testdata/chinese_markdown/chunk1_3.txt",
			expectStartPosition:    98,
			expectEndPosition:      140,
		},
		{
			name:                   "English text with Markdown Chunking 1",
			positionCalculatorType: "MarkdownPositionCalculator",
			rawTextFilePath:        "testdata/english/text1.txt",
			chunkTextFilePath:      "testdata/english/chunk1_1.txt",
			expectStartPosition:    4,
			expectEndPosition:      25,
		},
		{
			name:                   "English text with Markdown Chunking 2",
			positionCalculatorType: "MarkdownPositionCalculator",
			rawTextFilePath:        "testdata/english/text1.txt",
			chunkTextFilePath:      "testdata/english/chunk1_2.txt",
			expectStartPosition:    16,
			expectEndPosition:      47,
		},
		{
			name:                   "English text with Markdown Chunking 3",
			positionCalculatorType: "MarkdownPositionCalculator",
			rawTextFilePath:        "testdata/english/text1.txt",
			chunkTextFilePath:      "testdata/english/chunk1_3.txt",
			expectStartPosition:    38,
			expectEndPosition:      58,
		},
	}

	for _, tc := range testCases {
		c.Run(tc.name, func(c *quicktest.C) {
			var calculator ChunkPositionCalculator
			if tc.positionCalculatorType == "PositionCalculator" {
				calculator = PositionCalculator{}
			} else if tc.positionCalculatorType == "MarkdownPositionCalculator" {
				calculator = MarkdownPositionCalculator{}
			}
			rawTextBytes, err := os.ReadFile(tc.rawTextFilePath)
			c.Assert(err, quicktest.IsNil)
			rawTextRunes := []rune(string(rawTextBytes))

			chunkText, err := os.ReadFile(tc.chunkTextFilePath)
			c.Assert(err, quicktest.IsNil)

			chunkTextRunes := []rune(string(chunkText))

			startPosition, endPosition := calculator.getChunkPositions(rawTextRunes, chunkTextRunes, 0)

			c.Assert(startPosition, quicktest.Equals, tc.expectStartPosition)
			c.Assert(endPosition, quicktest.Equals, tc.expectEndPosition)

		})
	}
}
