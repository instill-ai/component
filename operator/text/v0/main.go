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
	return &execution{ComponentExecution: x}, nil
}

// Execute executes the derived execution
func (e *execution) Execute(ctx context.Context, in base.InputReader, out base.OutputWriter) error {
	inputs, err := in.Read(ctx)
	if err != nil {
		return err
	}
	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		switch e.Task {
		case taskChunkText:
			inputStruct := ChunkTextInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return err
			}

			var outputStruct ChunkTextOutput
			if inputStruct.Strategy.Setting.ChunkMethod == "Markdown" {
				outputStruct, err = markdownChunkText(inputStruct)
			} else {
				outputStruct, err = chunkText(inputStruct)
			}

			if err != nil {
				return err
			}
			output, err := base.ConvertToStructpb(outputStruct)
			if err != nil {
				return err
			}
			outputs = append(outputs, output)
		default:
			return fmt.Errorf("not supported task: %s", e.Task)
		}
	}
	return out.Write(ctx, outputs)
}
