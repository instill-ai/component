//go:generate compogen readme ./config ./README.mdx
package image

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"sync"

	_ "embed" // embed
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	once      sync.Once
	comp      *component
)

// Operator is the derived operator
type component struct {
	base.Component
}

// Execution is the derived execution
type execution struct {
	base.ComponentExecution
}

// Init initializes the operator
func Init(bc base.Component) *component {
	once.Do(func() {
		comp = &component{Component: bc}
		err := comp.LoadDefinition(definitionJSON, nil, tasksJSON, nil)
		if err != nil {
			panic(err)
		}
	})
	return comp
}

func (c *component) CreateExecution(sysVars map[string]any, setup *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	return &base.ExecutionWrapper{Execution: &execution{
		ComponentExecution: base.ComponentExecution{Component: c, SystemVariables: sysVars, Task: task},
	}}, nil
}

// Execute executes the derived execution
func (e *execution) Execute(_ context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := []*structpb.Struct{}
	var base64ByteImg []byte
	for _, input := range inputs {

		b, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(input.Fields["image"].GetStringValue()))
		if err != nil {
			return nil, err
		}

		img, _, err := image.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}

		switch e.Task {
		case "TASK_DRAW_CLASSIFICATION":
			base64ByteImg, err = drawClassification(img, input.Fields["category"].GetStringValue(), input.Fields["score"].GetNumberValue())
			if err != nil {
				return nil, err
			}
		case "TASK_DRAW_DETECTION":
			base64ByteImg, err = drawDetection(img, input.Fields["objects"].GetListValue().Values)
			if err != nil {
				return nil, err
			}
		case "TASK_DRAW_KEYPOINT":
			base64ByteImg, err = drawKeypoint(img, input.Fields["objects"].GetListValue().Values)
			if err != nil {
				return nil, err
			}
		case "TASK_DRAW_OCR":
			base64ByteImg, err = drawOCR(img, input.Fields["objects"].GetListValue().Values)
			if err != nil {
				return nil, err
			}
		case "TASK_DRAW_INSTANCE_SEGMENTATION":
			base64ByteImg, err = drawInstanceSegmentation(img, input.Fields["objects"].GetListValue().Values)
			if err != nil {
				return nil, err
			}
		case "TASK_DRAW_SEMANTIC_SEGMENTATION":
			base64ByteImg, err = drawSemanticSegmentation(img, input.Fields["stuffs"].GetListValue().Values)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("not supported task: %s", e.Task)
		}

		output := structpb.Struct{Fields: make(map[string]*structpb.Value)}

		output.Fields["image"] = &structpb.Value{
			Kind: &structpb.Value_StringValue{
				StringValue: fmt.Sprintf("data:image/jpeg;base64,%s", string(base64ByteImg)),
			},
		}

		outputs = append(outputs, &output)
	}
	return outputs, nil
}
