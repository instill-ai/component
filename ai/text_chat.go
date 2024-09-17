package ai

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/instill-ai/component/base"
)

type TextChatInput struct {
	Data      InputData `json:"data"`
	Parameter Parameter `json:"parameter"`
}

type InputData struct {
	Model    string         `json:"model"`
	Messages []InputMessage `json:"messages"`
}

type Parameter struct {
	MaxTokens   int     `json:"max-tokens"`
	Seed        int     `json:"seed"`
	N           int     `json:"n"`
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"top-p"`
	Stream      bool    `json:"stream"`
}

type InputMessage struct {
	Contents []Content `json:"content"`
	Role     string    `json:"role"`
	Name     string    `json:"name"`
}

type Content struct {
	Type        string `json:"type"`
	Text        string `json:"text,omitempty"`
	ImageURL    string `json:"image-url,omitempty"`
	ImageBase64 string `json:"image-base64,omitempty"`
}

type TextChatOutput struct {
	Data     OutputData `json:"data"`
	Metadata Metadata   `json:"metadata"`
}

type OutputData struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	FinishReason string        `json:"finish-reason"`
	Index        int           `json:"index"`
	Message      OutputMessage `json:"message"`
	// The Unix timestamp (in seconds) of when the chat completion was created.
	Created int `json:"created"`
}

type OutputMessage struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type Metadata struct {
	Usage Usage `json:"usage"`
}

type Usage struct {
	CompletionTokens int `json:"completion-tokens"`
	PromptTokens     int `json:"prompt-tokens"`
	TotalTokens      int `json:"total-tokens"`
}

func GetDataURL(base64Image string) string {

	if hasDataPrefix(base64Image) {
		return base64Image
	}

	b, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(base64Image))

	if err != nil {
		return base64Image
	}

	dataURL := fmt.Sprintf("data:%s;base64,%s", mimetype.Detect(b).String(), base.TrimBase64Mime(base64Image))

	return dataURL
}

func hasDataPrefix(base64Image string) bool {
	return strings.HasPrefix(base64Image, "data:")
}
