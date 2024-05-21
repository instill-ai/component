package pdf

import (
	"encoding/json"
	"io"
)

type MockCommandRunner struct {
}

type MockWriteCloser struct {
}

func (m *MockCommandRunner) CombinedOutput() ([]byte, error) {

	output := ConvertPdfToMarkdownOutput{
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

func (m *MockCommandRunner) StdinPipe() (io.WriteCloser, error) {
	stdin := MockWriteCloser{}
	return stdin, nil
}

func (m MockWriteCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m MockWriteCloser) Close() error {
	return nil
}
