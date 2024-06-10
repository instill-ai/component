package anthropic

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
)

const (
	taskGreet = "TASK_GREET"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte

	once sync.Once
	comp *component
)

type component struct {
	base.Component
}

// Init returns an implementation of IComponent that implements the greeting
// task.
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

type execution struct {
	base.ComponentExecution
	execute func(*structpb.Struct) (*structpb.Struct, error)
}

func (c *component) CreateExecution(sysVars map[string]any, setup *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	e := &execution{
		ComponentExecution: base.ComponentExecution{Component: c, SystemVariables: sysVars, Task: task},
	}

	// A simple if statement would be enough in a component with a single task.
	// If the number of task grows, here is where the execution task would be
	// selected.
	switch task {
	case taskGreet:
		e.execute = e.greet
	default:
		return nil, fmt.Errorf("unsupported task")
	}
	return &base.ExecutionWrapper{Execution: e}, nil
}

func (e *execution) Execute(_ context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := make([]*structpb.Struct, len(inputs))

	// An execution  might take several inputs. One result will be returned for
	// each one of them, containing the execution output for that set of
	// parameters.
	for i, input := range inputs {
		output, err := e.execute(input)
		if err != nil {
			return nil, err
		}

		outputs[i] = output
	}

	return outputs, nil
}

func (e *execution) greet(in *structpb.Struct) (*structpb.Struct, error) {
	out := new(structpb.Struct)

	target := in.Fields["target"].GetStringValue()
	greeting := "Hello, " + target + "!"

	out.Fields = map[string]*structpb.Value{
		"greeting": structpb.NewStringValue(greeting),
	}

	return out, nil
}
