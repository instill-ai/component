package jira

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/mock"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	email = "testemail@gmail.com"
	token = "testToken"
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
			wantErr: "unsuccessful HTTP response.*",
		},
	}
	taskTesting(testcases, taskListBoards, t)
}

func TestComponent_GetIssueTask(t *testing.T) {
	testcases := []TaskCase[GetIssueInput, GetIssueOutput]{
		{
			_type: "ok",
			name:  "get issue-Task",
			input: GetIssueInput{
				IssueKey:      "TST-1",
				UpdateHistory: true,
			},
			wantResp: GetIssueOutput{
				Issue: Issue{
					ID:  "1",
					Key: "TST-1",
					Fields: map[string]interface{}{
						"summary":     "Test issue 1",
						"description": "Test description 1",
						"status": map[string]interface{}{
							"name": "To Do",
						},
						"issuetype": map[string]interface{}{
							"name": "Task",
						},
					},
					Self:        "https://test.atlassian.net/rest/agile/1.0/issue/1",
					Summary:     "Test issue 1",
					Status:      "To Do",
					Description: "Test description 1",
					IssueType:   "Task",
				},
			},
		},
		{
			_type: "ok",
			name:  "get issue-Epic",
			input: GetIssueInput{
				IssueKey:      "KAN-4",
				UpdateHistory: false,
			},
			wantResp: GetIssueOutput{
				Issue: Issue{
					ID:  "4",
					Key: "KAN-4",
					Fields: map[string]interface{}{
						"summary": "Test issue 4",
						"status": map[string]interface{}{
							"name": "Done",
						},
						"issuetype": map[string]interface{}{
							"name": "Epic",
						},
					},
					Self:      "https://test.atlassian.net/rest/agile/1.0/issue/4",
					Summary:   "Test issue 4",
					Status:    "Done",
					IssueType: "Epic",
				},
			},
		},
		{
			_type: "nok",
			name:  "404 - Not Found",
			input: GetIssueInput{
				IssueKey:      "5",
				UpdateHistory: true,
			},
			wantErr: "unsuccessful HTTP response.*",
		},
	}
	taskTesting(testcases, taskGetIssue, t)
}

func TestComponent_GetSprintTask(t *testing.T) {
	testcases := []TaskCase[GetSprintInput, GetSprintOutput]{
		{
			_type: "ok",
			name:  "get sprint",
			input: GetSprintInput{
				SprintID: 1,
			},
			wantResp: GetSprintOutput{
				ID:            1,
				Self:          "https://test.atlassian.net/rest/agile/1.0/sprint/1",
				State:         "active",
				Name:          "Sprint 1",
				StartDate:     "2021-01-01T00:00:00.000Z",
				EndDate:       "2021-01-15T00:00:00.000Z",
				CompleteDate:  "2021-01-15T00:00:00.000Z",
				OriginBoardID: 1,
				Goal:          "Sprint goal",
			},
		},
		{
			_type: "nok",
			name:  "400 - Bad Request",
			input: GetSprintInput{
				SprintID: -1,
			},
			wantErr: "unsuccessful HTTP response.*",
		},
		{
			_type: "nok",
			name:  "404 - Not Found",
			input: GetSprintInput{
				SprintID: 2,
			},
			wantErr: "unsuccessful HTTP response.*",
		},
	}
	taskTesting(testcases, taskGetSprint, t)
}

