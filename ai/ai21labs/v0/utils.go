package ai21labs

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type scriptOutput struct {
	Text       string `json:"text"`
	Model      string `json:"model"`
	TokenCount int    `json:"token_count"`
}

func getTokenCountWithPythonScript(text string, model string) (scriptOutput, error) {

	paramsJSON, err := json.Marshal(map[string]interface{}{
		"text":  text,
		"model": model,
	})
	var output scriptOutput

	if err != nil {
		return output, fmt.Errorf("failed to marshal params: %w", err)
	}

	cmdRunner := exec.Command(pythonInterpreter, "-c", pythonAI21labsTokenizer)
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

	return output, nil
}
