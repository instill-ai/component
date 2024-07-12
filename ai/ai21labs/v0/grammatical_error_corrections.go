package ai21labs

// Source https://docs.ai21.com/reference/gec-ref on 2024-07-21

const grammaticalErrorCorrectionEndpoint = "/studio/v1/gec"

type GrammaticalErrorCorrectionsRequest struct {
	Text string `json:"text"`
}

type GrammaticalErrorCorrection struct {
	Suggestion     string         `json:"suggestion"`
	StartIndex     int            `json:"startIndex"`
	EndIndex       int            `json:"endIndex"`
	OriginalText   string         `json:"originalText"`
	CorrectionType CorrectionType `json:"correctionType"`
}

type CorrectionType string

const (
	Grammar        CorrectionType = "Grammar"
	MissingWord    CorrectionType = "Missing Word"
	Punctuation    CorrectionType = "Punctuation"
	Spelling       CorrectionType = "Spelling"
	WordRepetition CorrectionType = "Word Repetition"
	WrongWord      CorrectionType = "Wrong Word"
)

type GrammaticalErrorCorrectionsResponse struct {
	ID          string                       `json:"id"`
	Corrections []GrammaticalErrorCorrection `json:"corrections"`
}

func (c *AI21labsClient) GrammaticalErrorCorrections(req GrammaticalErrorCorrectionsRequest) (GrammaticalErrorCorrectionsResponse, error) {
	resp := GrammaticalErrorCorrectionsResponse{}
	httpReq := c.httpClient.R().SetResult(&resp).SetBody(req)
	if _, err := httpReq.Post(grammaticalErrorCorrectionEndpoint); err != nil {
		return resp, err
	}
	return resp, nil
}
