package text

import (
	"github.com/tmc/langchaingo/textsplitter"
)

func NewMarkdownTextSplitter(opts ...textsplitter.Option) MarkdownTextSplitter {
	options := textsplitter.DefaultOptions()

	for _, o := range opts {
		o(&options)
	}

	sp := MarkdownTextSplitter{
		ChunkSize:    options.ChunkSize,
		ChunkOverlap: options.ChunkOverlap,
		ContentSplitter: textsplitter.NewRecursiveCharacter(
			textsplitter.WithChunkSize(options.ChunkSize),
			textsplitter.WithChunkOverlap(options.ChunkOverlap),
			textsplitter.WithSeparators([]string{
				"\n\n",
				"\n",
				" ",
			})),
	}

	return sp
}

type MarkdownTextSplitter struct {
	ChunkSize       int
	ChunkOverlap    int
	ContentSplitter textsplitter.RecursiveCharacter
}

func (sp MarkdownTextSplitter) SplitText(text string) ([]string, error) {

	rawTextRunes := []rune(text)
	documents := sp.buildDocuments(rawTextRunes)

	finalChunks := []string{}

	for _, d := range documents {
		chunks, err := sp.splitDocument(d)

		if err != nil {
			return nil, err
		}

		finalChunks = append(finalChunks, chunks...)
	}

	return finalChunks, nil
}

type MarkdownDocument struct {
	Header1            string
	Header2            string
	Header3            string
	Header4            string
	Header5            string
	Header6            string
	SumHeaderChunkSize int
	Content            string
}

func (sp MarkdownTextSplitter) buildDocuments(rawRunes []rune) []MarkdownDocument {

	documents := []MarkdownDocument{}

	startPosition := 0
	document, startPosition := sp.buildDocument(rawRunes, MarkdownDocument{}, startPosition)

	documents = append(documents, document)

	for startPosition < len(rawRunes) {
		document, startPosition = sp.buildDocument(rawRunes, document, startPosition)

		if startPosition == 0 {
			break
		}
		documents = append(documents, document)
	}

	return documents
}

// Definition:
// Document means the struct with Header1, Header2, … Header6, and Content.
// The Content must not be blank.
// The Header can be blank.
// A Document can only have a header for each layer
func (sp MarkdownTextSplitter) buildDocument(rawRunes []rune, previousDocument MarkdownDocument, startPosition int) (doc MarkdownDocument, endPosition int) {

	if documentStartsWithoutHeader(rawRunes, startPosition, previousDocument) {
		document := MarkdownDocument{}
		for startPosition < len(rawRunes) {
			if string(rawRunes[startPosition]) == "#" && !isHashtagInContent(startPosition, rawRunes) {
				break
			}
			startPosition++
		}
		noHeaderContent := string(rawRunes[:startPosition])
		document.Content = noHeaderContent
		setHeaderSize(&document)
		return document, startPosition - 1
	}

	l := startPosition
	r := startPosition
	hashtagCount := 0
	layer1Count := 0
	layer2Count := 0
	layer3Count := 0
	layer4Count := 0
	layer5Count := 0
	layer6Count := 0

	document := previousDocument
	document.Content = ""
	var content string
	for r < len(rawRunes) {
		if string(rawRunes[r]) == "#" {
			if !isHashtagInContent(r, rawRunes) {
				isEnd, layer := endLayerOfDocument(content, layer1Count, layer2Count, layer3Count, layer4Count, layer5Count, layer6Count)
				if isEnd {
					document.Content = content
					clearLowerLayerHeader(&document, layer)
					setHeaderSize(&document)
					return document, r - 1
				}
				hashtagCount++
				r++
				continue
			}
		}
		if hashtagCount > 0 {
			if string(rawRunes[r]) == "\n" {
				if hashtagCount == 1 {
					document.Header1 = string(rawRunes[l : r+1])
					layer1Count++
				}
				if hashtagCount == 2 {
					document.Header2 = string(rawRunes[l : r+1])
					layer2Count++
				}
				if hashtagCount == 3 {
					document.Header3 = string(rawRunes[l : r+1])
					layer3Count++
				}
				if hashtagCount == 4 {
					document.Header4 = string(rawRunes[l : r+1])
					layer4Count++
				}
				if hashtagCount == 5 {
					document.Header5 = string(rawRunes[l : r+1])
					layer5Count++
				}
				if hashtagCount == 6 {
					document.Header6 = string(rawRunes[l : r+1])
					layer6Count++
				}
				hashtagCount = 0
				l = r
			}
			r++
			continue
		}

		if isContent(hashtagCount, layer1Count, layer2Count, layer3Count, layer4Count, layer5Count, layer6Count, content, string(rawRunes[r])) {
			content += string(rawRunes[r])
			r++
		} else {
			r++
		}

	}
	return document, 0
}

func endLayerOfDocument(content string, layer1Count, layer2Count, layer3Count, layer4Count, layer5Count, layer6Count int) (bool, int) {
	if content == "" {
		return false, 0
	}
	if layer6Count >= 1 {
		return true, 6
	}
	if layer5Count >= 1 {
		return true, 5
	}
	if layer4Count >= 1 {
		return true, 4
	}
	if layer3Count >= 1 {
		return true, 3
	}
	if layer2Count >= 1 {
		return true, 2
	}
	if layer1Count >= 1 {
		return true, 1
	}
	return true, 0
}

