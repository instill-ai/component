package jira

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/util/httpclient"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	email   = "testemail@gmail.com"
	token   = "testToken"
	errResp = `{"message": "Bad request"}`
	okResp  = `{"title": "Be the wheel"}`
)

type TaskCase[inType any, outType any] struct {
	_type    string
	name     string
	input    inType
	wantResp outType
	wantErr  string
}

func TestComponent_ListBoardsTask(t *testing.T) {
	testcases := []TaskCase[ListBoardsInput, ListBoardsOutput]{
		{
			_type: "ok",
			name:  "get all boards",
			input: ListBoardsInput{
				MaxResults: 10,
				StartAt:    0,
			},
			wantResp: ListBoardsOutput{
				Total:      1,
				StartAt:    0,
				MaxResults: 10,
				IsLast:     true,
				Boards: []Board{
					{
						ID:        3,
						Name:      "TST",
						BoardType: "simple",
						Self:      "https://test.atlassian.net/rest/agile/1.0/board/3",
					},
				},
			},
		},
		{
			_type: "ok",
			name:  "get filtered boards",
			input: ListBoardsInput{
				MaxResults: 10,
				StartAt:    1,
				BoardType:  "kanban",
			},
			wantResp: ListBoardsOutput{
				Total:      1,
				StartAt:    1,
				MaxResults: 10,
				IsLast:     true,
				Boards:     []Board{},
			},
		},
		{
			_type: "nok",
			name:  "400 - Not Found",
			input: ListBoardsInput{
				MaxResults:     10,
				StartAt:        1,
				ProjectKeyOrID: "test",
			},
			wantErr: "unsuccessful HTTP response",
		},
	}
	taskTesting(testcases, taskListBoards, t)
}

func taskTesting[inType any, outType any](testcases []TaskCase[inType, outType], task string, t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	for _, tc := range testcases {
		c.Run(tc._type+`-`+tc.name, func(c *qt.C) {
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/_edge/tenant_info" {
					auth := base64.StdEncoding.EncodeToString([]byte(email + ":" + token))
					c.Check(r.Header.Get("Authorization"), qt.Equals, "Basic "+auth)
				}
				w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
				router(w, r)
			})

			srv := httptest.NewServer(h)
			c.Cleanup(srv.Close)

			setup, err := structpb.NewStruct(map[string]any{
				"token":    token,
				"email":    email,
				"base-url": srv.URL,
			})
			c.Assert(err, qt.IsNil)

			exec, err := connector.CreateExecution(nil, setup, task)
			c.Assert(err, qt.IsNil)
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
