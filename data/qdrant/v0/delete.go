package qdrant

import (
	"fmt"

	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	deletePath = "/collections/%s/points/delete?wait=true&ordering=%s"
)

type DeleteInput struct {
	ID             string         `json:"id"`
	CollectionName string         `json:"collection-name"`
	Filter         map[string]any `json:"filter"`
	Ordering       string         `json:"ordering"`
}

type DeleteOutput struct {
	Status string `json:"status"`
}

type DeleteReq struct {
	Points []string       `json:"points"`
	Filter map[string]any `json:"filter"`
}

type DeleteResp struct {
	Time   float64           `json:"time"`
	Status string            `json:"status"`
	Result BatchUpsertResult `json:"result"`
}

func (e *execution) delete(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct DeleteInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	resp := DeleteResp{}

	reqParams := DeleteReq{}
	if inputStruct.ID != "" {
		reqParams.Points = []string{inputStruct.ID}
	}
	if inputStruct.Filter != nil {
		reqParams.Filter = inputStruct.Filter
	}

	req := e.client.R().SetBody(reqParams).SetResult(&resp)

	res, err := req.Post(fmt.Sprintf(deletePath, inputStruct.CollectionName, inputStruct.Ordering))

	if err != nil {
		return nil, err
	}

	if res.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to delete points: %s", res.String())
	}

	outputStruct := DeleteOutput{
		Status: "Successfully deleted points",
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}

	return output, nil
}
