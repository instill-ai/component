package openai

type TextEmbeddingsInput struct {
	Text       string `json:"text"`
	Model      string `json:"model"`
	Dimensions int    `json:"dimensions"`
}

type TextEmbeddingsOutput struct {
	Embedding []float32 `json:"embedding"`
}
