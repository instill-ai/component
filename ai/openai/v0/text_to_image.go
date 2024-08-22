package openai

type ImagesGenerationInput struct {
	Prompt  string `json:"prompt"`
	Model   string `json:"model"`
	N       int    `json:"n,omitempty"`
	Quality string `json:"quality,omitempty"`
	Size    string `json:"size,omitempty"`
	Style   string `json:"style,omitempty"`
}

type ImageGenerationsOutputResult struct {
	Image         string `json:"image"`
	RevisedPrompt string `json:"revised-prompt"`
}
type ImageGenerationsOutput struct {
	Results []ImageGenerationsOutputResult `json:"results"`
}
