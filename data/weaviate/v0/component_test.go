package weaviate

import (
	"context"
	"encoding/json"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/batch"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/data"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/filters"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/schema"
	"github.com/weaviate/weaviate/entities/models"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type MockWeaviateBatchAPIDeleterClient struct{}
type MockWeaviateBatchAPIBatcherClient struct{}
type MockWeaviateDataAPICreatorClient struct{}
type MockWeaviateGraphQLAPIGetClient struct{}
type MockWeaviateSchemaAPIDeleterClient struct{}
type MockWeaviateSchemaAPIClassGetterClient struct{}
type MockWeaviateGraphQLNearVectorArgumentBuilder struct{}

func (ob *MockWeaviateBatchAPIDeleterClient) WithClassName(className string) *batch.ObjectsBatchDeleter {
	return nil
}
func (ob *MockWeaviateBatchAPIDeleterClient) WithWhere(whereFilter *filters.WhereBuilder) *batch.ObjectsBatchDeleter {
	return nil
}
func (ob *MockWeaviateBatchAPIDeleterClient) Do(ctx context.Context) (*models.BatchDeleteResponse, error) {
	result := models.BatchDeleteResponseResults{
		Successful: 1,
	}

	return &models.BatchDeleteResponse{
		Results: &result,
	}, nil
}

func (ob *MockWeaviateBatchAPIBatcherClient) WithClassName(className string) *batch.ObjectsBatcher {
	return nil
}
func (ob *MockWeaviateBatchAPIBatcherClient) WithObjects(object ...*models.Object) *batch.ObjectsBatcher {
	return nil
}
func (ob *MockWeaviateBatchAPIBatcherClient) Do(ctx context.Context) ([]models.ObjectsGetResponse, error) {
	stringStatus := "SUCCESS"
	result := models.ObjectsGetResponseAO2Result{
		Status: &stringStatus,
	}

	return []models.ObjectsGetResponse{
		{
			Result: &result,
		},
	}, nil
}

func (ob *MockWeaviateSchemaAPIClassGetterClient) WithClassName(className string) *schema.ClassGetter {
	return nil
}
func (ob *MockWeaviateSchemaAPIClassGetterClient) Do(ctx context.Context) (*models.Class, error) {
	return &models.Class{
		Class: "test_coll",
	}, nil
}

func (ob *MockWeaviateSchemaAPIDeleterClient) WithClassName(className string) *schema.ClassDeleter {
	return nil
}
func (ob *MockWeaviateSchemaAPIDeleterClient) Do(ctx context.Context) error {
	return nil
}

func (ob *MockWeaviateDataAPICreatorClient) WithClassName(name string) *data.Creator {
	return nil
}
func (ob *MockWeaviateDataAPICreatorClient) WithProperties(propertySchema models.PropertySchema) *data.Creator {
	return nil
}
func (ob *MockWeaviateDataAPICreatorClient) WithVector(vector []float32) *data.Creator {
	return nil
}
func (ob *MockWeaviateDataAPICreatorClient) Do(ctx context.Context) (*data.ObjectWrapper, error) {
	return &data.ObjectWrapper{
		Object: &models.Object{
			Class: "test_coll",
			ID:    "test_id",
		},
	}, nil
}

func (ob *MockWeaviateGraphQLAPIGetClient) WithClassName(name string) *graphql.GetBuilder {
	return nil
}
func (ob *MockWeaviateGraphQLAPIGetClient) WithFields(fields ...graphql.Field) *graphql.GetBuilder {
	return nil
}
func (ob *MockWeaviateGraphQLAPIGetClient) WithLimit(limit int) *graphql.GetBuilder {
	return nil
}
func (ob *MockWeaviateGraphQLAPIGetClient) WithNearVector(nearVector *graphql.NearVectorArgumentBuilder) *graphql.GetBuilder {
	return nil
}
func (ob *MockWeaviateGraphQLAPIGetClient) WithTenant(tenant string) *graphql.GetBuilder {
	return nil
}
func (ob *MockWeaviateGraphQLAPIGetClient) WithWhere(where *filters.WhereBuilder) *graphql.GetBuilder {
	return nil
}
func (ob *MockWeaviateGraphQLAPIGetClient) Do(ctx context.Context) (*models.GraphQLResponse, error) {
	return &models.GraphQLResponse{
		Data: map[string]models.JSONObject{},
	}, nil
}

func (ob *MockWeaviateGraphQLNearVectorArgumentBuilder) WithVector(vector []float32) *graphql.NearVectorArgumentBuilder {
	return nil
}

func TestComponent_ExecuteInsertTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    InsertInput
		wantResp InsertOutput
		wantErr  string
	}{
		{
			name: "ok to insert",
			input: InsertInput{
				CollectionName: "test_coll",
				Vector:         []float32{0.1, 0.2},
				Metadata:       map[string]any{"name": "test"},
			},
			wantResp: InsertOutput{
				Status: "Successfully inserted 1 object",
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"url":     "mock-url",
				"api-key": "mock-api-key",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskInsert},
				client: &WeaviateClient{
					dataAPICreatorClient: &MockWeaviateDataAPICreatorClient{},
				},
			}
			e.execute = e.insert
			exec := &base.ExecutionWrapper{Execution: e}

			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			got, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})

			if tc.wantErr != "" {
				c.Assert(err, qt.ErrorMatches, tc.wantErr)
				return
			}

			wantJSON, err := json.Marshal(tc.wantResp)
			c.Assert(err, qt.IsNil)
			c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
		})
	}
}

