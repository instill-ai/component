package text

import (
	"fmt"
)

func markdownChunkText(input ChunkTextInput) (ChunkTextOutput, error) {

	setting := input.Strategy.Setting
	setting.SetDefault()

	splitter := NewMarkdownTextSplitter(
		WithChunkSize(setting.ChunkSize),
		WithChunkOverlap(setting.ChunkOverlap),
	)

	contentChunks, err := splitter.SplitText(input.Text)
	if err != nil {
		return ChunkTextOutput{}, fmt.Errorf("error splitting text: %w", err)
	}

	output := ChunkTextOutput{}
	for _, contentChunk := range contentChunks {
		output.ChunkNum += 1
		output.TextChunks = append(output.TextChunks, TextChunk{
			Text:          contentChunk.Chunk,
			StartPosition: contentChunk.ContentStartPosition,
			EndPosition:   contentChunk.ContentEndPosition,
		})
	}

	return output, nil
}
