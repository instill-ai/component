package cohere

import (
	"fmt"

	cohereSDK "github.com/cohere-ai/cohere-go/v2"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

type embeddingInput struct {
	Text          string `json:"text"`
	ModelName     string `json:"model-name"`
	InputType     string `json:"input-type"`
	EmbeddingType string `json:"embedding-type"`
}

type embeddingFloatOutput struct {
	Usage     embedUsage `json:"usage"`
	Embedding []float64  `json:"embedding"`
}

type embeddingIntOutput struct {
	Usage     embedUsage `json:"usage"`
	Embedding []int      `json:"embedding"`
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

	embeddingTypeArray := []cohereSDK.EmbeddingType{}
	switch inputStruct.EmbeddingType {
	case "float":
		embeddingTypeArray = append(embeddingTypeArray, cohereSDK.EmbeddingTypeFloat)
	case "int8":
		embeddingTypeArray = append(embeddingTypeArray, cohereSDK.EmbeddingTypeInt8)
	case "uint8":
		embeddingTypeArray = append(embeddingTypeArray, cohereSDK.EmbeddingTypeUint8)
	case "binary":
		embeddingTypeArray = append(embeddingTypeArray, cohereSDK.EmbeddingTypeBinary)
	case "ubinary":
		embeddingTypeArray = append(embeddingTypeArray, cohereSDK.EmbeddingTypeUbinary)
	}
	req := cohereSDK.EmbedRequest{
		Texts:          []string{inputStruct.Text},
		Model:          &inputStruct.ModelName,
		InputType:      (*cohereSDK.EmbedInputType)(&inputStruct.InputType),
		EmbeddingTypes: embeddingTypeArray,
	}
	resp, err := e.client.generateEmbedding(req)
	if err != nil {
		return nil, err
	}

	switch inputStruct.EmbeddingType {
	case "int8", "uint8", "binary", "ubinary":
		bills := resp.EmbeddingsByType.Meta.BilledUnits
		outputStruct := embeddingIntOutput{
			Usage: embedUsage{
				Tokens: int(*bills.InputTokens),
			},
		}
		switch inputStruct.EmbeddingType {
		case "int8":
			outputStruct.Embedding = resp.EmbeddingsByType.Embeddings.Int8[0]
		case "uint8":
			outputStruct.Embedding = resp.EmbeddingsByType.Embeddings.Uint8[0]
		case "binary":
			outputStruct.Embedding = resp.EmbeddingsByType.Embeddings.Binary[0]
		case "ubinary":
			outputStruct.Embedding = resp.EmbeddingsByType.Embeddings.Ubinary[0]
		}
		output, err := base.ConvertToStructpb(outputStruct)
		if err != nil {
			return nil, err
		}
		return output, nil
	case "float":
		bills := resp.EmbeddingsByType.Meta.BilledUnits
		outputStruct := embeddingFloatOutput{
			Usage: embedUsage{
				Tokens: int(*bills.InputTokens),
			},
			Embedding: resp.EmbeddingsByType.Embeddings.Float[0],
		}
		output, err := base.ConvertToStructpb(outputStruct)
		if err != nil {
			return nil, err
		}
		return output, nil
	default:
		bills := resp.EmbeddingsFloats.Meta.BilledUnits
		outputStruct := embeddingFloatOutput{
			Usage: embedUsage{
				Tokens: int(*bills.InputTokens),
			},
			Embedding: resp.EmbeddingsFloats.Embeddings[0],
		}
		output, err := base.ConvertToStructpb(outputStruct)
		if err != nil {
			return nil, err
		}
		return output, nil
	}
}
