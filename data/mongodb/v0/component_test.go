package mongodb

import (
	"context"
	"encoding/json"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type MockMongoClient struct{}

func (m *MockMongoClient) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	mockResult := &mongo.InsertOneResult{
		InsertedID: "mockID",
	}
	return mockResult, nil
}

func (m *MockMongoClient) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	mockDocs := []bson.M{
		{"_id": "mockID1", "name": "John Doe", "email": "john@example.com"},
		{"_id": "mockID2", "name": "Jane Smith", "email": "jane@example.com"},
	}
	var docs []interface{}
	for _, doc := range mockDocs {
		docs = append(docs, doc)
	}
	mockCursor, err := mongo.NewCursorFromDocuments(docs, nil, nil)
	if err != nil {
		return nil, err
	}
	return mockCursor, nil
}

func (m *MockMongoClient) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	mockResult := &mongo.UpdateResult{
		MatchedCount:  1,
		ModifiedCount: 1,
	}
	return mockResult, nil
}

func (m *MockMongoClient) DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	mockResult := &mongo.DeleteResult{
		DeletedCount: 1,
	}
	return mockResult, nil
}

func (m *MockMongoClient) Drop(ctx context.Context) error {
	return nil
}

func (m *MockMongoClient) SearchIndexes() mongo.SearchIndexView {
	return mongo.SearchIndexView{}
}

func (m *MockMongoClient) CreateOne(ctx context.Context, model mongo.SearchIndexModel, opts ...*options.CreateSearchIndexesOptions) (string, error) {
	return "mockIndex", nil
}

func (m *MockMongoClient) DropOne(ctx context.Context, name string, _ ...*options.DropSearchIndexOptions) error {
	return nil
}

func (m *MockMongoClient) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	mockDocs := []bson.M{
		{"vector": []float64{0.1, 0.2}, "name": "test"},
	}

	var docs []interface{}
	for _, doc := range mockDocs {
		docs = append(docs, doc)
	}
	mockCursor, err := mongo.NewCursorFromDocuments(docs, nil, nil)
	if err != nil {
		return nil, err
	}
	return mockCursor, nil
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

				Data: map[string]any{"name": "test", "email": "test@example.com"},
			},
			wantResp: InsertOutput{
				Status: "Successfully inserted document",
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"name":            "test",
				"collection-name": "test_coll",
				"uri":             "mongodb://localhost:27017",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskInsert},
				client: &MongoClient{
					collectionClient: &MockMongoClient{},
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

func TestComponent_ExecuteFindTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    FindInput
		wantResp FindOutput
		wantErr  string
	}{
		{
			name: "ok to find",
			input: FindInput{
				Filter: map[string]any{"name": "test"},
				Limit:  0,
			},
			wantResp: FindOutput{
				Status: "Successfully found documents",
				Documents: []map[string]any{
					{"_id": "mockID1", "name": "John Doe", "email": "john@example.com"},
					{"_id": "mockID2", "name": "Jane Smith", "email": "jane@example.com"},
				},
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"name":            "test",
				"collection-name": "test_coll",
				"uri":             "mongodb://localhost:27017",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskFind},
				client: &MongoClient{
					collectionClient: &MockMongoClient{},
				},
			}
			e.execute = e.find
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

func TestComponent_ExecuteUpdateTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    UpdateInput
		wantResp UpdateOutput
		wantErr  string
	}{
		{
			name: "ok to update",
			input: UpdateInput{
				Filter:     map[string]any{"name": "test1"},
				UpdateData: map[string]any{"name": "test2"},
			},
			wantResp: UpdateOutput{
				Status: "Successfully updated documents",
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"name":            "test",
				"collection-name": "test_coll",
				"uri":             "mongodb://localhost:27017",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskUpdate},
				client: &MongoClient{
					collectionClient: &MockMongoClient{},
				},
			}
			e.execute = e.update
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
				Filter: map[string]any{"name": "test"},
			},
			wantResp: DeleteOutput{
				Status: "Successfully deleted documents",
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"name":            "test",
				"collection-name": "test_coll",
				"uri":             "mongodb://localhost:27017",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskDelete},
				client: &MongoClient{
					collectionClient: &MockMongoClient{},
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
		input    DropCollectionInput
		wantResp DropCollectionOutput
		wantErr  string
	}{
		{
			name: "ok to delete collection",
			input: DropCollectionInput{
				CollectionName: "test_coll",
			},
			wantResp: DropCollectionOutput{
				Status: "Successfully dropped collection",
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"name":            "test",
				"collection-name": tc.input.CollectionName,
				"uri":             "mongodb://localhost:27017",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskDropCollection},
				client: &MongoClient{
					collectionClient: &MockMongoClient{},
				},
			}
			e.execute = e.dropCollection
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

func TestComponent_ExecuteDeleteDatabaseTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    DropDatabaseInput
		wantResp DropDatabaseOutput
		wantErr  string
	}{
		{
			name: "ok to delete database",
			input: DropDatabaseInput{
				DatabaseName: "test_db",
			},
			wantResp: DropDatabaseOutput{
				Status: "Successfully dropped database",
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"name":            tc.input.DatabaseName,
				"collection-name": "test_coll",
				"uri":             "mongodb://localhost:27017",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskDropDatabase},
				client: &MongoClient{
					databaseClient: &MockMongoClient{},
				},
			}
			e.execute = e.dropDatabase
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

func TestComponent_ExecuteCreateSearchIndexTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    CreateSearchIndexInput
		wantResp CreateSearchIndexOutput
		wantErr  string
	}{
		{
			name: "ok to create search index",
			input: CreateSearchIndexInput{
				Syntax: map[string]any{
					"Fields": []map[string]any{{
						"type":          "vector",
						"numDimensions": 2,
						"path":          "mock_collection",
						"similarity":    "cosine",
					},
					},
				},
			},
			wantResp: CreateSearchIndexOutput{
				Status: "Successfully created search index",
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"name":            "test",
				"collection-name": "test_coll",
				"uri":             "mongodb://localhost:27017",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskCreateSearchIndex},
				client: &MongoClient{
					collectionClient:  &MockMongoClient{},
					searchIndexClient: &MockMongoClient{},
				},
			}
			e.execute = e.createSearchIndex
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

func TestComponent_ExecuteDropSearchIndexTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    DropSearchIndexInput
		wantResp DropSearchIndexOutput
		wantErr  string
	}{
		{
			name: "ok to drop search index",
			input: DropSearchIndexInput{
				IndexName: "index_name",
			},
			wantResp: DropSearchIndexOutput{
				Status: "Successfully dropped search index",
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"name":            "test",
				"collection-name": "test_coll",
				"uri":             "mongodb://localhost:27017",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskDropSearchIndex},
				client: &MongoClient{
					collectionClient:  &MockMongoClient{},
					searchIndexClient: &MockMongoClient{},
				},
			}
			e.execute = e.dropSearchIndex
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
				Exact:         false,
				IndexName:     "index_name",
				Limit:         0,
				NumCandidates: 10,
				Path:          "vector",
				QueryVector:   []float64{0.1, 0.2},
				Filter:        map[string]any{"name": "test"},
			},
			wantResp: VectorSearchOutput{
				Status: "Successfully found documents",
				Documents: []map[string]any{
					{"vector": []float64{0.1, 0.2}, "name": "test"},
				},
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"name":            "test",
				"collection-name": "test_coll",
				"uri":             "mongodb://localhost:27017",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskVectorSearch},
				client: &MongoClient{
					collectionClient: &MockMongoClient{},
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
