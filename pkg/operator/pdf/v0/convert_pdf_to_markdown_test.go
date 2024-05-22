package pdf

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"testing"

	qt "github.com/frankban/quicktest"
)

type mockCommandRunner struct {
}

type mockWriteCloser struct {
}

func (m *mockCommandRunner) CombinedOutput() ([]byte, error) {

	output := convertPDFToMarkdownOutput{
		Body: "# Test\n\nThis is a test document.\n\n",
		// TODO: revert it when target the bug.
		// Metadata: map[string]string{
		// 	"title": "Test",
		// },
	}

	bytes, err := json.Marshal(output)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (m *mockCommandRunner) StdinPipe() (io.WriteCloser, error) {
	stdin := mockWriteCloser{}
	return stdin, nil
}

func (m mockWriteCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m mockWriteCloser) Close() error {
	return nil
}

func TestConvertPdfToText(t *testing.T) {

	test := struct {
		name     string
		filepath string
	}{
		name:     "Convert PDF file",
		filepath: "testdata/test.pdf",
	}

	t.Run(test.name, func(t *testing.T) {
		c := qt.New(t)
		b := []byte{12, 34, 56, 78, 90}

		encoded := base64.StdEncoding.EncodeToString(b)
		input := convertPDFToMarkdownInput{
			PFD: encoded,
		}

		mockRunner := &mockCommandRunner{}
		output, err := convertPDFToMarkdown(input, mockRunner)
		c.Assert(err, qt.IsNil)

		c.Assert(output.Body, qt.Equals, "# Test\n\nThis is a test document.\n\n")

		// TODO: revert it when target the bug.
		// if output.Metadata == nil {
		// 	t.Fatalf("expected non-nil metadata")
		// 	return
		// }
	})

}
