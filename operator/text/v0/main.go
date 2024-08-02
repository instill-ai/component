//go:generate compogen readme ./config ./README.mdx --extraContents TASK_CHUNK_TEXT=.compogen/extra-chunk-text.mdx
package text

import (
	"context"
	"fmt"
	"sync"

	_ "embed" // embed

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
)

const (
	taskChunkText string = "TASK_CHUNK_TEXT"
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
	execute func(*structpb.Struct) (*structpb.Struct, error)
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

// CreateExecution initializes a connector executor that can be used in a
// pipeline trigger.
func (c *component) CreateExecution(x base.ComponentExecution) (base.IExecution, error) {
	e := &execution{ComponentExecution: x}

	switch x.Task {
	case taskChunkText:
		e.execute = chunkText
	default:
		return nil, fmt.Errorf(x.Task + " task is not supported.")
	}

	return e, nil
}

// Execute executes the derived execution
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
