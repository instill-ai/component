package chroma

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	qt "github.com/frankban/quicktest"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/mock"
	"github.com/instill-ai/component/internal/util/httpclient"
)

func TestComponent_ExecuteQueryTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	cmp := Init(bc)

	testcases := []struct {
		name     string
		input    QueryInput
		wantResp QueryOutput
		wantErr  string

		wantClientPath string
		wantClientReq  any
		clientResp     string
	}{
		{
			name: "ok to vector search",
			input: QueryInput{
				CollectionName: "mock-collection",
				Vector:         []float64{0.1, 0.2},
				NResults:       2,
			},
			wantResp: QueryOutput{
				Status: "Successfully queryed 2 items",
				Result: Result{
					Ids:      []string{"mockID1", "mockID2"},
					Items:    []map[string]any{{"distance": 1, "id": "mockID1", "name": "a", "vector": []float32{0.1, 0.2}}, {"distance": 1, "id": "mockID2", "name": "b", "vector": []float32{0.2, 0.3}}},
					Vectors:  [][]float64{{0.1, 0.2}, {0.2, 0.3}},
					Metadata: []map[string]any{{"name": "a"}, {"name": "b"}},
				},
			},
			wantClientPath: fmt.Sprintf(queryPath, "mock-collection-id"),
			wantClientReq: QueryReq{
				QueryEmbeddings: []float64{0.1, 0.2},
				NResults:        2,
			},
			clientResp: `{
				"ids": ["mockID1", "mockID2"],
				"embeddings": [[0.1, 0.2], [0.2, 0.3]],
				"metadatas": [{"name": "a"}, {"name": "b"}],
				"distances": [1, 1]
			}`,
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.Check(r.Method, qt.Equals, http.MethodPost)
				c.Check(r.URL.Path, qt.Equals, tc.wantClientPath)

				c.Check(r.Header.Get("Content-Type"), qt.Equals, httpclient.MIMETypeJSON)
				c.Check(r.Header.Get("Accept"), qt.Equals, httpclient.MIMETypeJSON)
				c.Check(r.Header.Get("Authorization"), qt.Equals, "Bearer mock-api-key")

				c.Assert(r.Body, qt.IsNotNil)
				defer r.Body.Close()

				body, err := io.ReadAll(r.Body)
				c.Assert(err, qt.IsNil)
				c.Check(body, qt.JSONEquals, tc.wantClientReq)

				w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
				fmt.Fprintln(w, tc.clientResp)
			})

			milvusServer := httptest.NewServer(h)
			c.Cleanup(milvusServer.Close)

			setup, _ := structpb.NewStruct(map[string]any{
				"api-key": "mock-api-key",
				"url":     milvusServer.URL,
			})

			exec, err := cmp.CreateExecution(base.ComponentExecution{
				Component: cmp,
				Setup:     setup,
				Task:      TaskQuery,
			})
			c.Assert(err, qt.IsNil)

			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			ir := mock.NewInputReaderMock(c)
			ow := mock.NewOutputWriterMock(c)
			ir.ReadMock.Return([]*structpb.Struct{pbIn}, nil)
			ow.WriteMock.Optional().Set(func(ctx context.Context, outputs []*structpb.Struct) (err error) {
				c.Assert(outputs, qt.HasLen, 1)
				wantJSON, err := json.Marshal(tc.wantResp)
				c.Assert(err, qt.IsNil)
				c.Check(wantJSON, qt.JSONEquals, outputs[0].AsMap())
				return nil
			})

			err = exec.Execute(ctx, ir, ow)
			c.Check(err, qt.IsNil)

		})
	}
}