func TestComponent_ListIssuesTask(t *testing.T) {
	testcases := []TaskCase[ListIssuesInput, ListIssuesOutput]{
		{
			_type: "ok",
			name:  "All",
			input: ListIssuesInput{
				BoardName:  "KAN",
				MaxResults: 10,
				StartAt:    0,
				Range: Range{
					Range: "All",
				},
			},
			wantResp: ListIssuesOutput{
				Total:      1,
				StartAt:    0,
				MaxResults: 10,
				Issues: []Issue{
					{
						ID:  "4",
						Key: "KAN-4",
						Fields: map[string]interface{}{
							"summary": "Test issue 4",
							"status": map[string]interface{}{
								"name": "Done",
							},
							"issuetype": map[string]interface{}{
								"name": "Epic",
							},
						},
						IssueType: "Epic",
						Self:      "https://test.atlassian.net/rest/agile/1.0/issue/4",
						Status:    "Done",
						Summary:   "Test issue 4",
					},
				},
			},
		},
		{
			_type: "ok",
			name:  "Epics only",
			input: ListIssuesInput{
				BoardName:  "KAN",
				MaxResults: 10,
				StartAt:    0,
				Range: Range{
					Range: "Epics only",
				},
			},
			wantResp: ListIssuesOutput{
				Total:      1,
				StartAt:    0,
				MaxResults: 10,
				Issues: []Issue{
					{
						ID:  "4",
						Key: "KAN-4",
						Fields: map[string]interface{}{
							"summary": "Test issue 4",
							"status": map[string]interface{}{
								"name": "Done",
							},
							"issuetype": map[string]interface{}{
								"name": "Epic",
							},
						},
						IssueType: "Epic",
						Self:      "https://test.atlassian.net/rest/agile/1.0/issue/4",
						Status:    "Done",
						Summary:   "Test issue 4",
					},
				},
			},
		},
		{
			_type: "ok",
			name:  "In backlog only",
			input: ListIssuesInput{
				BoardName:  "KAN",
				MaxResults: 10,
				StartAt:    0,
				Range: Range{
					Range: "In backlog only",
				},
			},
			wantResp: ListIssuesOutput{
				Total:      1,
				StartAt:    0,
				MaxResults: 10,
				Issues: []Issue{
					{
						ID:  "4",
						Key: "KAN-4",
						Fields: map[string]interface{}{
							"summary": "Test issue 4",
							"status": map[string]interface{}{
								"name": "Done",
							},
							"issuetype": map[string]interface{}{
								"name": "Epic",
							},
						},
						IssueType: "Epic",
						Self:      "https://test.atlassian.net/rest/agile/1.0/issue/4",
						Status:    "Done",
						Summary:   "Test issue 4",
					},
				},
			},
		},
		{
			_type: "ok",
			name:  "Issues without epic assigned",
			input: ListIssuesInput{
				BoardName:  "KAN",
				MaxResults: 10,
				StartAt:    0,
				Range: Range{
					Range: "Issues without epic assigned",
				},
			},
			wantResp: ListIssuesOutput{
				Total:      1,
				StartAt:    0,
				MaxResults: 10,
				Issues: []Issue{
					{
						ID:  "4",
						Key: "KAN-4",
						Fields: map[string]interface{}{
							"summary": "Test issue 4",
							"status": map[string]interface{}{
								"name": "Done",
							},
							"issuetype": map[string]interface{}{
								"name": "Epic",
							},
						},
						IssueType: "Epic",
						Self:      "https://test.atlassian.net/rest/agile/1.0/issue/4",
						Status:    "Done",
						Summary:   "Test issue 4",
					},
				},
			},
		},
		{
			_type: "ok",
			name:  "Issues of an epic",
			input: ListIssuesInput{
				BoardName:  "KAN",
				MaxResults: 10,
				StartAt:    0,
				Range: Range{
					Range:   "Issues of an epic",
					EpicKey: "KAN-4",
				},
			},
			wantResp: ListIssuesOutput{
				Total:      0,
				StartAt:    0,
				MaxResults: 10,
				Issues:     []Issue{},
			},
		},
		{
			_type: "ok",
			name:  "Issues of an epic(long query)",
			input: ListIssuesInput{
				BoardName:  "KAN",
				MaxResults: 10,
				StartAt:    0,
				Range: Range{
					Range:   "Issues of an epic",
					EpicKey: "KAN-4" + strings.Repeat("-0", 100),
				},
			},
			wantResp: ListIssuesOutput{
				Total:      0,
				StartAt:    0,
				MaxResults: 10,
				Issues:     []Issue{},
			},
		},
		{
			_type: "ok",
			name:  "Issues of a sprint",
			input: ListIssuesInput{
				BoardName:  "KAN",
				MaxResults: 10,
				StartAt:    0,
				Range: Range{
					Range:      "Issues of a sprint",
					SprintName: "KAN Sprint 1",
				},
			},
			wantResp: ListIssuesOutput{
				Total:      0,
				StartAt:    0,
				MaxResults: 10,
				Issues:     []Issue{},
			},
		},
		{
			_type: "ok",
			name:  "Standard Issues",
			input: ListIssuesInput{
				BoardName:  "TST",
				MaxResults: 10,
				StartAt:    0,
				Range: Range{
					Range: "Standard Issues",
				},
			},
			wantResp: ListIssuesOutput{
				Total:      0,
				StartAt:    0,
				MaxResults: 10,
				Issues:     []Issue{},
			},
		},
		{
			_type: "ok",
			name:  "JQL",
			input: ListIssuesInput{
				BoardName:  "TST",
				MaxResults: 10,
				StartAt:    0,
				Range: Range{
					Range: "JQL query",
					JQL:   "project = TST",
				},
			},
			wantResp: ListIssuesOutput{
				Total:      0,
				StartAt:    0,
				MaxResults: 10,
				Issues:     []Issue{},
			},
		},
		{
			_type: "nok",
			name:  "invalid range",
			input: ListIssuesInput{
				BoardName:  "TST",
				MaxResults: 10,
				StartAt:    0,
				Range: Range{
					Range: "invalid",
				},
			},
			wantErr: "invalid range",
		},
	}
	taskTesting(testcases, taskListIssues, t)
}

