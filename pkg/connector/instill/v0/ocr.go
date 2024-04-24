package instill

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

func (e *execution) executeOCR(grpcClient modelPB.ModelPublicServiceClient, modelName string, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, modelName)
	}

	if grpcClient == nil {
		return nil, fmt.Errorf("uninitialized client")
	}

	outputs := []*structpb.Struct{}
	for _, input := range inputs {
		inputJSON, err := protojson.Marshal(input)
		if err != nil {
			return nil, err
		}
		ocrInput := &modelPB.OcrInput{}
		err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(inputJSON, ocrInput)
		if err != nil {
			return nil, err
		}

		taskInput := &modelPB.TaskInput_Ocr{
			Ocr: ocrInput,
		}

		// only support batch 1
		req := modelPB.TriggerUserModelRequest{
			Name:       modelName,
			TaskInputs: []*modelPB.TaskInput{{Input: taskInput}},
		}
		ctx := metadata.NewOutgoingContext(context.Background(), getRequestMetadata(e.SystemVariables))
		res, err := grpcClient.TriggerUserModel(ctx, &req)
		if err != nil || res == nil {
			return nil, err
		}
		taskOutputs := res.GetTaskOutputs()
		if len(taskOutputs) <= 0 {
			return nil, fmt.Errorf("invalid output: %v for model: %s", taskOutputs, modelName)
		}

		ocrOutput := taskOutputs[0].GetOcr()
		if ocrOutput == nil {
			return nil, fmt.Errorf("invalid output: %v for model: %s", ocrOutput, modelName)
		}
		outputJSON, err := protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		}.Marshal(ocrOutput)
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
