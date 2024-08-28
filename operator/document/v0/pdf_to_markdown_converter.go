package document

import (
	"encoding/json"
	"os/exec"

	"github.com/instill-ai/component/base"
)

func convertPDFToMarkdownWithPDFPlumber(base64Text string, displayImageTag bool) (string, error) {

	paramsJSON, err := json.Marshal(map[string]interface{}{
		"PDF":               base.TrimBase64Mime(base64Text),
		"display-image-tag": displayImageTag,
	})

	if err != nil {
		return "", err
	}

	cmdRunner := exec.Command(pythonInterpreter, "-c", pythonPDFPlumberConverter)
	stdin, err := cmdRunner.StdinPipe()

	if err != nil {
		return "", err
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
		return "", err
	}

	writeErr := <-errChan
	if writeErr != nil {
		return "", writeErr
	}

	var output pythonRunnerOutput
	err = json.Unmarshal(outputBytes, &output)
	if err != nil {
		return "", err
	}
	return output.Body, nil
}
