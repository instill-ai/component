package asana

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/mock"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	token = "testToken"
)

func TestGetGoal(t *testing.T) {
	testcases := []taskCase[structpb.Struct, structpb.Struct]{
		{
			_type: "happy-path",
			name:  "get-goal",
			input: structpb.Struct{
				Fields: map[string]*structpb.Value{
					"id": {Kind: &structpb.Value_StringValue{StringValue: "123"}},
				},
			},
			wantResp: structpb.Struct{
				Fields: map[string]*structpb.Value{
					"id": {Kind: &structpb.Value_StringValue{StringValue: "123"}},
				},
			},
		},
	}
	taskTesting(testcases, "get-goal", t)
}

type taskCase[inType any, outType any] struct {
	_type    string
	name     string
	input    inType
	wantResp outType
	wantErr  string
}

func taskTesting[inType any, outType any](testcases []taskCase[inType, outType], task string, t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	cmp := Init(bc)

	for _, tc := range testcases {
		c.Run(tc._type+`-`+tc.name, func(c *qt.C) {
			authenticationMiddleware := func(next http.Handler) http.Handler {
				fn := func(w http.ResponseWriter, r *http.Request) {
					c.Check(r.Header.Get("Authorization"), qt.Equals, "Bearer "+token)
					next.ServeHTTP(w, r)
				}
				return http.HandlerFunc(fn)
			}
			setContentTypeMiddleware := func(next http.Handler) http.Handler {
				fn := func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					next.ServeHTTP(w, r)
				}
				return http.HandlerFunc(fn)
			}
			srv := httptest.NewServer(router(authenticationMiddleware, setContentTypeMiddleware))
			c.Cleanup(srv.Close)

			setup, err := structpb.NewStruct(map[string]any{
				"token":    token,
				"base-url": srv.URL,
			})
			c.Assert(err, qt.IsNil)

			e, err := cmp.CreateExecution(base.ComponentExecution{
				Component: cmp,
				Setup:     setup,
				Task:      task,
			})
			c.Assert(err, qt.IsNil)
			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			ir := mock.NewInputReaderMock(c)
			ow := mock.NewOutputWriterMock(c)
			ir.ReadMock.Return([]*structpb.Struct{pbIn}, nil)
			ow.WriteMock.Optional().Set(func(ctx context.Context, outputs []*structpb.Struct) (err error) {
				wantJSON, err := json.Marshal(tc.wantResp)
				c.Assert(err, qt.IsNil)
				c.Assert(outputs, qt.HasLen, 1)
				c.Check(wantJSON, qt.JSONEquals, outputs[0].AsMap())
				return nil
			})
			err = e.Execute(ctx, ir, ow)
			if tc.wantErr != "" {
				c.Assert(err, qt.ErrorMatches, tc.wantErr)
				return
			}
			c.Assert(err, qt.IsNil)

		})
	}
}
