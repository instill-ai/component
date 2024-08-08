//go:generate compogen readme ./config ./README.mdx
package qdrant

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/util/httpclient"
	"github.com/instill-ai/x/errmsg"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	TaskVectorSearch     = "TASK_VECTOR_SEARCH"
	TaskDelete           = "TASK_DELETE"
	TaskBatchUpsert      = "TASK_BATCH_UPSERT"
	TaskUpsert           = "TASK_UPSERT"
	TaskCreateCollection = "TASK_CREATE_COLLECTION"
	TaskDeleteCollection = "TASK_DELETE_COLLECTION"
)

//go:embed config/definition.json
var definitionJSON []byte

//go:embed config/setup.json
var setupJSON []byte

//go:embed config/tasks.json
var tasksJSON []byte

var once sync.Once
var comp *component

type component struct {
	base.Component
}

type execution struct {
	base.ComponentExecution

	execute func(*structpb.Struct) (*structpb.Struct, error)
	client  *httpclient.Client
}

func Init(bc base.Component) *component {
	once.Do(func() {
		comp = &component{Component: bc}
		err := comp.LoadDefinition(definitionJSON, setupJSON, tasksJSON, nil)
		if err != nil {
			panic(err)
		}
	})

	return comp
}

func (c *component) CreateExecution(x base.ComponentExecution) (base.IExecution, error) {
	e := &execution{
		ComponentExecution: x,
		client:             newClient(x.Setup, x.GetLogger()),
	}

	switch x.Task {
	case TaskDelete:
		e.execute = e.delete
	case TaskBatchUpsert:
		e.execute = e.batchUpsert
	case TaskUpsert:
		e.execute = e.upsert
	case TaskCreateCollection:
		e.execute = e.createCollection
	case TaskDeleteCollection:
		e.execute = e.deleteCollection
	case TaskVectorSearch:
		e.execute = e.vectorSearch
	default:
		return nil, errmsg.AddMessage(
			fmt.Errorf("not supported task: %s", x.Task),
			fmt.Sprintf("%s task is not supported.", x.Task),
		)
	}

	return e, nil
}

func (e *execution) Execute(ctx context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
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