func TestComponent_ListSprintsTask(t *testing.T) {
	testcases := []TaskCase[ListSprintInput, ListSprintsOutput]{
		{
			_type: "ok",
			name:  "get all sprints",
			input: ListSprintInput{
				BoardID:    1,
				StartAt:    0,
				MaxResults: 10,
			},
			wantResp: ListSprintsOutput{
				Total:      1,
				StartAt:    0,
				MaxResults: 10,
				Sprints: []*GetSprintOutput{
					{
						ID:            1,
						Self:          "https://test.atlassian.net/rest/agile/1.0/sprint/1",
						State:         "active",
						Name:          "Sprint 1",
						StartDate:     "2021-01-01T00:00:00.000Z",
						EndDate:       "2021-01-15T00:00:00.000Z",
						CompleteDate:  "2021-01-15T00:00:00.000Z",
						OriginBoardID: 1,
						Goal:          "Sprint goal",
					},
				},
			},
		},
		{
			_type: "nok",
			name:  "400 - Bad Request",
			input: ListSprintInput{
				BoardID:    -1,
				StartAt:    0,
				MaxResults: 10,
			},
			wantErr: "unsuccessful HTTP response.*",
		},
	}
	taskTesting(testcases, taskListSprints, t)
}

func TestComponent_CreateIssueTask(t *testing.T) {
	testcases := []TaskCase[CreateIssueInput, CreateIssueOutput]{
		{
			_type: "ok",
			name:  "create issue",
			input: CreateIssueInput{
				ProjectKey: "CRI",
				IssueType: IssueType{
					IssueType: "Task",
				},
				Summary:     "Test issue 1",
				Description: "Test description 1",
			},
			wantResp: CreateIssueOutput{
				Issue{
					ID:  "30000",
					Key: "CRI-1",
					Fields: map[string]interface{}{
						"summary":     "Test issue 1",
						"description": "Test description 1",
						"issuetype": map[string]interface{}{
							"name": "Task",
						},
						"project": map[string]interface{}{
							"key": "CRI",
						},
					},
					Self:        "https://test.atlassian.net/rest/agile/1.0/issue/30000",
					Summary:     "Test issue 1",
					Description: "Test description 1",
					IssueType:   "Task",
				},
			},
		},
		{
			_type: "nok",
			name:  "400 - Bad Request",
			input: CreateIssueInput{
				ProjectKey: "INVALID",
			},
			wantErr: "unsuccessful HTTP response.*",
		},
	}
	taskTesting(testcases, taskCreateIssue, t)
}

