package vertexai

import (
	"context"
	"encoding/json"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/googleai/vertex"
	"go.uber.org/zap"
)

type VertexAIClient struct {
	langchainClient *vertex.Vertex
}

func NewClient(setup VertexAISetup, logger *zap.Logger) *VertexAIClient {
	ctx := context.Background()
	langchainClient, err := vertex.New(ctx, googleai.WithCloudLocation(setup.Location), googleai.WithCloudProject(setup.ProjectID), googleai.WithCredentialsJSON(json.RawMessage(setup.Cred)))
	if err != nil {
		logger.Fatal("failed to create langchain client", zap.Error(err))
	}
	return &VertexAIClient{
		langchainClient: langchainClient,
	}
}

type EmbedRequest struct {
	Text  string `json:"text"`
	Model string `json:"model"`
}

type EmbedResponse struct {
	Embedding []float32 `json:"embedding"`
	Tokens    int       `json:"tokens"`
}

func (c *VertexAIClient) Embed(req EmbedRequest) (EmbedResponse, error) {
	resp := EmbedResponse{}
	ctx := context.Background()
	// this function only support generate embedding with Palm, not with other models
	content, err := c.langchainClient.CreateEmbedding(ctx, []string{req.Text})
	if err != nil {
		return EmbedResponse{}, err
	}
	resp.Embedding = content[0]
	resp.Tokens = 0
	return resp, nil
}

type ChatRequest struct {
	Messages    []llms.MessageContent `json:"messages"`
	Model       string                `json:"model"`
	MaxTokens   int                   `json:"max_tokens"`
	Temperature float64               `json:"temperature"`
	TopP        float64               `json:"top_p"`
	TopK        int                   `json:"top_k"`
	Seed        int                   `json:"seed"`
}

type ChatResponse struct {
	Text         string `json:"text"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
}

func (c *VertexAIClient) Chat(req ChatRequest) (ChatResponse, error) {
	resp := ChatResponse{}
	ctx := context.Background()
	content, err := c.langchainClient.GenerateContent(ctx, req.Messages, llms.WithModel(req.Model), llms.WithMaxTokens(req.MaxTokens), llms.WithTemperature(req.Temperature), llms.WithTopP(req.TopP), llms.WithTopK(req.TopK), llms.WithSeed(req.Seed))
	if err != nil {
		return ChatResponse{}, err
	}
	resp.Text = content.Choices[0].Content
	print(content.Choices[0].GenerationInfo)
	resp.InputTokens = 0
	resp.OutputTokens = 0
	return resp, nil

}
