package chroma

import (
	"fmt"

	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

type BatchUpsertOutput struct {
	Status string `json:"status"`
}

type BatchUpsertInput struct {
	CollectionName string           `json:"collection-name"`
	ArrayID        []string         `json:"array-id"`
	ArrayVector    [][]float64      `json:"array-vector"`
	ArrayMetadata  []map[string]any `json:"array-metadata"`
	ArrayURI       []string         `json:"array-uri"`
	ArrayDocument  []string         `json:"array-document"`
}

func (e *execution) batchUpsert(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct BatchUpsertInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	resp := UpsertResp{}

	reqParams := UpsertReq{
		Embeddings: inputStruct.ArrayVector,
		Metadatas:  inputStruct.ArrayMetadata,
		IDs:        inputStruct.ArrayID,
	}
	if inputStruct.ArrayDocument != nil {
		reqParams.Documents = inputStruct.ArrayDocument
	}
	if inputStruct.ArrayURI != nil {
		reqParams.Uris = inputStruct.ArrayURI
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

	res, err := req.Post(fmt.Sprintf(upsertPath, collID))

	if err != nil {
		return nil, err
	}

	if res.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to batch upsert item: %s", res.String())
	}

	if resp.Error != "" && resp.Message != "" {
		return nil, fmt.Errorf("failed to batch upsert item: %s", resp.Message)
	}

	outputStruct := UpsertOutput{
		Status: fmt.Sprintf("Successfully batch upserted %d items", len(inputStruct.ArrayID)),
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}

	return output, nil
}
