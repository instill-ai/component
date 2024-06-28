package github

import (
	"context"
	"encoding/json"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

var MockGithubClient = GitHubClient{
	Repositories: &MockRepositoriesService{},
	PullRequests: &MockPullRequestService{},
}

var fakeHost = "https://fake-github.com"


const (
	apiKey = "testkey"
)

func TestComponent_GetAllPullRequestsTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    GetAllPullRequestsInput
		wantResp GetAllPullRequestsResp
		wantErr  string
	}{
		{
			name: "ok to write",
			input: GetAllPullRequestsInput{
				Owner: "test_owner",
				Repository:  "test_repo",
				State: "open",
				Direction: "asc",
				Sort: "created",
			},
			wantResp: GetAllPullRequestsResp{
				PullRequests: []PullRequest{
					{
						Base: "03a09068472b674749d1e8336679b84a31284e52",
						Body: "hi",
						CommentsNum: 1,
						Commits: []Commit{
							{
								Message: "This is a test to create comment",
								SHA: "5e1fc1c381cc59ff2ed35335b58651e90b1ad4fb",
								Stats: CommitStats{},
							},
						},
						CommitsNum: 1,
						CommitsURL: "https://api.github.com/repos/YCK1130/Github-API-test/pulls/1/commits",
						DiffURL: "https://github.com/YCK1130/Github-API-test/pull/1.diff",
						Head: "5e1fc1c381cc59ff2ed35335b58651e90b1ad4fb",
						ID: 1943139364,
						Number: 1,
						ReviewCommentsNum: 6,
						State: "open",
						Title: "This is a pr",
					},
				},
			},
		},
		{
			name: "fail to write",
			input: GetAllPullRequestsInput{
				Owner: "test_owner",
				Repository:  "test_repo",
				State: "open",
				Direction: "asc",
				Sort: "created",
			},
			wantErr: `403 API rate limit exceeded`,
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {

			setup, err := structpb.NewStruct(map[string]any{
				"api-key": apiKey,
			})
			c.Assert(err, qt.IsNil)

			// It will increase the modification range if we change the input of CreateExecution.
			// So, we replaced it with the code below to cover the test for taskFunctions.go
			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: taskGetAllPRs},
				client:             Client{client: MockGithubClient},
			}
			e.execute = e.client.getAllPullRequestsTask
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
