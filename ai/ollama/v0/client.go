package ollama

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/instill-ai/component/internal/util/httpclient"
	"go.uber.org/zap"
)

// reference: https://github.com/ollama/ollama/blob/main/docs/api.md
// Ollama v0.2.5 on 2024-07-17

type errBody struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (e errBody) Message() string {
	return e.Error.Message
}

type OllamaClient struct {
	httpClient *httpclient.Client
	autoPull   bool
}

func NewClient(endpoint string, autoPull bool, logger *zap.Logger) *OllamaClient {
	c := httpclient.New("Ollama", endpoint, httpclient.WithLogger(logger),
		httpclient.WithEndUserError(new(errBody)))
	return &OllamaClient{httpClient: c, autoPull: autoPull}
}

type OllamaModelInfo struct {
	Name       string `json:"name"`
	ModifiedAt string `json:"modified_at"`
	Size       int    `json:"size"`
	Dijest     string `json:"digest"`
	Details    struct {
		Format            string `json:"format"`
		Family            string `json:"family"`
		Families          string `json:"families"`
		ParameterSize     string `json:"parameter_size"`
		QuantizationLevel string `json:"quantization_level"`
	} `json:"details"`
}

type ListLocalModelsRequest struct {
}

type ListLocalModelsResponse struct {
	Models []OllamaModelInfo `json:"models"`
}

func (c *OllamaClient) CheckModelAvailability(modelName string) bool {
	request := &ListLocalModelsRequest{}
	response := &ListLocalModelsResponse{}
	req := c.httpClient.R().SetResult(&response).SetBody(request)
	if _, err := req.Get("/api/tags"); err != nil {
		return false
	}
	localModels := []string{}
	for _, m := range response.Models {
		localModels = append(localModels, m.Name)
	}
	return slices.Contains(localModels, modelName)
}

type PullModelRequest struct {
	Name   string `json:"name"`
	Stream bool   `json:"stream"`
}

type PullModelResponse struct {
}

func (c *OllamaClient) Pull(modelName string) error {
	request := &PullModelRequest{
		Name:   modelName,
		Stream: false,
	}
	response := &PullModelResponse{}
	req := c.httpClient.R().SetResult(&response).SetBody(request)
	if _, err := req.Post("/api/pull"); err != nil {
		return err
	}
	return nil

}

type OllamaChatMessage struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images"`
}

type OllamaOptions struct {
	Temperature float32 `json:"temperature,omitempty"`
	TopK        int     `json:"top_k,omitempty"`
	Seed        int     `json:"seed,omitempty"`
}

type ChatRequest struct {
	Model    string              `json:"model"`
	Messages []OllamaChatMessage `json:"messages"`
	Stream   bool                `json:"stream"`
	Options  OllamaOptions       `json:"options"`
}

type ChatResponse struct {
	Model              string              `json:"model"`
	CreatedAt          string              `json:"created_at"`
	Message            []OllamaChatMessage `json:"message"`
	Done               bool                `json:"done"`
	Context            []int               `json:"context"`
	TotalDuration      int                 `json:"total_duration"`
	LoadDuration       int                 `json:"load_duration"`
	PromptEvalCount    int                 `json:"prompt_eval_count"`
	PromptEvalDuration int                 `json:"prompt_eval_duration"`
	EvalCount          int                 `json:"eval_count"`
	EvalDuration       int                 `json:"eval_duration"`
}

func (c *OllamaClient) Chat(request ChatRequest) (ChatResponse, error) {
	response := ChatResponse{}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return response, fmt.Errorf("error when marshalling request %v", err)
	}
	println("### request body: ", string(requestJSON))

	if !c.CheckModelAvailability(request.Model) && c.autoPull {
		err := c.Pull(request.Model)
		if err != nil {
			return response, fmt.Errorf("error when auto pulling model %v", err)
		}
	}
	req := c.httpClient.R().SetResult(&response).SetBody(request)
	if _, err := req.Post("/api/chat"); err != nil {
		return response, fmt.Errorf("error when sending chat request %v", err)
	}
	return response, nil
}