func TestComponent_UpdateIssueTask(t *testing.T) {
	testcases := []TaskCase[UpdateIssueInput, UpdateIssueOutput]{
		{
			_type: "ok",
			name:  "update issue",
			input: UpdateIssueInput{
				IssueKey: "TST-1",
				Update: Update{
					UpdateType: "Custom Update",
					UpdateFields: []UpdateField{
						{
							FieldName: "summary",
							Action:    "set",
							Value:     "Test issue 1 updated",
						},
						{
							FieldName: "description",
							Action:    "set",
							Value:     "Test description 1 updated",
						},
					},
				},
			},
			wantResp: UpdateIssueOutput{
				Issue{
					ID:  "1",
					Key: "TST-1",
					Fields: map[string]interface{}{
						"summary":     "Test issue 1 updated",
						"description": "Test description 1 updated",
						"status": map[string]interface{}{
							"name": "To Do",
						},
						"issuetype": map[string]interface{}{
							"name": "Task",
						},
					},
					Self:        "https://test.atlassian.net/rest/agile/1.0/issue/1",
					Summary:     "Test issue 1 updated",
					Status:      "To Do",
					Description: "Test description 1 updated",
					IssueType:   "Task",
				},
			},
		},
		{
			_type: "nok",
			name:  "400 - Bad Request",
			input: UpdateIssueInput{
				IssueKey: "INVALID",
				Update: Update{
					UpdateType:   "Custom Update",
					UpdateFields: []UpdateField{},
				},
			},
			wantErr: "unsuccessful HTTP response.*",
		},
	}
	taskTesting(testcases, taskUpdateIssue, t)
}

func TestComponent_CreateSprintTask(t *testing.T) {
	testcases := []TaskCase[CreateSprintInput, CreateSprintOutput]{
		{
			_type: "ok",
			name:  "create sprint",
			input: CreateSprintInput{
				Name:      "Test Sprint",
				Goal:      "Sprint goal",
				StartDate: "2021-01-01T00:00:00.000Z",
				EndDate:   "2021-01-15T00:00:00.000Z",
				BoardName: "TST",
			},
			wantResp: CreateSprintOutput{
				ID:            1,
				Self:          "https://test.atlassian.net/rest/agile/1.0/sprint/1",
				State:         "active",
				Name:          "Test Sprint",
				StartDate:     "2021-01-01T00:00:00.000Z",
				EndDate:       "2021-01-15T00:00:00.000Z",
				CompleteDate:  "",
				OriginBoardID: 3,
				Goal:          "Sprint goal",
			},
		},
		{
			_type: "nok",
			name:  "400 - Bad Request",
			input: CreateSprintInput{
				Name:      "Test Sprint",
				BoardName: "INVALID",
			},
			wantErr: "end date is required",
		},
		{
			_type: "nok",
			name:  "400 - Bad Request",
			input: CreateSprintInput{
				Name:      "Test Sprint",
				BoardName: "INVALID",
				EndDate:   "2021-01-15T00:00:00.000Z",
			},
			wantErr: "board not found",
		},
	}
	taskTesting(testcases, taskCreateSprint, t)
}

