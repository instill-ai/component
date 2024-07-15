package mistralai

import (
	"fmt"

	mistralSDK "github.com/gage-technologies/mistral-go"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

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
	TopP         float64       `json:"top-p"`
	Safe         bool          `json:"safe"`
}

type chatUsage struct {
	InputTokens  int `json:"input-tokens"`
	OutputTokens int `json:"output-tokens"`
}

type textGenerationOutput struct {
	Text  string    `json:"text"`
	Usage chatUsage `json:"usage"`
}

type textEmbeddingInput struct {
	Text      string `json:"text"`
	ModelName string `json:"model-name"`
}

type textEmbeddingUsage struct {
	Tokens int `json:"tokens"`
}

type textEmbeddingOutput struct {
	Embedding []float64          `json:"embedding"`
	Usage     textEmbeddingUsage `json:"usage"`
}

func (e *execution) taskTextGeneration(in *structpb.Struct) (*structpb.Struct, error) {

	inputStruct := textGenerationInput{}
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, fmt.Errorf("error generating input struct: %v", err)
	}

	messages := []mistralSDK.ChatMessage{}

	if inputStruct.SystemMsg != "" {
		messages = append(messages, mistralSDK.ChatMessage{
			Role:    "system",
			Content: inputStruct.SystemMsg,
		})
	}
	for _, chatMessage := range inputStruct.ChatHistory {
		messageContent := ""
		for _, content := range chatMessage.Content {
			if content.Type == "text" {
				messageContent += content.Text
			}
		}
		if messageContent == "" {
			continue
		}
		messages = append(messages, mistralSDK.ChatMessage{
			Role:    chatMessage.Role,
			Content: messageContent,
		})
	}

	promptMessage := mistralSDK.ChatMessage{
		Role:    "user",
		Content: inputStruct.Prompt,
	}

	messages = append(messages, promptMessage)

	params := mistralSDK.ChatRequestParams{
		Temperature: inputStruct.Temperature,
		RandomSeed:  inputStruct.Seed,
		MaxTokens:   inputStruct.MaxNewTokens,
		TopP:        inputStruct.TopP,
		SafePrompt:  inputStruct.Safe,
	}

	resp, err := e.client.sdkClient.Chat(
		inputStruct.ModelName,
		messages,
		&params,
	)

	if err != nil {
		return nil, fmt.Errorf("error calling Chat: %v", err)
	}

	outputStruct := textGenerationOutput{}

	outputStruct.Text = resp.Choices[0].Message.Content
	outputStruct.Usage = chatUsage{
		InputTokens:  resp.Usage.PromptTokens,
		OutputTokens: resp.Usage.CompletionTokens,
	}
	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (e *execution) taskTextEmbedding(in *structpb.Struct) (*structpb.Struct, error) {
	inputStruct := textEmbeddingInput{}
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, fmt.Errorf("error generating input struct: %v", err)
	}

	resp, err := e.client.sdkClient.Embeddings(inputStruct.ModelName, []string{inputStruct.Text})
	if err != nil {
		return nil, fmt.Errorf("error calling Embeddings: %v", err)
	}
	outputStruct := textEmbeddingOutput{
		Embedding: resp.Data[0].Embedding,
		Usage: textEmbeddingUsage{
			Tokens: resp.Usage.TotalTokens,
		},
	}
	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil

}
