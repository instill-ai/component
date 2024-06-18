package instill

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

func (e *execution) executeKeyPointDetection(grpcClient modelPB.ModelPublicServiceClient, modelName string, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, modelName)
	}
	taskInputs := []*modelPB.TaskInput{}
	for _, input := range inputs {
		keypointInput := &modelPB.KeypointInput{}
		keypointInput.Type = &modelPB.KeypointInput_ImageBase64{
			ImageBase64: base.TrimBase64Mime(input.Fields["image-base64"].GetStringValue()),
		}

		taskInput := &modelPB.TaskInput_Keypoint{
			Keypoint: keypointInput,
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
		keyPointOutput := taskOutputs[idx].GetKeypoint()
		if keyPointOutput == nil {
			return nil, fmt.Errorf("invalid output: %v for model: %s", keyPointOutput, modelName)
		}
		objects := make([]any, len(keyPointOutput.Objects))

		for i := range keyPointOutput.Objects {
			keypoints := make([]any, len(keyPointOutput.Objects[i].Keypoints))
			for j := range keyPointOutput.Objects[i].Keypoints {
				keypoints[j] = map[string]any{
					"x": keyPointOutput.Objects[i].Keypoints[j].X,
					"y": keyPointOutput.Objects[i].Keypoints[j].Y,
					"v": keyPointOutput.Objects[i].Keypoints[j].V,
				}
			}
			objects[i] = map[string]any{
				"score": keyPointOutput.Objects[i].Score,
				"bounding-box": map[string]any{
					"top":    keyPointOutput.Objects[i].BoundingBox.Top,
					"left":   keyPointOutput.Objects[i].BoundingBox.Left,
					"width":  keyPointOutput.Objects[i].BoundingBox.Width,
					"height": keyPointOutput.Objects[i].BoundingBox.Height,
				},
				"keypoints": keypoints,
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
