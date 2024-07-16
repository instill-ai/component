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

func (m *MockMongoClient) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	mockDoc := bson.M{"_id": "mockID", "name": "John Doe", "email": "john.doe@example.com"}
	mockResult := mongo.NewSingleResultFromDocument(mockDoc, nil, nil)
	return mockResult
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
				collectionClient:   &MockMongoClient{},
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
				Criteria: map[string]any{"name": "test"},
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
				collectionClient:   &MockMongoClient{},
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
				Criteria: map[string]any{"name": "test1"},
				Update:   map[string]any{"name": "test2"},
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
				collectionClient:   &MockMongoClient{},
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
				Criteria: map[string]any{"name": "test"},
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
				collectionClient:   &MockMongoClient{},
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
				collectionClient:   &MockMongoClient{},
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
				dbClient:           &MockMongoClient{},
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
