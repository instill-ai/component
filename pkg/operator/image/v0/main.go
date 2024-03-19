// TODO go:generate compogen readme --operator ./config ./README.mdx
package image

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"sync"

	_ "embed" // embed
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
)

var (
	//go:embed config/definitions.json
	definitionsJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	once      sync.Once
	operator  base.IOperator
)

// Operator is the derived operator
type Operator struct {
	base.Operator
}

// Execution is the derived execution
type Execution struct {
	base.Execution
}

// Init initializes the operator
func Init(logger *zap.Logger) base.IOperator {
	once.Do(func() {
		operator = &Operator{
			Operator: base.Operator{
				Component: base.Component{Logger: logger},
			},
		}
		err := operator.LoadOperatorDefinitions(definitionsJSON, tasksJSON, nil)
		if err != nil {
			logger.Fatal(err.Error())
		}
	})
	return operator
}

// CreateExecution creates the derived execution
func (o *Operator) CreateExecution(defUID uuid.UUID, task string, config *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	e := &Execution{}
	e.Execution = base.CreateExecutionHelper(e, o, defUID, task, config, logger)
	return e, nil
}

// Execute executes the derived execution
func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
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
