package pdf

import (
	"encoding/base64"
	"encoding/json"
	"os/exec"

	"github.com/instill-ai/component/pkg/base"
)

func convertPdfToMarkdown(input ConvertPdfToMarkdownInput) (ConvertPdfToMarkdownOutput, error) {

	b, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(input.Doc))
	if err != nil {
		return ConvertPdfToMarkdownOutput{}, err
	}

	scriptPath := "/component/pkg/operator/pdf/v0/python/pdfTransformer.py"
	pythonInterpreter := "/opt/venv/bin/python"

	cmd := exec.Command(pythonInterpreter, scriptPath)
	stdin, err := cmd.StdinPipe()
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

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return ConvertPdfToMarkdownOutput{}, err
	}

	writeErr := <-errChan
	if writeErr != nil {
		return ConvertPdfToMarkdownOutput{}, err
	}

	var output PdfTransformerOutput
	err = json.Unmarshal(outputBytes, &output)
	if err != nil {
		return ConvertPdfToMarkdownOutput{}, err
	}

	resp := ConvertPdfToMarkdownOutput{
		Body:     output.Body,
		Metadata: output.Metadata,
	}

	return resp, nil
}
