package text

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/tmc/langchaingo/textsplitter"
)

type MarkdownTextSplitter struct {
	ChunkSize    int
	ChunkOverlap int
	RawText      string
}

func NewMarkdownTextSplitter(chunkSize, chunkOverlap int, rawText string) *MarkdownTextSplitter {
	return &MarkdownTextSplitter{
		ChunkSize:    chunkSize,
		ChunkOverlap: chunkOverlap,
		RawText:      rawText,
	}
}

func (sp *MarkdownTextSplitter) Validate() error {
	if sp.ChunkSize <= 0 {
		return fmt.Errorf("ChunkSize must be greater than 0")
	}
	if sp.ChunkOverlap < 0 {
		return fmt.Errorf("ChunkOverlap must be greater than or equal to 0")
	}
	if sp.ChunkOverlap >= sp.ChunkSize {
		return fmt.Errorf("ChunkOverlap must be less than ChunkSize")
	}
	return nil
}

func (sp *MarkdownTextSplitter) SplitText() ([]ContentChunk, error) {
	var chunks []ContentChunk

	rawRunes := []rune(sp.RawText)

	docs, err := buildDocuments(rawRunes)

	if err != nil {
		return chunks, fmt.Errorf("failed to build documents: %w", err)
	}
	chunkMap := make(map[string]bool)

	for _, doc := range docs {
		for _, content := range doc.Contents {
			var newChunks []ContentChunk
			switch content.Type {
			case "table":
				newChunks, err = sp.chunkTable(content, doc.Headers)
			case "list":
				newChunks, err = sp.chunkList(content, doc.Headers)
			case "plaintext":
				newChunks, err = sp.chunkPlainText(content, doc.Headers)
			}
			if err != nil {
				return chunks, fmt.Errorf("failed to chunk content: %w", err)
			}
			appendUniqueChunksMap(&chunks, newChunks, &chunkMap)
		}
	}

	return chunks, nil
}

