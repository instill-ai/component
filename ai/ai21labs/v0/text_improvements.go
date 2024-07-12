package ai21labs

// Source https://docs.ai21.com/reference/text-improvements-ref on 2024-07-21

const textImprovementEndpoint = "/studio/v1/improvements"

type TextImprovementsRequest struct {
	Text  string            `json:"text"`
	Types []ImprovementType `json:"types"`
}

type TextImprovement struct {
	Suggestions     []string        `json:"suggestions"`
	StartIndex      int             `json:"startIndex"`
	EndIndex        int             `json:"endIndex"`
	OriginalText    string          `json:"originalText"`
	ImprovementType ImprovementType `json:"improvementType"`
}

type ImprovementType string

const (
	Fluency        ImprovementType = "fluency"
	Specificity    ImprovementType = "vocabulary/specificity"
	Variety        ImprovementType = "vocabulary/variety"
	ShortSentences ImprovementType = "clarity/short-sentences"
	Conciseness    ImprovementType = "clarity/conciseness"
)

type TextImprovementsResponse struct {
	ID           string            `json:"id"`
	Improvements []TextImprovement `json:"improvements"`
}

func (c *AI21labsClient) TextImprovements(req TextImprovementsRequest) (TextImprovementsResponse, error) {
	resp := TextImprovementsResponse{}
	httpReq := c.httpClient.R().SetResult(&resp).SetBody(req)
	if _, err := httpReq.Post(textImprovementEndpoint); err != nil {
		return resp, err
	}
	return resp, nil
}
