package document

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
)

func TestConvertDocumentToMarkdown(t *testing.T) {
	c := quicktest.New(t)

	tests := []struct {
		name     string
		filepath string
	}{
		{
			name:     "Convert PDF file",
			filepath: "testdata/test.pdf",
		},
		{
			name:     "Convert DOCX file",
			filepath: "testdata/test.docx",
		},
		{
			name:     "Convert HTML file",
			filepath: "testdata/test.html",
		},
		{
			name:     "Convert PPTX file",
			filepath: "testdata/test.pptx",
		},
	}
	for _, test := range tests {
		c.Run(test.name, func(c *quicktest.C) {
			fileContent, err := os.ReadFile(test.filepath)
			c.Assert(err, quicktest.IsNil)

			base64DataURI := fmt.Sprintf("data:%s;base64,%s", mimeTypeByExtension(test.filepath), base64.StdEncoding.EncodeToString(fileContent))

			inputStruct := ConvertDocumentToMarkdownInput{
				Document:        base64DataURI,
				DisplayImageTag: false,
			}
			input, err := base.ConvertToStructpb(inputStruct)
			c.Assert(err, quicktest.IsNil)
			e := &execution{
				getMarkdownTransformer: fakeGetMarkdownTransformer,
			}
			e.Task = "TASK_CONVERT_TO_MARKDOWN"

			output, err := e.convertDocumentToMarkdown(input)
			c.Assert(err, quicktest.IsNil)

			outputStruct := ConvertDocumentToMarkdownOutput{}
			err = base.ConvertFromStructpb(output, &outputStruct)
			c.Assert(err, quicktest.IsNil)
			c.Assert(outputStruct.Body, quicktest.DeepEquals, "This is test file")

		})
	}

}

func mimeTypeByExtension(filepath string) string {
	switch filepath {
	case "testdata/test.pdf":
		return "application/pdf"
	case "testdata/test.docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "testdata/test.html":
		return "text/html"
	case "testdata/test.pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	default:
		return ""
	}
}

func fakeGetMarkdownTransformer(fileExtension string, inputStruct *ConvertDocumentToMarkdownInput) (MarkdownTransformer, error) {
	return FakeMarkdownTransformer{}, nil
}

type FakeMarkdownTransformer struct {
	Base64EncodedText string
	FileExtension     string
	DisplayImageTag   bool
	Converter         string
}

func (f FakeMarkdownTransformer) Transform() (string, error) {
	return "This is test file", nil
}
