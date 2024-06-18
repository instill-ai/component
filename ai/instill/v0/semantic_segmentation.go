package instill

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

func (e *execution) executeSemanticSegmentation(grpcClient modelPB.ModelPublicServiceClient, modelName string, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, modelName)
	}
	if grpcClient == nil {
		return nil, fmt.Errorf("uninitialized client")
	}
	taskInputs := []*modelPB.TaskInput{}
	for _, input := range inputs {
		semanticSegmentationInput := &modelPB.SemanticSegmentationInput{}
		semanticSegmentationInput.Type = &modelPB.SemanticSegmentationInput_ImageBase64{
			ImageBase64: base.TrimBase64Mime(input.Fields["image-base64"].GetStringValue()),
		}

		taskInput := &modelPB.TaskInput_SemanticSegmentation{
			SemanticSegmentation: semanticSegmentationInput,
		}
		taskInputs = append(taskInputs, &modelPB.TaskInput{Input: taskInput})

	}

	taskOutputs, err := trigger(grpcClient, e.SystemVariables, modelName, taskInputs)
	if err != nil {
		return nil, err
	}
	if len(taskOutputs) <= 0 {
		return nil, fmt.Errorf("invalid output: %v for model: %s", taskOutputs, modelName)
	}

	outputs := []*structpb.Struct{}
	for idx := range inputs {
		semanticSegmentationOp := taskOutputs[idx].GetSemanticSegmentation()
		if semanticSegmentationOp == nil {
			return nil, fmt.Errorf("invalid output: %v for model: %s", semanticSegmentationOp, modelName)
		}
		outputJSON, err := protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		}.Marshal(semanticSegmentationOp)
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
