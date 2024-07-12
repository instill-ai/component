package ai21labs

// Source https://docs.ai21.com/reference/summarize-ref on 2024-07-21
const summarizeEndpoint = "/studio/v1/summarize"
const summarizeBySegmentEndpoint = "/studio/v1/summarize-by-segment"

type SummarizeRequest struct {
	Source     string     `json:"source"`
	SourceType SourceType `json:"sourceType"`
	Focus      string     `json:"focus"`
}

type SummarizeResponse struct {
	ID      string `json:"id"`
	Summary string `json:"summary"`
}

type SummerizeHighlight struct {
	Text       string `json:"text"`
	StartIndex int    `json:"startIndex"`
	EndIndex   int    `json:"endIndex"`
}

type SummerizeSegment struct {
	Summary     string               `json:"summary"`
	SegmentText string               `json:"segmentText"`
	SegmentHTML string               `json:"segmentHtml"`
	SegmentType SegmentType          `json:"segmentType"`
	HasSummary  bool                 `json:"hasSummary"`
	Highlights  []SummerizeHighlight `json:"highlights"`
}

type SummerizeBySegmentResponse struct {
	ID        string             `json:"id"`
	Segements []SummerizeSegment `json:"segments"`
}

func (c *AI21labsClient) Summarize(req SummarizeRequest) (SummarizeResponse, error) {
	resp := SummarizeResponse{}
	httpReq := c.httpClient.R().SetResult(&resp).SetBody(req)
	if _, err := httpReq.Post(summarizeEndpoint); err != nil {
		return resp, err
	}
	return resp, nil
}

func (c *AI21labsClient) SummarizeBySegment(req SummarizeRequest) (SummerizeBySegmentResponse, error) {
	resp := SummerizeBySegmentResponse{}
	httpReq := c.httpClient.R().SetResult(&resp).SetBody(req)
	if _, err := httpReq.Post(summarizeBySegmentEndpoint); err != nil {
		return resp, err
	}
	return resp, nil
}
