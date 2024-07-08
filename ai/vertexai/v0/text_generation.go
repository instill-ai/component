package vertexai

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/vertexai/genai"
	"github.com/instill-ai/component/base"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

type textGenerationOutput struct {
	Text  string              `json:"text"`
	Usage textGenerationUsage `json:"usage"`
}
type textGenerationUsage struct {
	InputTokens  int `json:"input-tokens"`
	OutputTokens int `json:"output-tokens"`
}
type ChatMessage struct {
	Role    string              `json:"role"`
	Content []MultiModalContent `json:"content"`
}
type URL struct {
	URL string `json:"url"`
}

type MultiModalContent struct {
	ImageURL URL    `json:"image-url"`
	Text     string `json:"text"`
	Type     string `json:"type"`
}

type textGenerationInput struct {
	ChatHistory  []ChatMessage `json:"chat-history"`
	MaxNewTokens int           `json:"max-new-tokens"`
	ModelName    string        `json:"model-name"`
	Prompt       string        `json:"prompt"`
	PromptImages []string      `json:"prompt-images"`
	Seed         int           `json:"seed"`
	SystemMsg    string        `json:"system-message"`
	Temperature  float64       `json:"temperature"`
	TopK         int           `json:"top-k"`
}

func (e *execution) generateText(in *structpb.Struct) (*structpb.Struct, error) {
	setupStruct := vertexAISetup{}
	err := base.ConvertFromStructpb(e.GetSetup(), &setupStruct)
	if err != nil {
		return nil, err
	}
	inputStruct := textGenerationInput{}
	err = base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	// Temporary way to get the credential. It should be migrated to a more secure implementation.
	credJSON := []byte(setupStruct.Cred)
	client, err := genai.NewClient(ctx, setupStruct.ProjectID, setupStruct.Location, option.WithCredentialsJSON(credJSON))

	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	defer client.Close()

	model := client.GenerativeModel(inputStruct.ModelName)
	model.SetTemperature(float32(inputStruct.Temperature))
	model.SetMaxOutputTokens(int32(inputStruct.MaxNewTokens))
	model.SetTopK(float32(inputStruct.TopK))
	promptPart := genai.Text(inputStruct.Prompt)
	resp, err := model.GenerateContent(ctx, promptPart)
	if err != nil {
		return nil, fmt.Errorf("error generating content: %w", err)
	}

	outputStruct := textGenerationOutput{
		Text: concateResponse(resp),
		Usage: textGenerationUsage{
			InputTokens:  int(resp.UsageMetadata.PromptTokenCount),
			OutputTokens: int(resp.UsageMetadata.CandidatesTokenCount),
		},
	}

	outputJSON, err := json.Marshal(outputStruct)
	if err != nil {
		return nil, err
	}
	output := structpb.Struct{}
	err = protojson.Unmarshal(outputJSON, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func concateResponse(resp *genai.GenerateContentResponse) string {
	fullResponse := ""
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			fullResponse = fullResponse + fmt.Sprint(part)
		}
	}
	return fullResponse
}
