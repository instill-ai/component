//go:generate compogen readme ./config ./README.mdx
package weaviate

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/x/errmsg"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	TaskVectorSearch = "TASK_VECTOR_SEARCH"
	TaskUpsert       = "TASK_UPSERT"
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

// dbClient for task DropDatabase
// collectionClient for other than task DropDatabase
type execution struct {
	base.ComponentExecution

	execute func(context.Context, *structpb.Struct) (*structpb.Struct, error)
	client  *weaviate.Client
}

// Init returns an implementation of IConnector that interacts with MongoDB.
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

func (c *component) CreateExecution(sysVars map[string]any, setup *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	e := &execution{
		ComponentExecution: base.ComponentExecution{Component: c, SystemVariables: sysVars, Setup: setup, Task: task},
		client:             newClient(setup),
	}

	switch task {
	case TaskVectorSearch:
		e.execute = e.vectorSearch
	case TaskUpsert:
		e.execute = e.upsert
	default:
		return nil, errmsg.AddMessage(
			fmt.Errorf("not supported task: %s", task),
			fmt.Sprintf("%s task is not supported.", task),
		)
	}

	return &base.ExecutionWrapper{Execution: e}, nil
}

type Destination struct {
	DatabaseName   string `json:"database-name"`
	CollectionName string `json:"collection-name"`
}

// dbClient wont be nil on component test (use mock dbClient)
// collectionClient wont be nil on component test (use mock collectionClient)
// collectionClient will be nil on task DropDatabase
func (e *execution) Execute(ctx context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := make([]*structpb.Struct, len(inputs))

	for i, input := range inputs {
		var inputStruct Destination
		err := base.ConvertFromStructpb(input, &inputStruct)
		if err != nil {
			return nil, err
		}

		output, err := e.execute(ctx, input)
		if err != nil {
			return nil, err
		}

		outputs[i] = output
	}

	return outputs, nil
}
