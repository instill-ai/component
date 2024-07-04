package cohere

import (
	"encoding/json"
	"fmt"

	cohereSDK "github.com/cohere-ai/cohere-go/v2"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

type embeddingInput struct {
	Text      string `json:"text"`
	ModelName string `json:"model-name"`
	InputType string `json:"input-type"`
}

type embeddingOutput struct {
	Embedding []float64  `json:"embedding"`
	Usage     embedUsage `json:"usage"`
}

type embedUsage struct {
	Tokens int `json:"tokens"`
}

func (e *execution) taskEmbedding(in *structpb.Struct) (*structpb.Struct, error) {
	inputStruct := embeddingInput{}
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, fmt.Errorf("error generating input struct: %v", err)
	}
	req := cohereSDK.EmbedRequest{
		Texts:     []string{inputStruct.Text},
		Model:     &inputStruct.ModelName,
		InputType: (*cohereSDK.EmbedInputType)(&inputStruct.InputType),
	}
	resp, err := e.client.generateEmbedding(req)

	if err != nil {
		return nil, err
	}

	if resp.EmbeddingsFloats.Meta == nil {
		return nil, fmt.Errorf("meta is nil")
	}
	bills := resp.EmbeddingsFloats.Meta.BilledUnits
	if bills == nil || bills.InputTokens == nil {
		return nil, fmt.Errorf("billed units is nil")
	}

	outputStruct := embeddingOutput{
		Embedding: resp.EmbeddingsFloats.Embeddings[0],
		Usage: embedUsage{
			Tokens: int(*bills.InputTokens),
		},
	}

	outputJSON, err := json.Marshal(outputStruct)
	if err != nil {
		return nil, err
	}
	output := structpb.Struct{}
	err = protojson.Unmarshal(outputJSON, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}