func TestComponent_ExecuteDeleteTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    DeleteInput
		wantResp DeleteOutput
		wantErr  string
	}{
		{
			name: "ok to delete",
			input: DeleteInput{
				CollectionName: "test_coll",
				Filter:         map[string]any{"path": "text", "operator": "Equal", "valueText": "test"},
			},
			wantResp: DeleteOutput{
				Status: "Successfully deleted 1 documents",
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"url":     "mock-url",
				"api-key": "mock-api-key",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskDelete},
				client: &WeaviateClient{
					batchAPIDeleterClient:            &MockWeaviateBatchAPIDeleterClient{},
					graphQLNearVectorArgumentBuilder: &MockWeaviateGraphQLNearVectorArgumentBuilder{},
					graphQLAPIGetClient:              &MockWeaviateGraphQLAPIGetClient{},
				},
			}
			e.execute = e.delete
			exec := &base.ExecutionWrapper{Execution: e}

			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			got, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})

			if tc.wantErr != "" {
				c.Assert(err, qt.ErrorMatches, tc.wantErr)
				return
			}

			wantJSON, err := json.Marshal(tc.wantResp)
			c.Assert(err, qt.IsNil)
			c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
		})
	}
}

func TestComponent_ExecuteDeleteCollectionTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    DeleteCollectionInput
		wantResp DeleteCollectionOutput
		wantErr  string
	}{
		{
			name: "ok to delete collection",
			input: DeleteCollectionInput{
				CollectionName: "test_coll",
			},
			wantResp: DeleteCollectionOutput{
				Status: "Successfully dropped 1 collection",
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"url":     "mock-url",
				"api-key": "mock-api-key",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskDeleteCollection},
				client: &WeaviateClient{
					schemaAPIDeleterClient: &MockWeaviateSchemaAPIDeleterClient{},
				},
			}
			e.execute = e.deleteCollection
			exec := &base.ExecutionWrapper{Execution: e}

			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			got, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})

			if tc.wantErr != "" {
				c.Assert(err, qt.ErrorMatches, tc.wantErr)
				return
			}

			wantJSON, err := json.Marshal(tc.wantResp)
			c.Assert(err, qt.IsNil)
			c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
		})
	}
}

func TestComponent_ExecuteVectorSearchTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    VectorSearchInput
		wantResp VectorSearchOutput
		wantErr  string
	}{
		{
			name: "ok to vector search",
			input: VectorSearchInput{
				CollectionName: "test_coll",
				Vector:         []float32{0.1, 0.2},
				Limit:          1,
				Filter: map[string]any{
					"path":     "age",
					"operator": "Equal",
					"valueInt": 20,
				},
			},
			wantResp: VectorSearchOutput{
				Status: "Successfully found 1 documents",
				Result: Result{
					Vectors:  [][]float32{{0.1, 0.2}},
					Metadata: []map[string]any{{"name": "test"}},
					Objects:  []map[string]any{{"name": "test", "_additional": map[string]any{"vector": []float32{0.1, 0.2}}}},
				},
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"url":     "mock-url",
				"api-key": "mock-api-key",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskVectorSearch},
				client: &WeaviateClient{
					graphQLAPIGetClient:              &MockWeaviateGraphQLAPIGetClient{},
					schemaAPIClassGetterClient:       &MockWeaviateSchemaAPIClassGetterClient{},
					graphQLNearVectorArgumentBuilder: &MockWeaviateGraphQLNearVectorArgumentBuilder{},
				},
			}
			e.execute = e.vectorSearch
			exec := &base.ExecutionWrapper{Execution: e}

			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			got, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})

			if tc.wantErr != "" {
				c.Assert(err, qt.ErrorMatches, tc.wantErr)
				return
			}

			wantJSON, err := json.Marshal(tc.wantResp)
			c.Assert(err, qt.IsNil)
			c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
		})
	}
}

func TestComponent_ExecuteBatchInsertTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    BatchInsertInput
		wantResp BatchInsertOutput
		wantErr  string
	}{
		{
			name: "ok to insert many",
			input: BatchInsertInput{
				ArrayMetadata: []map[string]any{
					{"name": "test1", "email": "test1@example.com"},
					{"name": "test2", "email": "test2@example.com"},
				},
				ArrayVector: [][]float32{
					{0.1, 0.2},
					{0.3, 0.4},
				},
				CollectionName: "test_coll",
			},
			wantResp: BatchInsertOutput{
				Status: "Successfully batch inserted 2 objects",
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"url":     "mock-url",
				"api-key": "mock-api-key",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskBatchInsert},
				client: &WeaviateClient{
					batchAPIBatcherClient: &MockWeaviateBatchAPIBatcherClient{},
				},
			}
			e.execute = e.batchInsert
			exec := &base.ExecutionWrapper{Execution: e}

			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			got, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})

			if tc.wantErr != "" {
				c.Assert(err, qt.ErrorMatches, tc.wantErr)
				return
			}

			wantJSON, err := json.Marshal(tc.wantResp)
			c.Assert(err, qt.IsNil)
			c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
		})
	}
}
