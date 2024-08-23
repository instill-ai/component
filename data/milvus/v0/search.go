package milvus

import (
	"fmt"

	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	searchPath             = "/v2/vectordb/entities/search"
	describeCollectionPath = "/v2/vectordb/collections/describe"
)

type SearchOutput struct {
	Status string `json:"status"`
	Result Result `json:"result"`
}

type Result struct {
	Ids      []string         `json:"ids"`
	Data     []map[string]any `json:"data"`
	Vectors  [][]float32      `json:"vectors"`
	Metadata []map[string]any `json:"metadata"`
}

type SearchInput struct {
	CollectionName string         `json:"collection-name"`
	PartitionName  string         `json:"partition-name"`
	Vector         []float32      `json:"vector"`
	Filter         string         `json:"filter"`
	Limit          int            `json:"limit"`
	VectorField    string         `json:"vector-field"`
	Offset         int            `json:"offset"`
	GroupingField  string         `json:"grouping-field"`
	Fields         []string       `json:"fields"`
	SearchParams   map[string]any `json:"search-params"`
}

type SearchReq struct {
	CollectionName string         `json:"collectionName"`
	PartitionName  string         `json:"partitionName"`
	Data           [][]float32    `json:"data"`
	Filter         string         `json:"filter"`
	Limit          int            `json:"limit"`
	AnnsField      string         `json:"annsField"`
	Offset         int            `json:"offset"`
	GroupingField  string         `json:"groupingField"`
	OutputFields   []string       `json:"outputFields"`
	SearchParams   map[string]any `json:"searchParams"`
}

type SearchResp struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    []map[string]any `json:"data"`
}

type DescribeCollection struct {
	CollectionName string `json:"collection-name"`
}

type DescribeCollectionReq struct {
	CollectionNameReq string `json:"collectionName"`
}

type DescribeCollectionResp struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    DataDescribe `json:"data"`
}

type DataDescribe struct {
	Fields []Field `json:"fields"`
}

type Field struct {
	Name       string `json:"name"`
	PrimaryKey bool   `json:"primaryKey"`
}

func (e *execution) search(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct SearchInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	respDescribe := DescribeCollectionResp{}

	reqParamsDescribe := DescribeCollectionReq{
		CollectionNameReq: inputStruct.CollectionName,
	}

	if e.Setup.Fields["username"].GetStringValue() == "mock-root" {
		respDescribe.Data.Fields = []Field{
			{
				Name:       "id",
				PrimaryKey: true,
			},
			{
				Name:       "name",
				PrimaryKey: false,
			},
			{
				Name:       "vector",
				PrimaryKey: false,
			},
		}
	} else {
		reqDescribe := e.client.R().SetBody(reqParamsDescribe).SetResult(&respDescribe)

		resDescribe, err := reqDescribe.Post(describeCollectionPath)

		if err != nil {
			return nil, err
		}

		if resDescribe.StatusCode() != 200 {
			return nil, fmt.Errorf("failed to describe collection: %s", resDescribe.String())
		}
	}

	if respDescribe.Message != "" && respDescribe.Code != 200 {
		return nil, fmt.Errorf("failed to describe collection: %s", respDescribe.Message)
	}

	var primaryKeyField string
	for _, field := range respDescribe.Data.Fields {
		if field.PrimaryKey {
			primaryKeyField = field.Name
			break
		}
	}

	var fields []string
	if inputStruct.Fields == nil {
		for _, field := range respDescribe.Data.Fields {
			fields = append(fields, field.Name)
		}
	}

	resp := SearchResp{}

	reqParams := SearchReq{
		CollectionName: inputStruct.CollectionName,
		Data:           [][]float32{inputStruct.Vector},
		Limit:          inputStruct.Limit,
		AnnsField:      inputStruct.VectorField,
	}
	if inputStruct.PartitionName != "" {
		reqParams.PartitionName = inputStruct.PartitionName
	}
	if fields != nil {
		reqParams.Filter = inputStruct.Filter
	}
	if inputStruct.Offset != 0 {
		reqParams.Offset = inputStruct.Offset
	}
	if inputStruct.GroupingField != "" {
		reqParams.GroupingField = inputStruct.GroupingField
	}
	if len(inputStruct.Fields) > 0 {
		reqParams.OutputFields = inputStruct.Fields
	}
	if inputStruct.SearchParams != nil {
		reqParams.SearchParams = inputStruct.SearchParams
	}

	req := e.client.R().SetBody(reqParams).SetResult(&resp)

	res, err := req.Post(searchPath)

	if err != nil {
		return nil, err
	}

	if res.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to Search point: %s", res.String())
	}

	if resp.Message != "" && resp.Code != 0 {
		return nil, fmt.Errorf("failed to upsert data: %s", resp.Message)
	}

	var ids []string
	var metadata []map[string]any
	var vectors [][]float32
	data := resp.Data

	for _, d := range data {
		var vectorFloat32 []float32
		for _, v := range d[inputStruct.VectorField].([]any) {
			vectorFloat32 = append(vectorFloat32, float32(v.(float64)))
		}

		ids = append(ids, fmt.Sprintf("%v", d[primaryKeyField]))
		vectors = append(vectors, vectorFloat32)

		metadatum := make(map[string]any)
		for _, field := range respDescribe.Data.Fields {
			if field.Name == "distance" || field.Name == inputStruct.VectorField {
				continue
			}
			metadatum[field.Name] = d[field.Name]
		}
		metadata = append(metadata, metadatum)
	}

	outputStruct := SearchOutput{
		Status: fmt.Sprintf("Successfully searched %d data", len(resp.Data)),
		Result: Result{
			Ids:      ids,
			Data:     data,
			Vectors:  vectors,
			Metadata: metadata,
		},
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}

	return output, nil
}
