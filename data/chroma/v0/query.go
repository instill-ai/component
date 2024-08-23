package chroma

import (
	"fmt"

	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	queryPath = "/api/v1/collections/%s/query"
)

type QueryOutput struct {
	Status string `json:"status"`
	Result Result `json:"result"`
}

type Result struct {
	Ids      []string         `json:"ids"`
	Items    []map[string]any `json:"items"`
	Vectors  [][]float64      `json:"vectors"`
	Metadata []map[string]any `json:"metadata"`
}

type QueryInput struct {
	CollectionName string         `json:"collection-name"`
	Vector         []float64      `json:"vector"`
	Filter         map[string]any `json:"filter"`
	FilterDocument map[string]any `json:"filter-document"`
	NResults       int            `json:"n-results"`
}

type QueryReq struct {
	QueryEmbeddings []float64      `json:"query_embeddings"`
	Where           map[string]any `json:"where"`
	WhereDocument   map[string]any `json:"where_document"`
	NResults        int            `json:"n_results"`
}

type QueryResp struct {
	IDs        []string         `json:"ids"`
	Distances  []float64        `json:"distances"`
	Metadatas  []map[string]any `json:"metadatas"`
	Embeddings [][]float64      `json:"embeddings"`
	Documents  []string         `json:"documents"`

	Detail []map[string]any `json:"detail"`
}

func (e *execution) query(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct QueryInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	resp := QueryResp{}

	reqParams := QueryReq{
		QueryEmbeddings: inputStruct.Vector,
		NResults:        inputStruct.NResults,
	}
	if inputStruct.Filter != nil {
		reqParams.Where = inputStruct.Filter
	}
	if inputStruct.FilterDocument != nil {
		reqParams.WhereDocument = inputStruct.FilterDocument
	}

	var collID string
	if e.Setup.Fields["api-key"].GetStringValue() == "mock-api-key" {
		collID = "mock-collection-id"
	} else {
		collID, err = getCollectionID(inputStruct.CollectionName, e.client)
		if err != nil {
			return nil, err
		}
	}

	req := e.client.R().SetBody(reqParams).SetResult(&resp)

	res, err := req.Post(fmt.Sprintf(queryPath, collID))

	if err != nil {
		return nil, err
	}

	if res.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to query item: %s", res.String())
	}

	if resp.Detail != nil {
		return nil, fmt.Errorf("failed to query item: %s", resp.Detail[0]["msg"])
	}

	ids := resp.IDs
	metadatas := resp.Metadatas
	vectors := resp.Embeddings
	var items []map[string]any

	for i, metadata := range metadatas {
		item := make(map[string]any)
		for k, v := range metadata {
			if k != "id" {
				item[k] = v
			}
		}
		item["distance"] = resp.Distances[i]
		item["id"] = ids[i]
		item["vector"] = vectors[i]
		items = append(items, item)
	}

	outputStruct := QueryOutput{
		Status: fmt.Sprintf("Successfully queryed %d items", len(resp.IDs)),
		Result: Result{
			Ids:      ids,
			Items:    items,
			Vectors:  vectors,
			Metadata: metadatas,
		},
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}

	return output, nil
}