func clearLowerLayerHeader(doc *MarkdownDocument, layer int) {
	if layer == 0 {
		doc.Header1 = ""
		doc.Header2 = ""
		doc.Header3 = ""
		doc.Header4 = ""
		doc.Header5 = ""
		doc.Header6 = ""
	}
	if layer == 1 {
		doc.Header2 = ""
		doc.Header3 = ""
		doc.Header4 = ""
		doc.Header5 = ""
		doc.Header6 = ""
	}
	if layer == 2 {
		doc.Header3 = ""
		doc.Header4 = ""
		doc.Header5 = ""
		doc.Header6 = ""
	}
	if layer == 3 {
		doc.Header4 = ""
		doc.Header5 = ""
		doc.Header6 = ""
	}
	if layer == 4 {
		doc.Header5 = ""
		doc.Header6 = ""
	}
	if layer == 5 {
		doc.Header6 = ""
	}
}

func setHeaderSize(doc *MarkdownDocument) {
	doc.SumHeaderChunkSize = sizeOfString(doc.Header1) + sizeOfString(doc.Header2) + sizeOfString(doc.Header3) + sizeOfString(doc.Header4) + sizeOfString(doc.Header5) + sizeOfString(doc.Header6)
}

func isContent(hashtagCount, layer1Count, layer2Count, layer3Count, layer4Count, layer5Count, layer6Count int, content, chr string) bool {
	if isSeparator(chr) && len(content) == 0 {
		return false
	}

	return hashtagCount == 0 && (layer1Count >= 1 ||
		layer2Count >= 1 ||
		layer3Count >= 1 ||
		layer4Count >= 1 ||
		layer5Count >= 1 ||
		layer6Count >= 1)
}

func isSeparator(text string) bool {
	separators := []string{"\n\n", "\n", " ", ""}
	for _, sep := range separators {
		if sep == text {
			return true
		}
	}
	return false
}

func isHashtagInContent(position int, rawRunes []rune) bool {
	if position == 0 {
		return false
	}
	hashTagCount := 1
	breakChar := string(rawRunes[position-1])
	if string(rawRunes[position-1]) == "#" {
		hashTagCount++
		for i := position - 2; i >= 0; i-- {
			if string(rawRunes[i]) == "#" {
				hashTagCount++
			} else {
				breakChar = string(rawRunes[i])
				break
			}
		}
	}
	if hashTagCount > 6 {
		return true
	}
	if breakChar == "\n" {
		return false
	}
	return true
}

func documentStartsWithoutHeader(rawRunes []rune, startPosition int, previousDocument MarkdownDocument) bool {
	if isBlankDocument(previousDocument) && startPosition == 0 {
		for startPosition < len(rawRunes) {
			if isSeparator(string(rawRunes[startPosition])) {
				startPosition++
				continue
			}
			if string(rawRunes[startPosition]) == "#" {
				return false
			}
			return true
		}
	}
	return false
}

func isBlankDocument(document MarkdownDocument) bool {
	return document.Header1 == "" && document.Header2 == "" && document.Header3 == "" && document.Header4 == "" && document.Header5 == "" && document.Header6 == "" && document.Content == ""
}

func (sp MarkdownTextSplitter) splitDocument(document MarkdownDocument) ([]string, error) {

	contentSplitter := sp.ContentSplitter
	if sp.ChunkSize < document.SumHeaderChunkSize {
		contentSplitter.ChunkSize = sp.ChunkSize
	} else {
		contentSplitter.ChunkSize = sp.ChunkSize - document.SumHeaderChunkSize
	}
	contentSplitter.ChunkOverlap = sp.ChunkOverlap

	chunks, err := contentSplitter.SplitText(document.Content)
	if err != nil {
		return nil, err
	}

	var documentFinalChunks []string
	for _, chunk := range chunks {

		prependHeaderChunk := sp.prependHeaderWithoutOverChunkSizeSetting(document, chunk)

		documentFinalChunks = append(documentFinalChunks, prependHeaderChunk)
	}

	return documentFinalChunks, nil
}

func (sp MarkdownTextSplitter) prependHeaderWithoutOverChunkSizeSetting(document MarkdownDocument, chunk string) string {

	if sizeOfString(chunk) >= sp.ChunkSize {
		return chunk
	}
	midChunk6 := document.Header6 + chunk

	if sizeOfString(midChunk6) >= sp.ChunkSize {
		return chunk
	}

	midChunk5 := document.Header5 + midChunk6

	if sizeOfString(midChunk5) >= sp.ChunkSize {
		return midChunk6
	}

	midChunk4 := document.Header4 + midChunk5

	if sizeOfString(midChunk4) >= sp.ChunkSize {
		return midChunk5
	}

	midChunk3 := document.Header3 + midChunk4

	if sizeOfString(midChunk3) >= sp.ChunkSize {
		return midChunk4
	}

	midChunk2 := document.Header2 + midChunk3

	if sizeOfString(midChunk2) >= sp.ChunkSize {
		return midChunk3
	}

	midChunk1 := document.Header1 + midChunk2

	if sizeOfString(midChunk1) >= sp.ChunkSize {
		return midChunk2
	}

	return midChunk1
}

func sizeOfString(text string) int {
	return len([]rune(text))
}
