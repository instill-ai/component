package vertexai

import (
	"fmt"

	"github.com/instill-ai/component/base"
	"github.com/tmc/langchaingo/llms"
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

type TaskTextGenerationChatInput struct {
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

type TaskTextGenerationChatUsage struct {
	InputTokens  int `json:"input-tokens"`
	OutputTokens int `json:"output-tokens"`
}

type TaskTextGenerationChatOutput struct {
	Text  string                      `json:"text"`
	Usage TaskTextGenerationChatUsage `json:"usage"`
}

type TaskTextEmbeddingsInput struct {
	Text      string `json:"text"`
	ModelName string `json:"model-name"`
}

type TaskTextEmbeddingsUsage struct {
	Tokens int `json:"tokens"`
}

type TaskTextEmbeddingsOutput struct {
	Embedding []float64               `json:"embedding"`
	Usage     TaskTextEmbeddingsUsage `json:"usage"`
}

func (e *execution) TaskTextGenerationChat(in *structpb.Struct) (*structpb.Struct, error) {

	inputStruct := TaskTextGenerationChatInput{}
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, fmt.Errorf("error generating input struct: %v", err)
	}

	messages := []llms.MessageContent{}

	if inputStruct.SystemMsg != "" {
		messages = append(messages, llms.MessageContent{
			Role: llms.ChatMessageType("system"), // note: not sure if this is correct, go-langchain does not have a system role
			Parts: []llms.ContentPart{
				llms.TextPart(inputStruct.SystemMsg),
			},
		})
	}

	for _, chatMessage := range inputStruct.ChatHistory {
		messageContent := []llms.ContentPart{}
		for _, content := range chatMessage.Content {
			if content.Type == "text" {
				messageContent = append(messageContent, llms.TextPart(content.Text))
			} else if content.Type == "image" {
				messageContent = append(messageContent, llms.ImageURLPart(content.ImageURL.URL))
			}
		}
		if len(messageContent) == 0 {
			continue
		}
		messages = append(messages, llms.MessageContent{
			Role:  llms.ChatMessageType(chatMessage.Role),
			Parts: messageContent,
		})
	}

	promptContent := []llms.ContentPart{}

	for _, content := range inputStruct.PromptImages {
		promptContent = append(promptContent, llms.ImageURLPart(content))
	}

	promptContent = append(promptContent, llms.TextPart(inputStruct.Prompt))

	messages = append(messages, llms.MessageContent{
		Role:  llms.ChatMessageType("user"),
		Parts: promptContent,
	})

	req := ChatRequest{
		Messages:    messages,
		Model:       inputStruct.ModelName,
		MaxTokens:   inputStruct.MaxNewTokens,
		Temperature: inputStruct.Temperature,
		TopP:        inputStruct.TopP,
		TopK:        inputStruct.TopK,
		Seed:        inputStruct.Seed,
	}
	resp, err := e.client.Chat(req)
	if err != nil {
		return nil, fmt.Errorf("error calling Chat: %v", err)
	}
	outputStruct := TaskTextGenerationChatOutput{
		Text: resp.Text,
		Usage: TaskTextGenerationChatUsage{
			InputTokens:  resp.InputTokens,
			OutputTokens: resp.OutputTokens,
		},
	}
	return base.ConvertToStructpb(outputStruct)
}

func (e *execution) TaskTextEmbeddings(in *structpb.Struct) (*structpb.Struct, error) {

	inputStruct := TaskTextEmbeddingsInput{}
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, fmt.Errorf("error generating input struct: %v", err)
	}
	outputStruct := TaskTextEmbeddingsOutput{}
	return base.ConvertToStructpb(outputStruct)
}
