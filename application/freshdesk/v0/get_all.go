package freshdesk

import (
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

func (c *FreshdeskClient) GetAll(objectType string, pagination bool, paginationPath string) ([]TaskGetAllResponse, string, error) {

	resp := []TaskGetAllResponse{}

	httpReq := c.httpclient.R().SetResult(&resp)

	var rawResp *resty.Response
	var err error
	if !pagination {
		rawResp, err = httpReq.Get(fmt.Sprintf("/%s", strings.ToLower(objectType)))

	} else {
		rawResp, err = httpReq.Get(paginationPath)
	}

	if err != nil {
		return nil, "", err
	}

	// Will exist if there is a next page
	linkHeader := rawResp.Header().Get("Link")

	var nextPage string
	if linkHeader != "" {
		startIndex := strings.Index(linkHeader, "<")
		endIndex := strings.Index(linkHeader, ">")
		nextPage = linkHeader[startIndex+1 : endIndex]

		return resp, nextPage, nil
	}

	return resp, "", nil
}

// Task 1: Get All
type TaskGetAllInput struct {
	ObjectType string `json:"object-type"`
}

type TaskGetAllResponse struct {
	ID int64 `json:"id"`
}

type TaskGetAllOutput struct {
	IDs      []int64 `json:"ids"`
	IDLength int     `json:"id-length"`
}

func (e *execution) TaskGetAll(in *structpb.Struct) (*structpb.Struct, error) {
	inputStruct := TaskGetAllInput{}
	err := base.ConvertFromStructpb(in, &inputStruct)

	if err != nil {
		return nil, fmt.Errorf("failed to convert input to struct: %v", err)
	}

	resp, paginationPath, err := e.client.GetAll(inputStruct.ObjectType, false, "")

	if err != nil {
		return nil, err
	}

	outputStruct := TaskGetAllOutput{}
	outputStruct.IDs = make([]int64, len(resp))
	for index, value := range resp {
		outputStruct.IDs[index] = value.ID
	}

	for paginationPath != "" {
		respPage, nextPage, err := e.client.GetAll(inputStruct.ObjectType, true, paginationPath)

		if err != nil {
			return nil, err
		}

		for _, value := range respPage {
			outputStruct.IDs = append(outputStruct.IDs, value.ID)
		}
		paginationPath = nextPage
	}

	outputStruct.IDLength = len(outputStruct.IDs)

	output, err := base.ConvertToStructpb(outputStruct)

	if err != nil {
		return nil, fmt.Errorf("failed to convert output to struct: %v", err)
	}

	return output, nil
}
