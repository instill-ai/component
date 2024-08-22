package openai

type AudioTranscriptionInput struct {
	Audio       string  `json:"audio"`
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt,omitempty"`
	Temperature float32 `json:"temperature,omitempty"`
	Language    string  `json:"language,omitempty"`
}
