package ai21labs

// Source https://docs.ai21.com/reference/text-segmentation-api-ref on 2024-07-21

const textSegmentationEndpoint = "/studio/v1/segmentation"

type TextSegmentationRequest struct {
	Source     string     `json:"source"`
	SourceType SourceType `json:"sourceType"`
}

type TextSegmentationSegment struct {
	SegementText string      `json:"segmentText"`
	SegmentType  SegmentType `json:"segmentType"`
}

type TextSegmentationResponse struct {
	ID       string                    `json:"id"`
	Segments []TextSegmentationSegment `json:"segments"`
}

func (c *AI21labsClient) TextSegmentation(req TextSegmentationRequest) (TextSegmentationResponse, error) {
	resp := TextSegmentationResponse{}
	httpReq := c.httpClient.R().SetResult(&resp).SetBody(req)
	if _, err := httpReq.Post(textSegmentationEndpoint); err != nil {
		return resp, err
	}
	return resp, nil
}
