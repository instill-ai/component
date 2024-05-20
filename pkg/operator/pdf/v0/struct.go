package pdf

type ConvertPdfToMarkdownInput struct {
	// Pdf: PDF document to convert
	Pdf string `json:"pdf"`
}

type ConvertPdfToMarkdownOutput struct {
	// Markdown: Markdown content converted from the PDF document
	Markdown string `json:"markdown"`
	// Meta: Metadata extracted from the PDF document
	Meta map[string]string `json:"meta"`
	// Error: Error message if any during the conversion process
	Error string `json:"error"`
}