package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func MockESSearch(wantResp SearchOutput) *esapi.Response {
	var Hits []Hit
	documentsBytes, _ := json.Marshal(wantResp.Documents)
	_ = json.Unmarshal(documentsBytes, &Hits)

	resp := SearchResponse{
		Took:     1,
		TimedOut: false,
		Shards: struct {
			Total      int `json:"total"`
			Successful int `json:"successful"`
			Skipped    int `json:"skipped"`
			Failed     int `json:"failed"`
		}{
			Total:      1,
			Successful: 1,
			Skipped:    0,
			Failed:     0,
		},
		Hits: struct {
			Total struct {
				Value    int    `json:"value"`
				Relation string `json:"relation"`
			} `json:"total"`
			MaxScore float64 `json:"max_score"`
			Hits     []Hit   `json:"hits"`
		}{
			Total: struct {
				Value    int    `json:"value"`
				Relation string `json:"relation"`
			}{
				Value:    len(wantResp.Documents),
				Relation: "eq",
			},
			MaxScore: 2,
			Hits:     Hits,
		},
	}

	b, _ := json.Marshal(resp)
	return &esapi.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(b)),
		Header:     make(map[string][]string),
	}
}

func MockESIndex(wantResp IndexOutput) *esapi.Response {
	resp := map[string]string{"status": wantResp.Status}
	b, _ := json.Marshal(resp)
	return &esapi.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(b)),
		Header:     make(map[string][]string),
	}
}

func MockESUpdate(wantResp UpdateOutput) *esapi.Response {
	resp := map[string]string{"status": wantResp.Status}
	b, _ := json.Marshal(resp)
	return &esapi.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(b)),
		Header:     make(map[string][]string),
	}
}

func MockESDelete(wantResp DeleteOutput) *esapi.Response {
	resp := map[string]string{"status": wantResp.Status}
	b, _ := json.Marshal(resp)
	return &esapi.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(b)),
		Header:     make(map[string][]string),
	}
}

func MockESDeleteIndex(wantResp DeleteIndexOutput) *esapi.Response {
	resp := map[string]string{"status": wantResp.Status}
	b, _ := json.Marshal(resp)
	return &esapi.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(b)),
		Header:     make(map[string][]string),
	}
}

func TestComponent_ExecuteSearchTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    SearchInput
		wantResp SearchOutput
		wantErr  string
	}{
		{
			name: "ok to search",
			input: SearchInput{
				Mode:      "Hits",
				IndexName: "index_name",
				Filter:    map[string]any{"city": "New York"},
				Size:      0,
			},
			wantResp: SearchOutput{
				Status: "Successfully searched document",
				Documents: []map[string]any{
					{
						"_index":  "index_name",
						"_id":     "mockID1",
						"_score":  0,
						"_source": map[string]any{"name": "John Doe", "email": "john@example.com", "city": "New York"},
					},
					{
						"_index":  "index_name",
						"_id":     "mockID2",
						"_score":  0,
						"_source": map[string]any{"name": "Jane Smith", "email": "jane@example.com", "city": "New York"},
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"api-key":  "mock-api-key",
				"cloud-id": "mock-cloud-id",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskSearch},
				searchClient: func(o ...func(*esapi.SearchRequest)) (*esapi.Response, error) {
					return MockESSearch(tc.wantResp), nil
				},
			}

			e.execute = e.search
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

func TestComponent_ExecuteIndexTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    IndexInput
		wantResp IndexOutput
		wantErr  string
	}{
		{
			name: "ok to index",
			input: IndexInput{
				IndexName: "index_name",
				Data:      map[string]any{"name": "John Doe", "email": "john@example.com"},
			},
			wantResp: IndexOutput{
				Status: "Successfully indexed document",
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"api-key":  "mock-api-key",
				"cloud-id": "mock-cloud-id",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskIndex},
				indexClient: func(index string, body io.Reader, o ...func(*esapi.IndexRequest)) (*esapi.Response, error) {
					return MockESIndex(tc.wantResp), nil
				},
			}

			e.execute = e.index
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
				IndexName: "index_name",
				Filter:    map[string]any{"name": "John Doe", "city": "New York"},
				Update:    map[string]any{"name": "Pablo Vereira", "city": "Los Angeles"},
			},
			wantResp: UpdateOutput{
				Status: "Successfully updated document",
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"api-key":  "mock-api-key",
				"cloud-id": "mock-cloud-id",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskUpdate},
				updateClient: func(index []string, o ...func(*esapi.UpdateByQueryRequest)) (*esapi.Response, error) {
					return MockESUpdate(tc.wantResp), nil
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
				IndexName: "index_name",
				Filter:    map[string]any{"name": "John Doe", "city": "New York}"},
			},
			wantResp: DeleteOutput{
				Status: "Successfully deleted document",
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"api-key":  "mock-api-key",
				"cloud-id": "mock-cloud-id",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskDelete},
				deleteClient: func(index []string, body io.Reader, o ...func(*esapi.DeleteByQueryRequest)) (*esapi.Response, error) {
					return MockESDelete(tc.wantResp), nil
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

func TestComponent_ExecuteDeleteIndexTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    DeleteIndexInput
		wantResp DeleteIndexOutput
		wantErr  string
	}{
		{
			name: "ok to delete index",
			input: DeleteIndexInput{
				IndexName: "index_name",
			},
			wantResp: DeleteIndexOutput{
				Status: "Successfully deleted index",
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"api-key":  "mock-api-key",
				"cloud-id": "mock-cloud-id",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskDeleteIndex},
				deleteIndexClient: func(index []string, o ...func(*esapi.IndicesDeleteRequest)) (*esapi.Response, error) {
					return MockESDeleteIndex(tc.wantResp), nil
				},
			}

			e.execute = e.deleteIndex
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
