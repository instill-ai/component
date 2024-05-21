package pdf

import "io"

type CommandRunner interface {
	CombinedOutput() ([]byte, error)
	StdinPipe() (io.WriteCloser, error)
}

type ConvertPdfToMarkdownInput struct {
	// Pdf: PDF document to convert
	Doc string `json:"doc"`
}

type ConvertPdfToMarkdownOutput struct {
	// Markdown: Markdown content converted from the PDF document
	Body string `json:"body"`
	// Metadata: Metadata extracted from the PDF document
	
	// TODO: revert it when target the bug.
	// Metadata map[string]string `json:"metadata"`
}
