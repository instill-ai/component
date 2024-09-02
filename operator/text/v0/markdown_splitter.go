package text

import (
	"fmt"
)

func NewMarkdownTextSplitter(opts ...Option) MarkdownTextSplitter {
	options := DefaultOptions()

	for _, o := range opts {
		o(&options)
	}

	sp := MarkdownTextSplitter{
		ChunkSize:    options.ChunkSize,
		ChunkOverlap: options.ChunkOverlap,
	}

	return sp
}

type MarkdownTextSplitter struct {
	ChunkSize    int
	ChunkOverlap int
}

func (sp MarkdownTextSplitter) SplitText(text string) ([]ContentChunk, error) {

	rawTextRunes := []rune(text)
	docs, err := buildDocuments(rawTextRunes)

	if err != nil {
		return nil, fmt.Errorf("error building documents: %w", err)
	}

	finalContentChunks := []ContentChunk{}

	for _, doc := range docs {
		contentChunks, err := sp.splitDocument(doc)

		if err != nil {
			return nil, fmt.Errorf("error splitting document: %w", err)
		}

		finalContentChunks = append(finalContentChunks, contentChunks...)
	}

	return finalContentChunks, nil
}

func (sp MarkdownTextSplitter) splitDocument(doc MarkdownDocument) ([]ContentChunk, error) {

	contentChunks := []ContentChunk{}

	for _, content := range doc.Contents {

		var chunks []ContentChunk
		var err error

		if content.Type == "table" {
			chunks, err = sp.chunkTable(content, doc.Headers)
		} else if content.Type == "list" {
			chunks, err = sp.chunkList(content, doc.Headers)
		} else {
			chunks, err = sp.chunkPlainText(content, doc.Headers)
		}

		if err != nil {
			return nil, fmt.Errorf("error chunking content: %w", err)
		}

		contentChunks = append(contentChunks, chunks...)

	}

	return contentChunks, nil
}
