package instill

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

func (e *execution) executeInstanceSegmentation(grpcClient modelPB.ModelPublicServiceClient, modelName string, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, modelName)
	}

	if grpcClient == nil {
		return nil, fmt.Errorf("uninitialized client")
	}

	taskInputs := []*modelPB.TaskInput{}
	for _, input := range inputs {
		segmentationInput := &modelPB.InstanceSegmentationInput{}
		segmentationInput.Type = &modelPB.InstanceSegmentationInput_ImageBase64{
			ImageBase64: base.TrimBase64Mime(input.Fields["image-base64"].GetStringValue()),
		}

		taskInput := &modelPB.TaskInput_InstanceSegmentation{
			InstanceSegmentation: segmentationInput,
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
		instanceSegmentationOp := taskOutputs[idx].GetInstanceSegmentation()
		if instanceSegmentationOp == nil {
			return nil, fmt.Errorf("invalid output: %v for model: %s", instanceSegmentationOp, modelName)
		}

		objects := make([]any, len(instanceSegmentationOp.Objects))

		for i := range instanceSegmentationOp.Objects {
			objects[i] = map[string]any{
				"category": instanceSegmentationOp.Objects[i].Category,
				"rle":      instanceSegmentationOp.Objects[i].Rle,
				"score":    instanceSegmentationOp.Objects[i].Score,
				"bounding-box": map[string]any{
					"top":    instanceSegmentationOp.Objects[i].BoundingBox.Top,
					"left":   instanceSegmentationOp.Objects[i].BoundingBox.Left,
					"width":  instanceSegmentationOp.Objects[i].BoundingBox.Width,
					"height": instanceSegmentationOp.Objects[i].BoundingBox.Height,
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
