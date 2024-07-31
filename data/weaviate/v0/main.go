//go:generate compogen readme ./config ./README.mdx
package weaviate

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/x/errmsg"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/batch"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/data"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/filters"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/schema"
	"github.com/weaviate/weaviate/entities/models"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	TaskVectorSearch     = "TASK_VECTOR_SEARCH"
	TaskInsert           = "TASK_INSERT"
	TaskDelete           = "TASK_DELETE"
	TaskBatchInsert      = "TASK_BATCH_INSERT"
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

	execute func(context.Context, *structpb.Struct) (*structpb.Struct, error)
	client  *WeaviateClient
}

type WeaviateDataAPICreatorClient interface {
	WithClassName(name string) *data.Creator
	WithVector(vector []float32) *data.Creator
	WithProperties(propertySchema models.PropertySchema) *data.Creator
	Do(ctx context.Context) (*data.ObjectWrapper, error)
}

type WeaviateGraphQLAPIGetClient interface {
	WithClassName(name string) *graphql.GetBuilder
	WithWhere(where *filters.WhereBuilder) *graphql.GetBuilder
	WithNearVector(nearVector *graphql.NearVectorArgumentBuilder) *graphql.GetBuilder
	WithLimit(limit int) *graphql.GetBuilder
	WithFields(fields ...graphql.Field) *graphql.GetBuilder
	WithTenant(tenant string) *graphql.GetBuilder
	Do(ctx context.Context) (*models.GraphQLResponse, error)
}

type WeaviateBatchAPIDeleterClient interface {
	WithClassName(className string) *batch.ObjectsBatchDeleter
	WithWhere(whereFilter *filters.WhereBuilder) *batch.ObjectsBatchDeleter
	Do(ctx context.Context) (*models.BatchDeleteResponse, error)
}

type WeaviateBatchAPIBatcherClient interface {
	WithObjects(object ...*models.Object) *batch.ObjectsBatcher
	Do(ctx context.Context) ([]models.ObjectsGetResponse, error)
}

type WeaviateSchemaAPIDeleterClient interface {
	WithClassName(className string) *schema.ClassDeleter
	Do(ctx context.Context) error
}

type WeaviateSchemaAPIClassGetterClient interface {
	WithClassName(className string) *schema.ClassGetter
	Do(ctx context.Context) (*models.Class, error)
}

type WeaviateGraphQLNearVectorArgumentBuilder interface {
	WithVector(vector []float32) *graphql.NearVectorArgumentBuilder
}

type WeaviateClient struct {
	dataAPICreatorClient             WeaviateDataAPICreatorClient
	graphQLAPIGetClient              WeaviateGraphQLAPIGetClient
	batchAPIDeleterClient            WeaviateBatchAPIDeleterClient
	batchAPIBatcherClient            WeaviateBatchAPIBatcherClient
	schemaAPIDeleterClient           WeaviateSchemaAPIDeleterClient
	schemaAPIClassGetterClient       WeaviateSchemaAPIClassGetterClient
	graphQLNearVectorArgumentBuilder WeaviateGraphQLNearVectorArgumentBuilder
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

func (c *component) CreateExecution(sysVars map[string]any, setup *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	e := &execution{
		ComponentExecution: base.ComponentExecution{Component: c, SystemVariables: sysVars, Setup: setup, Task: task},
		client:             newClient(setup),
	}

	switch task {
	case TaskVectorSearch:
		e.execute = e.vectorSearch
	case TaskInsert:
		e.execute = e.insert
	case TaskDelete:
		e.execute = e.delete
	case TaskBatchInsert:
		e.execute = e.batchInsert
	case TaskDeleteCollection:
		e.execute = e.deleteCollection
	default:
		return nil, errmsg.AddMessage(
			fmt.Errorf("not supported task: %s", task),
			fmt.Sprintf("%s task is not supported.", task),
		)
	}

	return &base.ExecutionWrapper{Execution: e}, nil
}

func (e *execution) Execute(ctx context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := make([]*structpb.Struct, len(inputs))

	for i, input := range inputs {
		output, err := e.execute(ctx, input)
		if err != nil {
			return nil, err
		}

		outputs[i] = output
	}

	return outputs, nil
}