func appendUniqueChunksMap(chunks *[]ContentChunk, newChunks []ContentChunk, chunkMap *map[string]bool) {
	for _, newChunk := range newChunks {
		key := fmt.Sprintf("%d-%d", newChunk.ContentStartPosition, newChunk.ContentEndPosition)
		if _, exists := (*chunkMap)[key]; !exists {
			*chunks = append(*chunks, newChunk)
			(*chunkMap)[key] = true
		}
	}
}

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

	startPosition := content.BlockStartPosition

	tableHeader := content.Table.HeaderText
	if len(tableHeader) > 0 {
		headerString += content.Table.HeaderText + "\n"
		startPosition += sizeOfString(content.Table.HeaderText) + 1
	}

	headerRow := content.Table.HeaderRow
	if len(headerRow) > 0 {
		headerString += headerRow + "\n"
		startPosition += sizeOfString(headerRow) + 1
	}

	tableSeparator := content.Table.TableSeparator
	if len(tableSeparator) > 0 {
		headerString += content.Table.TableSeparator + "\n"
		startPosition += sizeOfString(content.Table.TableSeparator) + 1
	}

	var endPosition int
	for i := 0; i < len(rows); i++ {
		chunk := headerString

		if i > 0 && sizeOfString(rows[i-1]) < chunkOverlap {
			chunk += rows[i-1] + "\n"
			startPosition -= sizeOfString(rows[i-1]) + 1
			endPosition = startPosition + sizeOfString(rows[i-1]) + 1
		} else {
			endPosition = startPosition
		}

		chunk += rows[i] + "\n"
		endPosition += sizeOfString(rows[i]) - 1

		for j := i + 1; j < len(rows) && sizeOfString(chunk+rows[j]) < chunkSize; j++ {
			chunk += rows[j] + "\n"
			endPosition += sizeOfString(rows[j]) + 1
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

func (sp MarkdownTextSplitter) chunkList(content Content, _ []Header) ([]ContentChunk, error) {
	var chunks []ContentChunk

	lists := content.Lists

	chunks = sp.processChunks(lists)

	return chunks, nil
}

func (sp MarkdownTextSplitter) processChunks(lists []List) []ContentChunk {
	contentChunks := []ContentChunk{}
	currentChunk := ""
	currentChunkSize := 0
	currentStartPosition := 0
	currentEndPosition := 0
	isPrepended := false

	addListCount := 0
	for i := 0; i < len(lists); i++ {
		list := lists[i]

		// Add the title
		if addListCount == 1 && sizeOfString(currentChunk)+sizeOfString(list.HeaderText) < sp.ChunkSize {
			currentChunk = list.HeaderText + "\n" + currentChunk
			currentChunkSize += sizeOfString(list.Text) + 1
		}

		if sizeOfString(list.Text) > sp.ChunkSize {

			if len(currentChunk) > 0 {
				previousChunk := ContentChunk{
					Chunk:                currentChunk,
					ContentStartPosition: currentStartPosition,
					ContentEndPosition:   currentEndPosition,
				}
				currentChunk = ""
				currentChunkSize = 0
				contentChunks = append(contentChunks, previousChunk)
				isPrepended = false
			}

			prependList := &list
			var prependString string

			for prependList.PreviousLevelList != nil {
				prependList = prependList.PreviousLevelList

				if len(prependList.Text) > 0 &&
					sizeOfString(prependList.Text) <= sp.ChunkSize { // Do not prepend if the list is too large
					prependString = prependList.Text + "\n" + prependString
				}
			}

			prependString = list.HeaderText + "\n" + prependString

			smallerChunks := sp.chunkLargeList(list, sizeOfString(prependString))

			if len(prependString) > 0 {
				for i := range smallerChunks {
					smallerChunks[i].Chunk = prependString + smallerChunks[i].Chunk
				}
			}

			contentChunks = append(contentChunks, smallerChunks...)
			addListCount = 0

		} else {
			if !isPrepended {
				prependList := &list
				var prependString string

				for prependList.PreviousLevelList != nil {

					prependList = prependList.PreviousLevelList
					if len(prependList.Text) > 0 &&
						sizeOfString(prependList.Text) <= sp.ChunkSize { // Do not prepend if the list is too large
						prependString = prependList.Text + "\n" + prependString
					}
				}
				isPrepended = true
				currentChunk += prependString + list.Text + "\n"
				currentChunkSize += sizeOfString(list.Text)
				currentStartPosition = list.StartPosition
				currentEndPosition = list.EndPosition
				addListCount++
			} else if currentChunkSize+sizeOfString(list.Text) < sp.ChunkSize {
				currentChunk += list.Text + "\n"
				currentChunkSize += sizeOfString(list.Text) + 1
				currentEndPosition = list.EndPosition
				addListCount++

			} else {
				contentChunks = append(contentChunks, ContentChunk{
					Chunk:                currentChunk,
					ContentStartPosition: currentStartPosition,
					ContentEndPosition:   currentEndPosition,
				})
				isPrepended = false
				currentChunk = ""
				currentChunkSize = 0
				currentStartPosition = 0 // To be set in !isPrepended Block
				currentEndPosition = 0   // To be set in isPrepended Block

				// Overlap Handling: Start from final appended list if it's smaller than the overlap
				if i > 0 && sizeOfString(lists[i-1].Text) <= sp.ChunkOverlap {
					i -= 2
					addListCount = -1
				} else {
					addListCount = 0
				}
			}
		}
	}

	if currentChunkSize > 0 {
		contentChunks = append(contentChunks, ContentChunk{
			Chunk:                currentChunk,
			ContentStartPosition: currentStartPosition,
			ContentEndPosition:   currentEndPosition,
		})
	}

	return contentChunks
}

// chunkLargeList splits a large list item into multiple chunks by words
func (sp MarkdownTextSplitter) chunkLargeList(list List, prependStringSize int) []ContentChunk {
	var chunks []ContentChunk
	words := strings.Fields(list.Text)
	currentChunk := ""
	currentChunkSize := 0
	currentStartPosition := list.StartPosition
	currentEndPosition := list.StartPosition + list.indentation

	chunkSizeToUse := sp.ChunkSize - prependStringSize
	if chunkSizeToUse <= 0 || sp.ChunkOverlap >= chunkSizeToUse { // avoid edge case where chunkSize is too small
		chunkSizeToUse = sp.ChunkSize
	}

	for i := 0; i < len(words); {
		wordSize := sizeOfString(words[i])
		if currentChunkSize+wordSize <= chunkSizeToUse {
			currentChunk += words[i] + " "
			currentChunkSize += wordSize
			currentEndPosition += wordSize + 1
			i++
		} else {
			chunks = append(chunks, ContentChunk{
				Chunk:                currentChunk,
				ContentStartPosition: currentStartPosition,
				ContentEndPosition:   currentEndPosition - 2,
			})
			overlapSize := sp.ChunkOverlap
			for overlapSize-sizeOfString(words[i])+1 >= 0 {
				i--
				if i == 0 {
					break
				}
				overlapSize -= sizeOfString(words[i]) + 1
				currentEndPosition -= sizeOfString(words[i]) + 1
			}
			currentStartPosition = currentEndPosition
			currentChunk = ""
			currentChunkSize = 0
		}
	}

	// Add the last chunk
	if currentChunkSize > 0 {
		chunks = append(chunks, ContentChunk{
			Chunk:                currentChunk,
			ContentStartPosition: currentStartPosition,
			ContentEndPosition:   currentEndPosition - 2,
		})
	}
	return chunks
}

func (sp MarkdownTextSplitter) chunkPlainText(content Content, headers []Header) ([]ContentChunk, error) {

	split := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(sp.ChunkSize),
		textsplitter.WithChunkOverlap(sp.ChunkOverlap),
	)

	chunks, err := split.SplitText(content.PlainText)

	if err != nil {
		return nil, err
	}

	prependHeader := ""
	for _, header := range headers {
		trimmedHeader := strings.TrimSpace(header.Text)
		if len(trimmedHeader) == 0 {
			continue
		}
		prependHeader += header.Text + "\n"
	}

	rawRunes := []rune(sp.RawText)
	startScanPosition := 0

	contentChunks := []ContentChunk{}
	for _, chunk := range chunks {
		chunkRunes := []rune(chunk)

		startPosition, endPosition := getChunkPositions(rawRunes, chunkRunes, startScanPosition)

		if shouldScanRawTextFromPreviousChunk(startPosition, endPosition) {
			previousChunkIndex := len(contentChunks) - 1
			previousChunk := contentChunks[previousChunkIndex]
			startPosition, endPosition = getChunkPositions(rawRunes, chunkRunes, previousChunk.ContentStartPosition+1)
		}

		if startPosition == endPosition {
			continue
		}

		contentChunks = append(contentChunks, ContentChunk{
			Chunk:                prependHeader + "\n" + chunk,
			ContentStartPosition: startPosition,
			ContentEndPosition:   endPosition,
		})
		startScanPosition = startPosition + 1
	}

	return contentChunks, nil
}

func getChunkPositions(rawText, chunk []rune, startScanPosition int) (startPosition int, endPosition int) {

	for i := startScanPosition; i < len(rawText); i++ {
		if rawText[i] == chunk[0] {

			if i+len(chunk) > len(rawText) {
				break
			}

			if reflect.DeepEqual(rawText[i:i+len(chunk)], chunk) {
				startPosition = i
				endPosition = len(chunk) + i - 1
				break
			}
		}
	}
	return startPosition, endPosition
}