func TestComponent_ExecuteDeleteTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	cmp := Init(bc)

	testcases := []struct {
		name     string
		input    DeleteInput
		wantResp DeleteOutput
		wantErr  string

		wantClientPath string
		wantClientReq  any
		clientResp     string
	}{
		{
			name: "ok to delete search",
			input: DeleteInput{
				CollectionName: "mock-collection",
				ID:             "mockID1",
			},
			wantResp: DeleteOutput{
				Status: "Successfully deleted 1 items",
			},
			wantClientPath: fmt.Sprintf(deletePath, "mock-collection-id"),
			wantClientReq: DeleteReq{
				IDs: []string{"mockID1"},
			},
			clientResp: `["mockID1"]`,
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.Check(r.Method, qt.Equals, http.MethodPost)
				c.Check(r.URL.Path, qt.Equals, tc.wantClientPath)

				c.Check(r.Header.Get("Content-Type"), qt.Equals, httpclient.MIMETypeJSON)
				c.Check(r.Header.Get("Accept"), qt.Equals, httpclient.MIMETypeJSON)
				c.Check(r.Header.Get("Authorization"), qt.Equals, "Bearer mock-api-key")

				c.Assert(r.Body, qt.IsNotNil)
				defer r.Body.Close()

				body, err := io.ReadAll(r.Body)
				c.Assert(err, qt.IsNil)
				c.Check(body, qt.JSONEquals, tc.wantClientReq)

				w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
				fmt.Fprintln(w, tc.clientResp)
			})

			milvusServer := httptest.NewServer(h)
			c.Cleanup(milvusServer.Close)

			setup, _ := structpb.NewStruct(map[string]any{
				"api-key": "mock-api-key",
				"url":     milvusServer.URL,
			})

			exec, err := cmp.CreateExecution(base.ComponentExecution{
				Component: cmp,
				Setup:     setup,
				Task:      TaskDelete,
			})

			c.Assert(err, qt.IsNil)

			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			ir := mock.NewInputReaderMock(c)
			ow := mock.NewOutputWriterMock(c)
			ir.ReadMock.Return([]*structpb.Struct{pbIn}, nil)
			ow.WriteMock.Optional().Set(func(ctx context.Context, outputs []*structpb.Struct) (err error) {
				c.Assert(outputs, qt.HasLen, 1)
				wantJSON, err := json.Marshal(tc.wantResp)
				c.Assert(err, qt.IsNil)
				c.Check(wantJSON, qt.JSONEquals, outputs[0].AsMap())
				return nil
			})

			err = exec.Execute(ctx, ir, ow)
			c.Check(err, qt.IsNil)

		})
	}
}

func TestComponent_ExecuteUpsertTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	cmp := Init(bc)

	testcases := []struct {
		name     string
		input    UpsertInput
		wantResp UpsertOutput
		wantErr  string

		wantClientPath string
		wantClientReq  any
		clientResp     string
	}{
		{
			name: "ok to upsert search",
			input: UpsertInput{
				CollectionName: "mock-collection",
				ID:             "mockID1",
				Vector:         []float64{0.1, 0.2},
				Metadata:       map[string]any{"name": "a"},
			},
			wantResp: UpsertOutput{
				Status: "Successfully upserted 1 item",
			},
			wantClientPath: fmt.Sprintf(upsertPath, "mock-collection-id"),
			wantClientReq: UpsertReq{
				Embeddings: [][]float64{{0.1, 0.2}},
				Metadatas:  []map[string]any{{"name": "a"}},
				IDs:        []string{"mockID1"},
			},
			clientResp: `null`,
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.Check(r.Method, qt.Equals, http.MethodPost)
				c.Check(r.URL.Path, qt.Equals, tc.wantClientPath)

				c.Check(r.Header.Get("Content-Type"), qt.Equals, httpclient.MIMETypeJSON)
				c.Check(r.Header.Get("Accept"), qt.Equals, httpclient.MIMETypeJSON)
				c.Check(r.Header.Get("Authorization"), qt.Equals, "Bearer mock-api-key")

				c.Assert(r.Body, qt.IsNotNil)
				defer r.Body.Close()

				body, err := io.ReadAll(r.Body)
				c.Assert(err, qt.IsNil)
				c.Check(body, qt.JSONEquals, tc.wantClientReq)

				w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
				fmt.Fprintln(w, tc.clientResp)
			})

			milvusServer := httptest.NewServer(h)
			c.Cleanup(milvusServer.Close)

			setup, _ := structpb.NewStruct(map[string]any{
				"api-key": "mock-api-key",
				"url":     milvusServer.URL,
			})

			exec, err := cmp.CreateExecution(base.ComponentExecution{
				Component: cmp,
				Setup:     setup,
				Task:      TaskUpsert,
			})

			c.Assert(err, qt.IsNil)

			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			ir := mock.NewInputReaderMock(c)
			ow := mock.NewOutputWriterMock(c)
			ir.ReadMock.Return([]*structpb.Struct{pbIn}, nil)
			ow.WriteMock.Optional().Set(func(ctx context.Context, outputs []*structpb.Struct) (err error) {
				c.Assert(outputs, qt.HasLen, 1)
				wantJSON, err := json.Marshal(tc.wantResp)
				c.Assert(err, qt.IsNil)
				c.Check(wantJSON, qt.JSONEquals, outputs[0].AsMap())
				return nil
			})

			err = exec.Execute(ctx, ir, ow)
			c.Check(err, qt.IsNil)

		})
	}
}

func TestComponent_ExecuteBatchUpsertTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	cmp := Init(bc)

	testcases := []struct {
		name     string
		input    BatchUpsertInput
		wantResp BatchUpsertOutput
		wantErr  string

		wantClientPath string
		wantClientReq  any
		clientResp     string
	}{
		{
			name: "ok to batch upsert search",
			input: BatchUpsertInput{
				CollectionName: "mock-collection",
				ArrayID:        []string{"mockID1", "mockID2"},
				ArrayVector:    [][]float64{{0.1, 0.2}, {0.2, 0.3}},
				ArrayMetadata:  []map[string]any{{"name": "a"}, {"name": "b"}},
			},
			wantResp: BatchUpsertOutput{
				Status: "Successfully batch upserted 2 items",
			},
			wantClientPath: fmt.Sprintf(upsertPath, "mock-collection-id"),
			wantClientReq: UpsertReq{
				Embeddings: [][]float64{{0.1, 0.2}, {0.2, 0.3}},
				Metadatas:  []map[string]any{{"name": "a"}, {"name": "b"}},
				IDs:        []string{"mockID1", "mockID2"},
			},
			clientResp: `null`,
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.Check(r.Method, qt.Equals, http.MethodPost)
				c.Check(r.URL.Path, qt.Equals, tc.wantClientPath)

				c.Check(r.Header.Get("Content-Type"), qt.Equals, httpclient.MIMETypeJSON)
				c.Check(r.Header.Get("Accept"), qt.Equals, httpclient.MIMETypeJSON)
				c.Check(r.Header.Get("Authorization"), qt.Equals, "Bearer mock-api-key")

				c.Assert(r.Body, qt.IsNotNil)
				defer r.Body.Close()

				body, err := io.ReadAll(r.Body)
				c.Assert(err, qt.IsNil)
				c.Check(body, qt.JSONEquals, tc.wantClientReq)

				w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
				fmt.Fprintln(w, tc.clientResp)
			})

			milvusServer := httptest.NewServer(h)
			c.Cleanup(milvusServer.Close)

			setup, _ := structpb.NewStruct(map[string]any{
				"api-key": "mock-api-key",
				"url":     milvusServer.URL,
			})

			exec, err := cmp.CreateExecution(base.ComponentExecution{
				Component: cmp,
				Setup:     setup,
				Task:      TaskBatchUpsert,
			})

			c.Assert(err, qt.IsNil)

			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			ir := mock.NewInputReaderMock(c)
			ow := mock.NewOutputWriterMock(c)
			ir.ReadMock.Return([]*structpb.Struct{pbIn}, nil)
			ow.WriteMock.Optional().Set(func(ctx context.Context, outputs []*structpb.Struct) (err error) {
				c.Assert(outputs, qt.HasLen, 1)
				wantJSON, err := json.Marshal(tc.wantResp)
				c.Assert(err, qt.IsNil)
				c.Check(wantJSON, qt.JSONEquals, outputs[0].AsMap())
				return nil
			})

			err = exec.Execute(ctx, ir, ow)
			c.Check(err, qt.IsNil)

		})
	}
}

func TestComponent_ExecuteCreateCollectionTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	cmp := Init(bc)

	testcases := []struct {
		name     string
		input    CreateCollectionInput
		wantResp CreateCollectionOutput
		wantErr  string

		wantClientPath string
		wantClientReq  any
		clientResp     string
	}{
		{
			name: "ok to create collection",
			input: CreateCollectionInput{
				CollectionName: "mock-collection",
			},
			wantResp: CreateCollectionOutput{
				Status: "Successfully created 1 collection",
			},
			wantClientPath: createCollectionPath,
			wantClientReq: CreateCollectionReq{
				Name: "mock-collection",
			},
			clientResp: `null`,
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.Check(r.Method, qt.Equals, http.MethodPost)
				c.Check(r.URL.Path, qt.Equals, tc.wantClientPath)

				c.Check(r.Header.Get("Content-Type"), qt.Equals, httpclient.MIMETypeJSON)
				c.Check(r.Header.Get("Accept"), qt.Equals, httpclient.MIMETypeJSON)
				c.Check(r.Header.Get("Authorization"), qt.Equals, "Bearer mock-api-key")

				c.Assert(r.Body, qt.IsNotNil)
				defer r.Body.Close()

				body, err := io.ReadAll(r.Body)
				c.Assert(err, qt.IsNil)
				c.Check(body, qt.JSONEquals, tc.wantClientReq)

				w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
				fmt.Fprintln(w, tc.clientResp)
			})

			milvusServer := httptest.NewServer(h)
			c.Cleanup(milvusServer.Close)

			setup, _ := structpb.NewStruct(map[string]any{
				"api-key": "mock-api-key",
				"url":     milvusServer.URL,
			})

			exec, err := cmp.CreateExecution(base.ComponentExecution{
				Component: cmp,
				Setup:     setup,
				Task:      TaskCreateCollection,
			})

			c.Assert(err, qt.IsNil)

			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			ir := mock.NewInputReaderMock(c)
			ow := mock.NewOutputWriterMock(c)
			ir.ReadMock.Return([]*structpb.Struct{pbIn}, nil)
			ow.WriteMock.Optional().Set(func(ctx context.Context, outputs []*structpb.Struct) (err error) {
				c.Assert(outputs, qt.HasLen, 1)
				wantJSON, err := json.Marshal(tc.wantResp)
				c.Assert(err, qt.IsNil)
				c.Check(wantJSON, qt.JSONEquals, outputs[0].AsMap())
				return nil
			})

			err = exec.Execute(ctx, ir, ow)
			c.Check(err, qt.IsNil)

		})
	}
}

func TestComponent_ExecuteDeleteCollectionTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	cmp := Init(bc)

	testcases := []struct {
		name     string
		input    DeleteCollectionInput
		wantResp DeleteCollectionOutput
		wantErr  string

		wantClientPath string
		clientResp     string
	}{
		{
			name: "ok to delete collection",
			input: DeleteCollectionInput{
				CollectionName: "mock-collection",
			},
			wantResp: DeleteCollectionOutput{
				Status: "Successfully deleted 1 collection",
			},
			wantClientPath: fmt.Sprintf(deleteCollectionPath, "mock-collection"),
			clientResp:     `null`,
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.Check(r.Method, qt.Equals, http.MethodDelete)
				c.Check(r.URL.Path, qt.Equals, tc.wantClientPath)

				c.Check(r.Header.Get("Content-Type"), qt.Equals, httpclient.MIMETypeJSON)
				c.Check(r.Header.Get("Accept"), qt.Equals, httpclient.MIMETypeJSON)
				c.Check(r.Header.Get("Authorization"), qt.Equals, "Bearer mock-api-key")

				w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
				fmt.Fprintln(w, tc.clientResp)
			})

			milvusServer := httptest.NewServer(h)
			c.Cleanup(milvusServer.Close)

			setup, _ := structpb.NewStruct(map[string]any{
				"api-key": "mock-api-key",
				"url":     milvusServer.URL,
			})

			exec, err := cmp.CreateExecution(base.ComponentExecution{
				Component: cmp,
				Setup:     setup,
				Task:      TaskDeleteCollection,
			})

			c.Assert(err, qt.IsNil)

			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			ir := mock.NewInputReaderMock(c)
			ow := mock.NewOutputWriterMock(c)
			ir.ReadMock.Return([]*structpb.Struct{pbIn}, nil)
			ow.WriteMock.Optional().Set(func(ctx context.Context, outputs []*structpb.Struct) (err error) {
				c.Assert(outputs, qt.HasLen, 1)
				wantJSON, err := json.Marshal(tc.wantResp)
				c.Assert(err, qt.IsNil)
				c.Check(wantJSON, qt.JSONEquals, outputs[0].AsMap())
				return nil
			})

			err = exec.Execute(ctx, ir, ow)
			c.Check(err, qt.IsNil)

		})
	}
}
