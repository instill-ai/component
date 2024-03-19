package instill

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

func (e *Execution) executeTextToImage(grpcClient modelPB.ModelPublicServiceClient, modelName string, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
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
		if _, ok := input.GetFields()["image_base64"]; ok {
			textToImageInput.Type = &modelPB.TextToImageInput_PromptImageBase64{
				PromptImageBase64: base.TrimBase64Mime(input.GetFields()["image_base64"].GetStringValue()),
			}
		}
		if _, ok := input.GetFields()["cfg_scale"]; ok {
			v := float32(input.GetFields()["cfg_scale"].GetNumberValue())
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
		req := modelPB.TriggerUserModelRequest{
			Name:       modelName,
			TaskInputs: []*modelPB.TaskInput{{Input: taskInput}},
		}
		ctx := metadata.NewOutgoingContext(context.Background(), getRequestMetadata(e.Config))
		res, err := grpcClient.TriggerUserModel(ctx, &req)
		if err != nil || res == nil {
			return nil, err
		}
		taskOutputs := res.GetTaskOutputs()
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
