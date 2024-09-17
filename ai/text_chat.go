package ai

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
