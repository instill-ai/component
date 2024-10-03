package ai21labs

// Source https://docs.ai21.com/reference/paraphrase-ref on 2024-07-21

const paraphraseEndpoint = "/studio/v1/paraphrase"

type ParaphraseRequest struct {
	Text  string          `json:"text"`
	Style ParaphraseStyle `json:"style"`
	// note: conflicting name convention
	StartIndex int `json:"startIndex,omitempty"`
	EndIndex   int `json:"endIndex,omitempty"`
}

type ParaphraseStyle string

const (
	General ParaphraseStyle = "general"
	Casual  ParaphraseStyle = "casual"
	Formal  ParaphraseStyle = "formal"
	Long    ParaphraseStyle = "long"
	Short   ParaphraseStyle = "short"
)

type ParaphraseSuggestion struct {
	Text string `json:"text"`
}

type ParaphraseResponse struct {
	ID          string                 `json:"id"`
	Suggestions []ParaphraseSuggestion `json:"suggestions"`
}

func (c *AI21labsClient) Paraphrase(req ParaphraseRequest) (ParaphraseResponse, error) {
	resp := ParaphraseResponse{}
	httpReq := c.httpClient.R().SetResult(&resp).SetBody(req)
	if _, err := httpReq.Post(paraphraseEndpoint); err != nil {
		return resp, err
	}
	return resp, nil
}