func TestComponent_UpdateSprintTask(t *testing.T) {
	testcases := []TaskCase[UpdateSprintInput, UpdateSprintOutput]{
		{
			_type: "ok",
			name:  "update sprint",
			input: UpdateSprintInput{
				SprintID:       1,
				Name:           "Test Sprint updated",
				Goal:           "Sprint goal updated",
				StartDate:      "2021-01-01T00:00:00.000Z",
				EndDate:        "2021-01-15T00:00:00.000Z",
				CurrentState:   "active",
				EnterNextState: false,
			},
			wantResp: UpdateSprintOutput{
				ID:            1,
				Self:          "https://test.atlassian.net/rest/agile/1.0/sprint/1",
				State:         "active",
				Name:          "Test Sprint updated",
				StartDate:     "2021-01-01T00:00:00.000Z",
				EndDate:       "2021-01-15T00:00:00.000Z",
				CompleteDate:  "2021-01-15T00:00:00.000Z",
				OriginBoardID: 1,
				Goal:          "Sprint goal updated",
			},
		},
		{
			_type: "ok",
			name:  "future to active",
			input: UpdateSprintInput{
				SprintID:       1,
				Name:           "Test Sprint updated",
				Goal:           "Sprint goal updated",
				StartDate:      "2021-01-01T00:00:00.000Z",
				EndDate:        "2021-01-15T00:00:00.000Z",
				CurrentState:   "future",
				EnterNextState: true,
			},
			wantResp: UpdateSprintOutput{
				ID:            1,
				Self:          "https://test.atlassian.net/rest/agile/1.0/sprint/1",
				State:         "active",
				Name:          "Test Sprint updated",
				StartDate:     "2021-01-01T00:00:00.000Z",
				EndDate:       "2021-01-15T00:00:00.000Z",
				CompleteDate:  "2021-01-15T00:00:00.000Z",
				OriginBoardID: 1,
				Goal:          "Sprint goal updated",
			},
		},
		{
			_type: "ok",
			name:  "active to closed",
			input: UpdateSprintInput{
				SprintID:       1,
				Name:           "Test Sprint updated",
				Goal:           "Sprint goal updated",
				StartDate:      "2021-01-01T00:00:00.000Z",
				EndDate:        "2021-01-15T00:00:00.000Z",
				CurrentState:   "active",
				EnterNextState: true,
			},
			wantResp: UpdateSprintOutput{
				ID:            1,
				Self:          "https://test.atlassian.net/rest/agile/1.0/sprint/1",
				State:         "closed",
				Name:          "Test Sprint updated",
				StartDate:     "2021-01-01T00:00:00.000Z",
				EndDate:       "2021-01-15T00:00:00.000Z",
				CompleteDate:  "2021-01-15T00:00:00.000Z",
				OriginBoardID: 1,
				Goal:          "Sprint goal updated",
			},
		},
		{
			_type: "nok",
			name:  "400 - Bad Request",
			input: UpdateSprintInput{
				SprintID: -1,
			},
			wantErr: "end date is required",
		},
	}
	taskTesting(testcases, taskUpdateSprint, t)
}

func TestAuth_nok(t *testing.T) {
	c := qt.New(t)
	bc := base.Component{Logger: zap.NewNop()}
	cmp := Init(bc)
	c.Run("nok-empty token", func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"token":    "",
			"email":    email,
			"base-url": "url",
		})
		c.Assert(err, qt.IsNil)
		_, err = cmp.CreateExecution(base.ComponentExecution{
			Component: cmp,
			Setup:     setup,
			Task:      "invalid",
		})
		c.Assert(err, qt.ErrorMatches, "token not provided")
	})
	c.Run("nok-empty email", func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"token":    token,
			"email":    "",
			"base-url": "url",
		})
		c.Assert(err, qt.IsNil)
		_, err = cmp.CreateExecution(base.ComponentExecution{
			Component: cmp,
			Setup:     setup,
			Task:      "invalid",
		})
		c.Assert(err, qt.ErrorMatches, "email not provided")
	})
}

func taskTesting[inType any, outType any](testcases []TaskCase[inType, outType], task string, t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	cmp := Init(bc)

	for _, tc := range testcases {
		c.Run(tc._type+`-`+tc.name, func(c *qt.C) {
			authenticationMiddleware := func(next http.Handler) http.Handler {
				fn := func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/_edge/tenant_info" {
						auth := base64.StdEncoding.EncodeToString([]byte(email + ":" + token))
						c.Check(r.Header.Get("Authorization"), qt.Equals, "Basic "+auth)
					}
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
				"email":    email,
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
