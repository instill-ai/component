package text

import "strings"

type ContentChunk struct {
	Chunk                string
	ContentStartPosition int
	ContentEndPosition   int
}

func (sp MarkdownTextSplitter) chunkTable(content Content, headers []Header) ([]ContentChunk, error) {

	rows := content.Table.Rows
	chunks := []ContentChunk{}

	chunkSize := sp.ChunkSize
	chunkOverlap := sp.ChunkOverlap

	headerString := ""
	for _, header := range headers {
		trimmedHeader := strings.TrimSpace(header.Text)
		if len(trimmedHeader) == 0 {
			continue
		}
		headerString += header.Text + "\n"
	}

	// Block starts
	startPosition := content.BlockStartPosition

	tableHeader := content.Table.HeaderText
	if len(tableHeader) > 0 {
		headerString += content.Table.HeaderText + "\n"
		startPosition += len(content.Table.HeaderText) + 1
	}

	headerRow := content.Table.HeaderRow
	if len(headerRow) > 0 {
		headerString += headerRow + "\n"
		startPosition += len(headerRow) + 1
	}

	tableSeparator := content.Table.TableSeparator
	if len(tableSeparator) > 0 {
		headerString += content.Table.TableSeparator + "\n"
		startPosition += len(content.Table.TableSeparator) + 1
	}

	var endPosition int
	for i := 0; i < len(rows); i++ {
		chunk := headerString

		if i > 0 && len(rows[i-1]) < chunkOverlap {
			chunk += rows[i-1] + "\n"
			startPosition -= len(rows[i-1]) + 1
			endPosition = startPosition + len(rows[i-1]) + 1
		} else {
			endPosition = startPosition
		}

		chunk += rows[i] + "\n"
		endPosition += len(rows[i]) - 1

		for j := i + 1; j < len(rows) && len(chunk+rows[j]) < chunkSize; j++ {
			chunk += rows[j] + "\n"
			endPosition += len(rows[j]) + 1
			i = j
		}

		chunks = append(chunks, ContentChunk{
			Chunk:                chunk,
			ContentStartPosition: startPosition,
			ContentEndPosition:   endPosition,
		})

		startPosition = endPosition + 2 // new line and the first character of the next row

	}

	return chunks, nil
}

func (sp MarkdownTextSplitter) chunkList(content Content, headers []Header) ([]ContentChunk, error) {

	return []ContentChunk{}, nil
}

func (sp MarkdownTextSplitter) chunkPlainText(content Content, headers []Header) ([]ContentChunk, error) {
	return []ContentChunk{}, nil
}
