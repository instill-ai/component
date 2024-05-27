package instill

import (
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

type LLMInput struct {

	// The prompt text
	Prompt string
	// The prompt images
	PromptImages []*modelPB.PromptImage
	// The chat history
	ChatHistory []*modelPB.Message
	// The system message
	SystemMessage *string
	// The maximum number of tokens for model to generate
	MaxNewTokens *int32
	// The temperature for sampling
	Temperature *float32
	// Top k for sampling
	TopK *int32
	// The seed
	Seed *int32
	// The extra parameters
	ExtraParams *structpb.Struct
}

func (e *execution) convertLLMInput(input *structpb.Struct) *LLMInput {
	llmInput := &LLMInput{
		Prompt: input.GetFields()["prompt"].GetStringValue(),
	}

	if _, ok := input.GetFields()["system_message"]; ok {
		v := input.GetFields()["system_message"].GetStringValue()
		llmInput.SystemMessage = &v
	}

	if _, ok := input.GetFields()["prompt_images"]; ok {
		promptImages := []*modelPB.PromptImage{}
		for _, item := range input.GetFields()["prompt_images"].GetListValue().GetValues() {
			image := &modelPB.PromptImage{}
			image.Type = &modelPB.PromptImage_PromptImageBase64{
				PromptImageBase64: base.TrimBase64Mime(item.GetStringValue()),
			}
			promptImages = append(promptImages, image)
		}
		llmInput.PromptImages = promptImages
	}

	if _, ok := input.GetFields()["chat_history"]; ok {
		history := []*modelPB.Message{}
		for _, item := range input.GetFields()["chat_history"].GetListValue().GetValues() {
			contents := []*modelPB.MessageContent{}
			for _, contentItem := range item.GetStructValue().Fields["content"].GetListValue().GetValues() {
				t := contentItem.GetStructValue().Fields["type"].GetStringValue()
				content := &modelPB.MessageContent{
					Type: t,
				}
				if t == "text" {
					content.Content = &modelPB.MessageContent_Text{
						Text: contentItem.GetStructValue().Fields["text"].GetStringValue(),
					}
				} else {
					image := &modelPB.PromptImage{}
					image.Type = &modelPB.PromptImage_PromptImageBase64{
						PromptImageBase64: contentItem.GetStructValue().Fields["image_url"].GetStructValue().Fields["url"].GetStringValue(),
					}
					content.Content = &modelPB.MessageContent_ImageUrl{
						ImageUrl: &modelPB.ImageContent{
							ImageUrl: image,
						},
					}
				}
				contents = append(contents, content)
			}
			// Note: Instill Model require the order of chat_history be [user, assistant, user, assistant...]
			if len(history) == 0 && item.GetStructValue().Fields["role"].GetStringValue() != "user" {
				continue
			}
			if len(history) > 0 && history[len(history)-1].Role == item.GetStructValue().Fields["role"].GetStringValue() {
				for _, content := range contents {
					if content.Type == "text" {
						for cIdx := range history[len(history)-1].Content {
							if history[len(history)-1].Content[cIdx].Type == "text" {
								history[len(history)-1].Content[cIdx].Content = &modelPB.MessageContent_Text{
									Text: history[len(history)-1].Content[cIdx].GetText() + "\n" + content.GetText(),
								}
							}
						}
					} else {
						history[len(history)-1].Content = append(history[len(history)-1].Content, content)
					}
				}

			} else {
				history = append(history, &modelPB.Message{
					Role:    item.GetStructValue().Fields["role"].GetStringValue(),
					Content: contents,
				})
			}
		}
		llmInput.ChatHistory = history
	}

	if _, ok := input.GetFields()["max_new_tokens"]; ok {
		v := int32(input.GetFields()["max_new_tokens"].GetNumberValue())
		llmInput.MaxNewTokens = &v
	}
	if _, ok := input.GetFields()["temperature"]; ok {
		v := float32(input.GetFields()["temperature"].GetNumberValue())
		llmInput.Temperature = &v
	}
	if _, ok := input.GetFields()["top_k"]; ok {
		v := int32(input.GetFields()["top_k"].GetNumberValue())
		llmInput.TopK = &v
	}
	if _, ok := input.GetFields()["seed"]; ok {
		v := int32(input.GetFields()["seed"].GetNumberValue())
		llmInput.Seed = &v
	}
	if _, ok := input.GetFields()["extra_params"]; ok {
		v := input.GetFields()["extra_params"].GetStructValue()
		llmInput.ExtraParams = v
	}
	return llmInput

}
