package instill

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

func (e *execution) executeOCR(grpcClient modelPB.ModelPublicServiceClient, nsID string, modelID string, version string, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s/%s/%s", inputs, nsID, modelID, version)
	}

	if grpcClient == nil {
		return nil, fmt.Errorf("uninitialized client")
	}

	outputs := []*structpb.Struct{}
	for _, input := range inputs {
		taskInput := &modelPB.TaskInput_Ocr{
			Ocr: &modelPB.OcrInput{
				Type: &modelPB.OcrInput_ImageBase64{
					ImageBase64: base.TrimBase64Mime(input.Fields["image-base64"].GetStringValue()),
				},
			},
		}

		// only support batch 1
		taskOutputs, err := trigger(grpcClient, e.SystemVariables, nsID, modelID, version, []*modelPB.TaskInput{{Input: taskInput}})
		if err != nil {
			return nil, err
		}
		if len(taskOutputs) <= 0 {
			return nil, fmt.Errorf("invalid output: %v for model: %s/%s/%s", taskOutputs, nsID, modelID, version)
		}

		ocrOutput := taskOutputs[0].GetOcr()
		if ocrOutput == nil {
			return nil, fmt.Errorf("invalid output: %v for model: %s/%s/%s", ocrOutput, nsID, modelID, version)
		}
		objects := make([]any, len(ocrOutput.Objects))

		for i := range ocrOutput.Objects {
			objects[i] = map[string]any{
				"text":  ocrOutput.Objects[i].Text,
				"score": ocrOutput.Objects[i].Score,
				"bounding-box": map[string]any{
					"top":    ocrOutput.Objects[i].BoundingBox.Top,
					"left":   ocrOutput.Objects[i].BoundingBox.Left,
					"width":  ocrOutput.Objects[i].BoundingBox.Width,
					"height": ocrOutput.Objects[i].BoundingBox.Height,
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
