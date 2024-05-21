package pdf

import "io"

type CommandRunner interface {
	CombinedOutput() ([]byte, error)
	StdinPipe() (io.WriteCloser, error)
}
type PdfTransformerOutput struct {
	Body     string            `json:"body"`
	Metadata map[string]string `json:"metadata"`
}

type ConvertPdfToMarkdownInput struct {
	// Pdf: PDF document to convert
	Doc string `json:"doc"`
}

type ConvertPdfToMarkdownOutput struct {
	// Markdown: Markdown content converted from the PDF document
	Body string `json:"body"`
	// Meta: Metadata extracted from the PDF document
	Metadata map[string]string `json:"metadata"`
}
