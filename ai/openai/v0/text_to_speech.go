package openai

type TextToSpeechInput struct {
	Text           string  `json:"text"`
	Model          string  `json:"model"`
	Voice          string  `json:"voice"`
	ResponseFormat string  `json:"response-format,omitempty"`
	Speed          float64 `json:"speed,omitempty"`
}

type TextToSpeechOutput struct {
	Audio string `json:"audio"`
}
