package ai21labs

// Source https://docs.ai21.com/reference/embeddings-ref on 2024-07-21

const embeddingsEndpoint = "/studio/v1/embed"

type EmbeddingsRequest struct {
	Texts []string       `json:"texts"`
	Type  EmbeddingsType `json:"type"`
}

type EmbeddingsType string

const (
	Segment EmbeddingsType = "segment"
	Query   EmbeddingsType = "query"
)

type EmbeddingsResponse struct {
	ID      string      `json:"id"`
	Results [][]float32 `json:"results"`
}

func (c *AI21labsClient) Embeddings(req EmbeddingsRequest) (EmbeddingsResponse, error) {
	resp := EmbeddingsResponse{}
	httpReq := c.httpClient.R().SetResult(&resp).SetBody(req)
	if _, err := httpReq.Post(embeddingsEndpoint); err != nil {
		return resp, err
	}
	return resp, nil
}
