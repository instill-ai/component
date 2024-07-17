package ollama

import (
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

type TaskTextGenerationChatInput struct {
	ChatHistory  []ChatMessage `json:"chat-history"`
	MaxNewTokens int           `json:"max-new-tokens"`
	ModelName    string        `json:"model-name"`
	Prompt       string        `json:"prompt"`
	PromptImages []string      `json:"prompt-images"`
	Seed         int           `json:"seed"`
	SystemMsg    string        `json:"system-message"`
	Temperature  float32       `json:"temperature"`
	TopK         int           `json:"top-k"`
}

type ChatMessage struct {
	Role    string              `json:"role"`
	Content []MultiModalContent `json:"content"`
}

type MultiModalContent struct {
	ImageURL URL    `json:"image-url"`
	Text     string `json:"text"`
	Type     string `json:"type"`
}

type URL struct {
	URL string `json:"url"`
}

type TaskTextGenerationChatOuput struct {
	Text  string                      `json:"text"`
	Usage TaskTextGenerationChatUsage `json:"usage"`
}

type TaskTextGenerationChatUsage struct {
	InputTokens  int `json:"input-tokens"`
	OutputTokens int `json:"output-tokens"`
}

func (e *execution) TaskTextGenerationChat(in *structpb.Struct) (*structpb.Struct, error) {
	input := TaskTextGenerationChatInput{}
	if err := base.ConvertFromStructpb(in, &input); err != nil {
		return nil, err
	}

	messages := []OllamaChatMessage{{
		Role:    "system",
		Content: input.SystemMsg},
	}

	for _, msg := range input.ChatHistory {
		textContent := ""
		imageContent := []string{}
		for _, content := range msg.Content {
			if content.Type == "text" {
				textContent = textContent + content.Text
			} else {
				imageContent = append(imageContent, content.ImageURL.URL)
			}
		}
		messages = append(messages, OllamaChatMessage{
			Role:    msg.Role,
			Content: textContent,
			Images:  imageContent,
		})
	}

	messages = append(messages, OllamaChatMessage{
		Role:    "user",
		Content: input.Prompt,
		Images:  input.PromptImages,
	})

	request := ChatRequest{
		Model:    input.ModelName,
		Messages: messages,
		Stream:   false,
		Options: OllamaOptions{
			Temperature: input.Temperature,
			TopK:        input.TopK,
			Seed:        input.Seed,
		},
	}

	response, err := e.client.Chat(request)
	if err != nil {
		return nil, err
	}

	output := TaskTextGenerationChatOuput{
		Text: response.Message.Content,
		Usage: TaskTextGenerationChatUsage{
			InputTokens:  response.PromptEvalCount - response.EvalCount,
			OutputTokens: response.EvalCount,
		},
	}
	return base.ConvertToStructpb(output)
}
