package document

import (
	"encoding/base64"
	"encoding/json"
	"io"

	"github.com/instill-ai/component/pkg/base"
)

type commandRunner interface {
	CombinedOutput() ([]byte, error)
	StdinPipe() (io.WriteCloser, error)
}

type convertPDFToMarkdownInput struct {
	PDF string `json:"pdf"`
}

type convertPDFToMarkdownOutput struct {
	Body string `json:"body"`
}

func convertPDFToMarkdown(input convertPDFToMarkdownInput, cmdRunner commandRunner) (convertPDFToMarkdownOutput, error) {

	b, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(input.PDF))
	if err != nil {
		return convertPDFToMarkdownOutput{}, err
	}

	stdin, err := cmdRunner.StdinPipe()
	if err != nil {
		return convertPDFToMarkdownOutput{}, err
	}

	errChan := make(chan error, 1)

	go func() {
		defer stdin.Close()
		_, err := stdin.Write(b)
		if err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	outputBytes, err := cmdRunner.CombinedOutput()
	if err != nil {
		return convertPDFToMarkdownOutput{}, err
	}

	writeErr := <-errChan
	if writeErr != nil {
		return convertPDFToMarkdownOutput{}, writeErr
	}

	var output convertPDFToMarkdownOutput
	err = json.Unmarshal(outputBytes, &output)
	if err != nil {
		return convertPDFToMarkdownOutput{}, err
	}

	return output, nil
}
