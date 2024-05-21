package pdf

import (
	"encoding/base64"
	"os"
	"testing"
)

func TestConvertPdfToText(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
	}{
		{
			name:     "Convert pdf file",
			filepath: "testdata/test.pdf",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b, err := os.ReadFile(test.filepath)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}

			encoded := base64.StdEncoding.EncodeToString(b)
			input := ConvertPdfToMarkdownInput{
				Doc: encoded,
			}

			// mock python code
			mockRunner := &MockCommandRunner{}
			output, err := convertPdfToMarkdown(input, mockRunner)
			if err != nil {
				t.Fatalf("failed to convert pdf to text: %v", err)
				return
			}

			if output.Body == "" {
				t.Fatalf("expected non-empty body")
				return
			}

			if output.Metadata == nil {
				t.Fatalf("expected non-nil metadata")
				return
			}
		})
	}
}
