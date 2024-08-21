package text

import (
	"reflect"
	"strings"
)

type ChunkPositionCalculator interface {
	getChunkPositions(rawText, chunk []rune, startScanPosition int) (startPosition int, endPosition int)
}
type PositionCalculator struct{}
type MarkdownPositionCalculator struct{}

func (output *ChunkTextOutput) setChunksWithPosition(chunks []string, rawText, chunkMethod string) {

	rawRunes := []rune(rawText)
	var positionCalculator ChunkPositionCalculator

	switch chunkMethod {
	case "Token", "Recursive":
		positionCalculator = PositionCalculator{}
	case "Markdown":
		positionCalculator = MarkdownPositionCalculator{}
	}

	startScanPosition := 0
	for i, chunk := range chunks {
		chunkRunes := []rune(chunk)

		startPosition, endPosition := positionCalculator.getChunkPositions(rawRunes, chunkRunes, startScanPosition)

		if shouldScanRawTextFromPreviousChunk(startPosition, endPosition) {
			previousChunk := output.TextChunks[i-1]
			startPosition, endPosition = positionCalculator.getChunkPositions(rawRunes, chunkRunes, previousChunk.StartPosition)
		}

		if startPosition == endPosition {
			continue
		}
		output.TextChunks = append(output.TextChunks, TextChunk{
			Text:          chunk,
			StartPosition: startPosition,
			EndPosition:   endPosition,
		})
		startScanPosition = startPosition + 1
	}

	if len(output.TextChunks) == 0 {
		output.TextChunks = append(output.TextChunks, TextChunk{
			Text:          rawText,
			StartPosition: 0,
			EndPosition:   len(rawRunes) - 1,
		})
	}
}

func (PositionCalculator) getChunkPositions(rawText, chunk []rune, startScanPosition int) (startPosition int, endPosition int) {

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

	// In some models, the chunks are transformed into lowercase before scanning.
	// This is to handle the case where the chunk is not found in the raw text.
	if startPosition == 0 && endPosition == 0 {
		lowerString := strings.ToLower(string(rawText))
		checkerString := strings.ReplaceAll(lowerString, "\n", "")
		checker := []rune(checkerString)
		for i := startScanPosition; i < len(checker); i++ {
			if checker[i] == chunk[0] {

				if i+len(chunk) > len(checker) {
					break
				}

				if reflect.DeepEqual(checker[i:i+len(chunk)], chunk) {
					startPosition = i
					endPosition = len(chunk) + i - 1
					break
				}
			}
		}
	}

	return startPosition, endPosition
}

func (MarkdownPositionCalculator) getChunkPositions(rawText, chunk []rune, startScanPosition int) (startPosition int, endPosition int) {

	skipHeaderIndex := getSkipHeaderIndex(chunk)

	for i := startScanPosition; i < len(rawText); i++ {

		if rawText[i] == chunk[skipHeaderIndex] {

			if i+len(chunk)-skipHeaderIndex > len(rawText) {
				break
			}

			if reflect.DeepEqual(rawText[i:(i+len(chunk)-skipHeaderIndex)], chunk[skipHeaderIndex:]) {
				startPosition = i
				endPosition = len(chunk) + i - 1 - skipHeaderIndex
				break
			}
		}
	}
	return startPosition, endPosition
}

func shouldScanRawTextFromPreviousChunk(startPosition, endPosition int) bool {
	return startPosition == 0 && endPosition == 0
}

func getSkipHeaderIndex(chunk []rune) int {
	hashtagCount := 0
	skipPosition := 0
	for i := 0; i < len(chunk); i++ {
		if chunk[i] == '#' {
			hashtagCount++
		}

		if hashtagCount >= 1 && chunk[i] == '\n' {
			skipPosition = i + 1
			hashtagCount = 0
		}
	}
	return skipPosition
}
