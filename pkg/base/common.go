package base

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

func ConvertFromStructpb(from *structpb.Struct, to interface{}) error {
	inputJson, err := protojson.Marshal(from)
	if err != nil {
		return err
	}

	err = json.Unmarshal(inputJson, to)
	if err != nil {
		return err
	}
	return nil
}

func ConvertToStructpb(from interface{}) (*structpb.Struct, error) {
	to := &structpb.Struct{}
	outputJson, err := json.Marshal(from)
	if err != nil {
		return nil, err
	}

	err = protojson.Unmarshal(outputJson, to)
	if err != nil {
		return nil, err
	}
	return to, nil
}
