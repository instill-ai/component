//go:generate compogen readme ./config ./README.mdx
package artifact

import (
	"context"
	"fmt"
	"sync"

	_ "embed"

	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	taskUploadFiles       string = "TASK_UPLOAD_FILES"
	taskGetFilesMetadata  string = "TASK_GET_FILES_METADATA"
	taskGetChunksMetadata string = "TASK_GET_CHUNKS_METADATA"
	taskGetFileInMarkdown string = "TASK_GET_FILE_IN_MARKDOWN"
	taskMatchFileStatus   string = "TASK_MATCH_FILE_STATUS"
	taskSearchChunks      string = "TASK_SEARCH_CHUNKS"
	taskQuery             string = "TASK_QUERY"
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

func (c *component) CreateExecution(x base.ComponentExecution) (base.IExecution, error) {
	e := &execution{ComponentExecution: x}

	switch x.Task {
	case taskUploadFiles:
		e.execute = uploadFiles
	case taskGetFilesMetadata:
		e.execute = getFilesMetadata
	case taskGetChunksMetadata:
		e.execute = getChunksMetadata
	case taskGetFileInMarkdown:
		e.execute = getFileInMarkdown
	case taskMatchFileStatus:
		e.execute = matchFileStatus
	case taskSearchChunks:
		e.execute = searchChunks
	case taskQuery:
		e.execute = query
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
