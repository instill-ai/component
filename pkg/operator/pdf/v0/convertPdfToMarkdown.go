package pdf

import (
	"encoding/base64"
	"encoding/json"

	"github.com/instill-ai/component/pkg/base"
)

func convertPdfToMarkdown(input ConvertPdfToMarkdownInput, cmdRunner CommandRunner) (ConvertPdfToMarkdownOutput, error) {

	b, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(input.Doc))
	if err != nil {
		return ConvertPdfToMarkdownOutput{}, err
	}

	stdin, err := cmdRunner.StdinPipe()
	if err != nil {
		return ConvertPdfToMarkdownOutput{}, err
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
		return ConvertPdfToMarkdownOutput{}, err
	}

	writeErr := <-errChan
	if writeErr != nil {
		return ConvertPdfToMarkdownOutput{}, writeErr
	}

	var output ConvertPdfToMarkdownOutput
	err = json.Unmarshal(outputBytes, &output)
	if err != nil {
		return ConvertPdfToMarkdownOutput{}, err
	}

	return output, nil
}
