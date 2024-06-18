package instill

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

func (e *execution) executeObjectDetection(grpcClient modelPB.ModelPublicServiceClient, modelName string, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, modelName)
	}

	if grpcClient == nil {
		return nil, fmt.Errorf("uninitialized client")
	}

	taskInputs := []*modelPB.TaskInput{}
	for _, input := range inputs {
		detectionInput := &modelPB.DetectionInput{}
		detectionInput.Type = &modelPB.DetectionInput_ImageBase64{
			ImageBase64: base.TrimBase64Mime(input.Fields["image-base64"].GetStringValue()),
		}

		modelInput := &modelPB.TaskInput_Detection{
			Detection: detectionInput,
		}
		taskInputs = append(taskInputs, &modelPB.TaskInput{Input: modelInput})
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
		objDetectionOutput := taskOutputs[idx].GetDetection()
		if objDetectionOutput == nil {
			return nil, fmt.Errorf("invalid output: %v for model: %s", objDetectionOutput, modelName)
		}
		objects := make([]any, len(objDetectionOutput.Objects))

		for i := range objDetectionOutput.Objects {
			objects[i] = map[string]any{
				"category": objDetectionOutput.Objects[i].Category,
				"score":    objDetectionOutput.Objects[i].Score,
				"bounding-box": map[string]any{
					"top":    objDetectionOutput.Objects[i].BoundingBox.Top,
					"left":   objDetectionOutput.Objects[i].BoundingBox.Left,
					"width":  objDetectionOutput.Objects[i].BoundingBox.Width,
					"height": objDetectionOutput.Objects[i].BoundingBox.Height,
				},
			}
		}
		output, err := structpb.NewStruct(map[string]any{
			"objects": objects,
		})
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, output)

	}

	return outputs, nil
}
