package document

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/instill-ai/component/base"
)

type converterOutput struct {
	Body   string   `json:"body"`
	Images []string `json:"images"`
}

func convertPDFToMarkdownWithPDFPlumber(base64Text string, displayImageTag bool) (converterOutput, error) {

	paramsJSON, err := json.Marshal(map[string]interface{}{
		"PDF":               base.TrimBase64Mime(base64Text),
		"display-image-tag": displayImageTag,
	})
	var output converterOutput

	if err != nil {
		return output, fmt.Errorf("failed to marshal params: %w", err)
	}

	cmdRunner := exec.Command(pythonInterpreter, "-c", pythonPDFPlumberConverter)
	stdin, err := cmdRunner.StdinPipe()

	if err != nil {
		return output, fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	errChan := make(chan error, 1)
	go func() {
		defer stdin.Close()
		_, err := stdin.Write(paramsJSON)
		if err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	outputBytes, err := cmdRunner.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("failed to run python script: %w", err)
	}

	writeErr := <-errChan
	if writeErr != nil {
		return output, fmt.Errorf("failed to write to stdin: %w", writeErr)
	}

	err = json.Unmarshal(outputBytes, &output)
	if err != nil {
		return output, fmt.Errorf("failed to unmarshal output: %w", err)
	}

	// TODO: Take it off
	fmt.Println("===== \n\n\n output", output, "\n\n\n =====")

	return output, nil
}
