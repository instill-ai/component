//go:generate compogen readme ./config ./README.mdx
package document

import (
	"context"
	"fmt"
	"os/exec"
	"sync"

	_ "embed"

	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	taskConvertToMarkdown string = "TASK_CONVERT_TO_MARKDOWN"
	pythonInterpreter     string = "/opt/venv/bin/python"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	//go:embed python/transformPDFToMarkdown.py
	pythonCode string
	once       sync.Once
	comp       *component
)

type component struct {
	base.Component
}

type execution struct {
	base.ComponentExecution
}

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

			cmd := exec.Command(pythonInterpreter, "-c", pythonCode)

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
