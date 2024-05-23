//go:generate compogen readme --operator ./config ./README.mdx
package pdf

import (
	"context"
	"fmt"
	"os/exec"
	"sync"

	_ "embed"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
)

const (
	taskConvertToMarkdown string = "TASK_CONVERT_TO_MARKDOWN"
	scriptPath            string = "/component/pkg/operator/pdf/v0/python/pdf_transformer.py"
	pythonInterpreter     string = "/opt/venv/bin/python"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	once      sync.Once
	op        *operator
)

type operator struct {
	base.Operator
}

type execution struct {
	base.OperatorExecution
}

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

func (e *execution) Execute(_ context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		switch e.Task {
		case taskConvertToMarkdown:
			inputStruct := convertPDFToMarkdownInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}

			cmd := exec.Command(pythonInterpreter, scriptPath)

			outputStruct, err := convertPDFToMarkdown(inputStruct, cmd)
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
