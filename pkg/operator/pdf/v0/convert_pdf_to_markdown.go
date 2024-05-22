package pdf

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
	PFD string `json:"pdf"`
}

type convertPDFToMarkdownOutput struct {
	// Markdown: Markdown content converted from the PDF document
	Body string `json:"body"`
	// Metadata: Metadata extracted from the PDF document

	// https://linear.app/instill-ai/issue/INS-3098/[component][pdf-operator]-add-task-convert-to-markdown#comment-dc17f0f3
	// TODO: revert it when target the bug.
	// Metadata map[string]string `json:"metadata"`
}

func convertPDFToMarkdown(input convertPDFToMarkdownInput, cmdRunner commandRunner) (convertPDFToMarkdownOutput, error) {

	b, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(input.PFD))
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
