package instill

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

func (e *execution) executeImageToImage(grpcClient modelPB.ModelPublicServiceClient, nsID string, modelID string, version string, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s/%s/%s", inputs, nsID, modelID, version)
	}

	if grpcClient == nil {
		return nil, fmt.Errorf("uninitialized client")
	}

	outputs := []*structpb.Struct{}

	for _, input := range inputs {

		prompt := input.GetFields()["prompt"].GetStringValue()
		imageToImageInput := &modelPB.ImageToImageInput{
			Prompt: &prompt,
		}
		if _, ok := input.GetFields()["steps"]; ok {
			v := int32(input.GetFields()["steps"].GetNumberValue())
			imageToImageInput.Steps = &v
		}
		if _, ok := input.GetFields()["image-base64"]; ok {
			imageToImageInput.Type = &modelPB.ImageToImageInput_PromptImageBase64{
				PromptImageBase64: base.TrimBase64Mime(input.GetFields()["image-base64"].GetStringValue()),
			}
		}
		if _, ok := input.GetFields()["temperature"]; ok {
			v := int32(input.GetFields()["temperature"].GetNumberValue())
			imageToImageInput.Seed = &v
		}
		if _, ok := input.GetFields()["cfg-scale"]; ok {
			v := float32(input.GetFields()["cfg-scale"].GetNumberValue())
			imageToImageInput.CfgScale = &v
		}

		if _, ok := input.GetFields()["seed"]; ok {
			v := int32(input.GetFields()["seed"].GetNumberValue())
			imageToImageInput.Seed = &v
		}

		taskInput := &modelPB.TaskInput_ImageToImage{
			ImageToImage: imageToImageInput,
		}

		// only support batch 1
		taskOutputs, err := trigger(grpcClient, e.SystemVariables, nsID, modelID, version, []*modelPB.TaskInput{{Input: taskInput}})
		if err != nil {
			return nil, err
		}
		if len(taskOutputs) <= 0 {
			return nil, fmt.Errorf("invalid output: %v for model: %s/%s/%s", taskOutputs, nsID, modelID, version)
		}

		imageToImageOutput := taskOutputs[0].GetImageToImage()
		if imageToImageOutput == nil {
			return nil, fmt.Errorf("invalid output: %v for model: %s/%s/%s", imageToImageOutput, nsID, modelID, version)
		}
		for imageIdx := range imageToImageOutput.Images {
			imageToImageOutput.Images[imageIdx] = fmt.Sprintf("data:image/jpeg;base64,%s", imageToImageOutput.Images[imageIdx])
		}

		outputJSON, err := protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		}.Marshal(imageToImageOutput)
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
