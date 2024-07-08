package vertexai

import (
	"encoding/json"

	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

type imageGenerationInput struct {
	CFGScale    float64 `json:"cfg-scale"`
	ModelName   string  `json:"model-name"`
	Prompt      string  `json:"prompt"`
	Samples     int     `json:"samples"`
	Seed        int     `json:"seed"`
	Steps       int     `json:"steps"`
	PromptImage string  `json:"prompt-image"`
}

type imageGenerationOutput struct {
	Images []string `json:"images"`
}

func (e *execution) generateImage(in *structpb.Struct) (*structpb.Struct, error) {
	setupStruct := vertexAISetup{}
	err := base.ConvertFromStructpb(e.GetSetup(), &setupStruct)
	if err != nil {
		return nil, err
	}
	inputStruct := imageGenerationInput{}
	err = base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}
	outputStruct := imageGenerationOutput{}
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
