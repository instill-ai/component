//go:generate compogen readme ./config ./README.mdx
package audio

import (
	"context"
	_ "embed" // embed
	"fmt"
	"sync"

	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	taskChunkAudios string = "TASK_CHUNK_AUDIOS"
	taskSliceAudio  string = "TASK_SLICE_AUDIO"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	once      sync.Once
	comp      *component
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

// CreateExecution initializes a connector executor that can be used in a
// pipeline trigger.
func (c *component) CreateExecution(x base.ComponentExecution) (base.IExecution, error) {
	e := &execution{ComponentExecution: x}

	switch x.Task {
	case taskChunkAudios:
		e.execute = chunkAudios
	case taskSliceAudio:
		e.execute = sliceAudio
	default:
		return nil, fmt.Errorf(x.Task + " task is not supported.")
	}

	return e, nil
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
