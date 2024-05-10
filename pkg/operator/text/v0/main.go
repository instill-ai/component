//go:generate compogen readme --operator ./config ./README.mdx
package text

import (
	"context"
	"fmt"
	"sync"

	_ "embed" // embed

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
)

const (
	taskConvertToText string = "TASK_CONVERT_TO_TEXT"
	taskSplitByToken  string = "TASK_SPLIT_BY_TOKEN"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	once      sync.Once
	op        *operator
)

// Operator is the derived operator
type operator struct {
	base.Operator
}

// Execution is the derived execution
type execution struct {
	base.OperatorExecution
}

// Init initializes the operator
func Init(bo base.Operator) *operator {
	once.Do(func() {
		op = &operator{Operator: bo}
		err := op.LoadOperatorDefinition(definitionJSON, tasksJSON, nil)
		if err != nil {
			panic(err)
		}
	})
	return op
}

func (o *operator) CreateExecution(sysVars map[string]any, task string) (*base.ExecutionWrapper, error) {
	return &base.ExecutionWrapper{Execution: &execution{
		OperatorExecution: base.OperatorExecution{Operator: o, SystemVariables: sysVars, Task: task},
	}}, nil
}

// Execute executes the derived execution
func (e *execution) Execute(_ context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		switch e.Task {
		case taskConvertToText:
			inputStruct := ConvertToTextInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			outputStruct, err := convertToText(inputStruct)
			if err != nil {
				return nil, err
			}
			output, err := base.ConvertToStructpb(outputStruct)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, output)
		case taskSplitByToken:
			inputStruct := SplitByTokenInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			outputStruct, err := splitTextIntoChunks(inputStruct)
			if err != nil {
				return nil, err
			}
			output, err := base.ConvertToStructpb(outputStruct)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, output)
		default:
			return nil, fmt.Errorf("not supported task: %s", e.Task)
		}
	}
	return outputs, nil
}
