package instill

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

func (e *execution) executeTextToImage(grpcClient modelPB.ModelPublicServiceClient, modelName string, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, modelName)
	}

	if grpcClient == nil {
		return nil, fmt.Errorf("uninitialized client")
	}

	outputs := []*structpb.Struct{}
	for _, input := range inputs {
		textToImageInput := &modelPB.TextToImageInput{
			Prompt: input.GetFields()["prompt"].GetStringValue(),
		}
		if _, ok := input.GetFields()["steps"]; ok {
			v := int32(input.GetFields()["steps"].GetNumberValue())
			textToImageInput.Steps = &v
		}
		if _, ok := input.GetFields()["image-base64"]; ok {
			textToImageInput.Type = &modelPB.TextToImageInput_PromptImageBase64{
				PromptImageBase64: base.TrimBase64Mime(input.GetFields()["image-base64"].GetStringValue()),
			}
		}
		if _, ok := input.GetFields()["cfg-scale"]; ok {
			v := float32(input.GetFields()["cfg-scale"].GetNumberValue())
			textToImageInput.CfgScale = &v
		}
		if _, ok := input.GetFields()["samples"]; ok {
			v := int32(input.GetFields()["samples"].GetNumberValue())
			textToImageInput.Samples = &v
		}
		if _, ok := input.GetFields()["seed"]; ok {
			v := int32(input.GetFields()["seed"].GetNumberValue())
			textToImageInput.Seed = &v
		}
		taskInput := &modelPB.TaskInput_TextToImage{
			TextToImage: textToImageInput,
		}

		// only support batch 1
		taskOutputs, err := trigger(grpcClient, e.SystemVariables, modelName, []*modelPB.TaskInput{{Input: taskInput}})
		if err != nil {
			return nil, err
		}
		if len(taskOutputs) <= 0 {
			return nil, fmt.Errorf("invalid output: %v for model: %s", taskOutputs, modelName)
		}

		textToImgOutput := taskOutputs[0].GetTextToImage()

		for imageIdx := range textToImgOutput.Images {
			textToImgOutput.Images[imageIdx] = fmt.Sprintf("data:image/jpeg;base64,%s", textToImgOutput.Images[imageIdx])
		}

		if textToImgOutput == nil || len(textToImgOutput.Images) <= 0 {
			return nil, fmt.Errorf("invalid output: %v for model: %s", textToImgOutput, modelName)
		}

		outputJSON, err := protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		}.Marshal(textToImgOutput)
		if err != nil {
			return nil, err
		}
		output := &structpb.Struct{}
		err = protojson.Unmarshal(outputJSON, output)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, output)
	}
	return outputs, nil
}
