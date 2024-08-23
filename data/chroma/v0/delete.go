package chroma

import (
	"fmt"

	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	deletePath = "/api/v1/collections/%s/delete"
)

type DeleteOutput struct {
	Status string `json:"status"`
}

type DeleteInput struct {
	CollectionName string         `json:"collection-name"`
	ID             string         `json:"id"`
	Filter         map[string]any `json:"filter"`
	FilterDocument map[string]any `json:"filter-document"`
}

type DeleteReq struct {
	IDs           []string       `json:"ids"`
	Where         map[string]any `json:"where"`
	WhereDocument map[string]any `json:"where_document"`
}

func (e *execution) delete(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct DeleteInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	var resp []string

	// one of id or filter or filter document should be present otherwise if all empty then error
	if inputStruct.ID == "" && inputStruct.Filter == nil && inputStruct.FilterDocument == nil {
		return nil, fmt.Errorf("one of id or filter or filter document should be present")
	}

	reqParams := DeleteReq{
		IDs:           []string{inputStruct.ID},
		Where:         inputStruct.Filter,
		WhereDocument: inputStruct.FilterDocument,
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

	res, err := req.Post(fmt.Sprintf(deletePath, collID))

	if err != nil {
		return nil, err
	}

	if res.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to delete item: %s", res.String())
	}

	if len(resp) == 0 {
		return nil, fmt.Errorf("failed to delete item")
	}

	outputStruct := DeleteOutput{
		Status: fmt.Sprintf("Successfully deleted %d items", len(resp)),
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}

	return output, nil
}