package openai

type textMessage struct {
	Role    string    `json:"role"`
	Content []content `json:"content"`
}

type TextCompletionInput struct {
	Prompt           string                     `json:"prompt"`
	Images           []string                   `json:"images"`
	ChatHistory      []*textMessage             `json:"chat-history,omitempty"`
	Model            string                     `json:"model"`
	SystemMessage    string                     `json:"system-message,omitempty"`
	Temperature      float32                    `json:"temperature,omitempty"`
	TopP             float32                    `json:"top-p,omitempty"`
	N                int                        `json:"n,omitempty"`
	MaxTokens        int                        `json:"max-tokens,omitempty"`
	PresencePenalty  float32                    `json:"presence-penalty,omitempty"`
	FrequencyPenalty float32                    `json:"frequency-penalty,omitempty"`
	ResponseFormat   *responseFormatInputStruct `json:"response-format,omitempty"`
}

type responseFormatInputStruct struct {
	Type       string `json:"type,omitempty"`
	JSONSchema string `json:"json-schema,omitempty"`
}

type TextCompletionOutput struct {
	Texts []string `json:"texts"`
	Usage usage    `json:"usage"`
}

type imageURL struct {
	URL string `json:"url"`
}

type content struct {
	Type     string   `json:"type"`
	Text     string   `json:"text,omitempty"`
	ImageURL imageURL `json:"image_url,omitempty"`
}

type usage struct {
	PromptTokens     int `json:"prompt-tokens"`
	CompletionTokens int `json:"completion-tokens"`
	TotalTokens      int `json:"total-tokens"`
}
