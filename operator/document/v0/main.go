//go:generate compogen readme ./config ./README.mdx
package document

import (
	"context"
	"fmt"
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
	execute func(*structpb.Struct) (*structpb.Struct, error)
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
	e := &execution{
		ComponentExecution: base.ComponentExecution{Component: c, SystemVariables: sysVars, Setup: setup, Task: task},
	}

	switch task {
	case taskConvertToMarkdown:
		e.execute = e.convertDocumentToMarkdown
	default:
		return nil, fmt.Errorf(fmt.Sprintf("%s task is not supported.", task))
	}

	return &base.ExecutionWrapper{Execution: e}, nil
}

func (e *execution) Execute(_ context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := make([]*structpb.Struct, len(inputs))

	for i, input := range inputs {
		output, err := e.execute(input)
		if err != nil {
			return nil, err
		}

		outputs[i] = output
	}
	return outputs, nil
}
