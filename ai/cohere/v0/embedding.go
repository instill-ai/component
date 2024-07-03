package cohere

import (
	"encoding/json"

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
	Embedding []float64   `json:"embedding"`
	Usage     cohereUsage `json:"usage"`
}

func (e *execution) taskEmbedding(in *structpb.Struct) (*structpb.Struct, error) {
	inputStruct := embeddingInput{}
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
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

	outputStruct := embeddingOutput{
		Embedding: resp.EmbeddingsFloats.Embeddings[0],
		Usage: cohereUsage{
			InputTokens:  int(*resp.EmbeddingsFloats.Meta.BilledUnits.InputTokens),
			OutputTokens: int(*resp.EmbeddingsFloats.Meta.BilledUnits.OutputTokens),
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
