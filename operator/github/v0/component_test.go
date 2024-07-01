package github

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-github/v62/github"
	"github.com/instill-ai/component/base"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

var MockGithubClient = GitHubClient{
	Repositories: &MockRepositoriesService{},
	PullRequests: &MockPullRequestService{},
	Issues: &MockIssuesService{},
}

var fakeHost = "https://fake-github.com"


const (
	token = "testkey"
)

type TaskCase[inType any, outType any] struct {
	_type 	 string
	name     string
	input    inType
	wantResp outType
	wantErr  string
}
func TestComponent_GetAllPullRequestsTask(t *testing.T) {
	testcases := []TaskCase[GetAllPullRequestsInput, GetAllPullRequestsResp]{
		{
			_type: "ok",
			name: "get all pull requests",
			input: GetAllPullRequestsInput{
				Owner: "test_owner",
				Repository:  "test_repo",
				State: "open",
				Direction: "asc",
				Sort: "created",
			},
			// TODO: These are bullshit values. Need to replace them with real values.
			wantResp: GetAllPullRequestsResp{
				PullRequests: []PullRequest{
					{
						Base: "baseSHA",
						Body: "PR Body",
						Commits: []Commit{
							{
								Message: "This is a fake commit",
								SHA: "commitSHA",
								Stats: CommitStats{
									Additions: 1,
									Deletions: 1,
									Changes: 2,
								},
								Files: []CommitFile{
									{
										Filename: "filename",
										Patch: "patch",
										CommitStats: CommitStats{
											Additions: 1,
											Deletions: 1,
											Changes: 2,
										},
									},
								},
							},
						},
						DiffURL: "https://fake-github.com/test_owner/test_repo/pull/1.diff",
						Head: "headSHA",
						ID: 1,
						Number: 1,
						CommentsNum: 0,
						CommitsNum: 1,
						ReviewCommentsNum: 2,
						State: "open",
						Title: "This is a fake PR",
					},
				},
			},
		},
		{
			_type: "nok",
			name: "403 API rate limit exceeded",
			input: GetAllPullRequestsInput{
				Owner: "rate_limit",
				Repository:  "test_repo",
				State: "open",
				Direction: "asc",
				Sort: "created",
			},
			wantErr: `403 API rate limit exceeded`,
		},
		{
			_type: "nok",
			name: "404 Not Found",
			input: GetAllPullRequestsInput{
				Owner: "not_found",
				Repository:  "test_repo",
				State: "open",
				Direction: "asc",
				Sort: "created",
			},
			wantErr: `404 Not Found`,
		},
		{
			_type: "nok",
			name: "no PRs found",
			input: GetAllPullRequestsInput{
				Owner: "no_pr",
				Repository:  "test_repo",
				State: "open",
				Direction: "asc",
				Sort: "created",
			},
			wantResp: GetAllPullRequestsResp{
				PullRequests: []PullRequest{},
			},
		},
	}
	taskTesting(testcases, taskGetAllPRs, t)
}

func TestComponent_GetPullRequestTask(t *testing.T) {
	testcases := []TaskCase[GetPullRequestInput, GetPullRequestResp]{
		{
			_type: "ok",
			name: "get pull request",
			input: GetPullRequestInput{
				Owner: "test_owner",
				Repository:  "test_repo",
				PrNumber: 1,
			},
			wantResp: GetPullRequestResp{
				PullRequest: PullRequest{
					Base: "baseSHA",
					Body: "PR Body",
					Commits: []Commit{
						{
							Message: "This is a fake commit",
							SHA: "commitSHA",
							Stats: CommitStats{
								Additions: 1,
								Deletions: 1,
								Changes: 2,
							},
							Files: []CommitFile{
								{
									Filename: "filename",
									Patch: "patch",
									CommitStats: CommitStats{
										Additions: 1,
										Deletions: 1,
										Changes: 2,
									},
								},
							},
						},
					},
					DiffURL: "https://fake-github.com/test_owner/test_repo/pull/1.diff",
					Head: "headSHA",
					ID: 1,
					Number: 1,
					CommentsNum: 0,
					CommitsNum: 1,
					ReviewCommentsNum: 2,
					State: "open",
					Title: "This is a fake PR",
				},
			},
		},
		{
			_type: "nok",
			name: "403 API rate limit exceeded",
			input: GetPullRequestInput{
				Owner: "rate_limit",
				Repository:  "test_repo",
				PrNumber: 1,
			},
			wantErr: `403 API rate limit exceeded`,
		},
		{
			_type: "nok",
			name: "404 Not Found",
			input: GetPullRequestInput{
				Owner: "not_found",
				Repository:  "test_repo",
				PrNumber: 1,
			},
			wantErr: `404 Not Found`,
		},
	}
	taskTesting(testcases, taskGetPR, t)
}

func TestComponent_GetAllReviewCommentsTask(t *testing.T) {
	testcases := []TaskCase[GetAllReviewCommentsInput, GetAllReviewCommentsResp]{
		{
			_type: "ok",
			name: "get review comments",
			input: GetAllReviewCommentsInput{
				Owner: "test_owner",
				Repository:  "test_repo",
				PrNumber: 1,
				Sort: "created",
				Direction: "asc",
				Since: "2021-01-01T00:00:00Z",
			},
			wantResp: GetAllReviewCommentsResp{
				ReviewComments: []ReviewComment{
					{
						PullRequestComment: github.PullRequestComment{
							Body: github.String("This is a fake comment"),
							ID: github.Int64(1),
						},
					},
				},
			},
		},
		{
			_type: "nok",
			name: "403 API rate limit exceeded",
			input: GetAllReviewCommentsInput{
				Owner: "rate_limit",
				Repository:  "test_repo",
				PrNumber: 1,
				Sort: "created",
				Direction: "asc",
				Since: "2021-01-01T00:00:00Z",
			},
			wantErr: `403 API rate limit exceeded`,
		},
		{
			_type: "nok",
			name: "404 Not Found",
			input: GetAllReviewCommentsInput{
				Owner: "not_found",
				Repository:  "test_repo",
				PrNumber: 1,
				Sort: "created",
				Direction: "asc",
				Since: "2021-01-01T00:00:00Z",
			},
			wantErr: `404 Not Found`,
		},
		{
			_type: "nok",
			name: "invalid time format",
			input: GetAllReviewCommentsInput{
				Owner: "not_found",
				Repository:  "test_repo",
				PrNumber: 1,
				Sort: "created",
				Direction: "asc",
				Since: "2021-0100:00:00Z",
			},
			wantErr: `invalid time format`,
		},
	}
	taskTesting(testcases, taskGetReviewComments, t)
}

func TestComponent_CreateReviewCommentTask(t *testing.T) {
	testcases := []TaskCase[CreateReviewCommentInput, CreateReviewCommentResp]{
		{
			_type: "ok",
			name: "create review comment",
			input: CreateReviewCommentInput{
				Owner: "test_owner",
				Repository:  "test_repo",
				PrNumber: 1,
				Comment: github.PullRequestComment{
					Body: github.String("This is a fake comment"),
					Line: github.Int(2),
					StartLine: github.Int(1),
					Side: github.String("side"),
					StartSide: github.String("side"),
					SubjectType: github.String("line"),
				},
			},
			wantResp: CreateReviewCommentResp{
				ReviewComment: ReviewComment{
					PullRequestComment: github.PullRequestComment{
						Body: github.String("This is a fake comment"),
						ID: github.Int64(1),
						Line: github.Int(2),
						Position: github.Int(2),
						StartLine: github.Int(1),
						Side: github.String("side"),
						StartSide: github.String("side"),
						SubjectType: github.String("line"),
					},
				},
			},
		},
		{
			_type: "nok",
			name: "403 API rate limit exceeded",
			input: CreateReviewCommentInput{
				Owner: "rate_limit",
				Repository:  "test_repo",
				PrNumber: 1,
				Comment: github.PullRequestComment{
					Body: github.String("This is a fake comment"),
					Line: github.Int(2),
					StartLine: github.Int(1),
					Side: github.String("RIGHT"),
					StartSide: github.String("RIGHT"),
					SubjectType: github.String("line"),
				},
			},
			wantErr: `403 API rate limit exceeded`,
		},
		{
			_type: "nok",
			name: "404 Not Found",
			input: CreateReviewCommentInput{
				Owner: "not_found",
				Repository:  "test_repo",
				PrNumber: 1,
				Comment: github.PullRequestComment{
					Body: github.String("This is a fake comment"),
					Line: github.Int(2),
					StartLine: github.Int(1),
					Side: github.String("RIGHT"),
					StartSide: github.String("RIGHT"),
					SubjectType: github.String("line"),
				},
			},
			wantErr: `404 Not Found`,
		},
		{
			_type: "nok",
			name: "422 Unprocessable Entity",
			input: CreateReviewCommentInput{
				Owner: "unprocessable_entity",
				Repository:  "test_repo",
				PrNumber: 1,
				Comment: github.PullRequestComment{
					Body: github.String("This is a fake comment"),
					Line: github.Int(2),
					StartLine: github.Int(1),
					Side: github.String("RIGHT"),
					StartSide: github.String("RIGHT"),
					SubjectType: github.String("line"),
				},
			},
			wantErr: `422 Unprocessable Entity`,
		},
		{
			_type: "nok",
			name: "422 Unprocessable Entity",
			input: CreateReviewCommentInput{
				Owner: "test_owner",
				Repository:  "test_repo",
				PrNumber: 1,
				Comment: github.PullRequestComment{
					Body: github.String("This is a fake comment"),
					Line: github.Int(1),
					StartLine: github.Int(1),
					Side: github.String("RIGHT"),
					StartSide: github.String("RIGHT"),
					SubjectType: github.String("line"),
				},
			},
			wantErr: `422 Unprocessable Entity`,
		},
	}
	taskTesting(testcases, taskCreateReviewComment, t)
}

func TestComponent_GetCommitTask(t *testing.T){
	testcases := []TaskCase[GetCommitInput, GetCommitResp]{
		{
			_type: "ok",
			name: "get commit",
			input: GetCommitInput{
				Owner: "test_owner",
				Repository:  "test_repo",
				SHA: "commitSHA",
			},
			wantResp: GetCommitResp{
				Commit: Commit{
					Message: "This is a fake commit",
					SHA: "commitSHA",
					Stats: CommitStats{
						Additions: 1,
						Deletions: 1,
						Changes: 2,
					},
					Files: []CommitFile{
						{
							Filename: "filename",
							Patch: "patch",
							CommitStats: CommitStats{
								Additions: 1,
								Deletions: 1,
								Changes: 2,
							},
						},
					},
				},
			},
		},
		{
			_type: "nok",
			name: "403 API rate limit exceeded",
			input: GetCommitInput{
				Owner: "rate_limit",
				Repository:  "test_repo",
				SHA: "commitSHA",
			},
			wantErr: `403 API rate limit exceeded`,
		},
	}
	taskTesting(testcases, taskGetCommit, t)
}

func TestComponent_GetAllIssuesTask(t * testing.T){
	testcases := []TaskCase[GetAllIssuesInput, GetAllIssuesResp]{
		{
			_type: "ok",
			name: "get all issues",
			input: GetAllIssuesInput{
				Owner: "test_owner",
				Repository:  "test_repo",
				State: "open",
				Direction: "asc",
				Sort: "created",
				Since: "2021-01-01T00:00:00Z",
				NoPullRequest: true,
			},
			wantResp: GetAllIssuesResp{
				Issues: []Issue{
					{
						Number: 1,
						Title: "This is a fake Issue",
						Body: "Issue Body",
						State: "open",
						Assignee: "assignee",
						Assignees: []string{"assignee1", "assignee2"},
						Labels: []string{"label1", "label2"},
					},
				},
			},
		},
		{
			_type: "nok",
			name: "403 API rate limit exceeded",
			input: GetAllIssuesInput{
				Owner: "rate_limit",
				Repository:  "test_repo",
				State: "open",
				Direction: "asc",
				Sort: "created",
				Since: "2021-01-01T00:00:00Z",
				NoPullRequest: true,
			},
			wantErr: `403 API rate limit exceeded`,
		},
		{
			_type: "nok",
			name: "404 Not Found",
			input: GetAllIssuesInput{
				Owner: "not_found",
				Repository:  "test_repo",
				State: "open",
				Direction: "asc",
				Sort: "created",
				Since: "2021-01-01T00:00:00Z",
				NoPullRequest: true,
			},
			wantErr: `404 Not Found`,
		},
		{
			_type: "nok",
			name: "invalid time format",
			input: GetAllIssuesInput{
				Owner: "not_found",
				Repository:  "test_repo",
				State: "open",
				Direction: "asc",
				Sort: "created",
				Since: "2021-0Z",
				NoPullRequest: true,
			},
			wantErr: `invalid time format`,
		},
	}
	taskTesting(testcases, taskGetAllIssues, t)
}
func TestComponent_GetIssueTask(t * testing.T){
	testcases := []TaskCase[GetIssueInput, GetIssueResp]{
		{
			_type: "ok",
			name: "get all issues",
			input: GetIssueInput{
				Owner: "test_owner",
				Repository:  "test_repo",
				IssueNumber: 1,
			},
			wantResp: GetIssueResp{
				Issue: Issue{
					Number: 1,
					Title: "This is a fake Issue",
					Body: "Issue Body",
					State: "open",
					Assignee: "assignee",
					Assignees: []string{"assignee1", "assignee2"},
					Labels: []string{"label1", "label2"},
					IsPullRequest: false,
				},
			},
		},
		{
			_type: "nok",
			name: "403 API rate limit exceeded",
			input: GetIssueInput{
				Owner: "rate_limit",
				Repository:  "test_repo",
				IssueNumber: 1,
			},
			wantErr: `403 API rate limit exceeded`,
		},
		{
			_type: "nok",
			name: "404 Not Found",
			input: GetIssueInput{
				Owner: "not_found",
				Repository:  "test_repo",
				IssueNumber: 1,
			},
			wantErr: `404 Not Found`,
		},
	}
	taskTesting(testcases, taskGetIssue, t)
}
func TestComponent_CreateIssueTask(t * testing.T){
	testcases := []TaskCase[CreateIssueInput, CreateIssueResp]{
		{
			_type: "ok",
			name: "get all issues",
			input: CreateIssueInput{
				Owner: "test_owner",
				Repository:  "test_repo",
				Title: "This is a fake Issue",
				Body: "Issue Body",
			},
			wantResp: CreateIssueResp{
				Issue: Issue{
					Number: 1,
					Title: "This is a fake Issue",
					Body: "Issue Body",
					State: "open",
					IsPullRequest: false,
					Assignees: []string{},
					Labels: []string{},
					Assignee: "",
				},
			},
		},
		{
			_type: "nok",
			name: "403 API rate limit exceeded",
			input: CreateIssueInput{
				Owner: "rate_limit",
				Repository:  "test_repo",
				Title: "This is a fake Issue",
				Body: "Issue Body",
			},
			wantErr: `403 API rate limit exceeded`,
		},
		{
			_type: "nok",
			name: "404 Not Found",
			input: CreateIssueInput{
				Owner: "not_found",
				Repository:  "test_repo",
				Title: "This is a fake Issue",
				Body: "Issue Body",
			},
			wantErr: `404 Not Found`,
		},
	}
	taskTesting(testcases, taskCreateIssue, t)
}



func taskTesting[inType any, outType any](testcases []TaskCase[inType, outType], task string, t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	for _, tc := range testcases {
		c.Run(tc._type + `-` + tc.name, func(c *qt.C) {

			setup, err := structpb.NewStruct(map[string]any{
				"token": token,
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: task},
				client:             Client{client: MockGithubClient},
			}
			switch task {
			case taskGetAllPRs:
				e.execute = e.client.getAllPullRequestsTask
			case taskGetPR:
				e.execute = e.client.getPullRequestTask
			case taskGetReviewComments:
				e.execute = e.client.getAllReviewCommentsTask
			case taskCreateReviewComment:
				e.execute = e.client.createReviewCommentTask
			case taskGetCommit:
				e.execute = e.client.getCommitTask
			case taskGetAllIssues:
				e.execute = e.client.getAllIssuesTask
			case taskGetIssue:
				e.execute = e.client.getIssueTask
			case taskCreateIssue:
				e.execute = e.client.createIssueTask
			default:
				c.Fatalf("not supported testing task: %s", task)
			}
			exec := &base.ExecutionWrapper{Execution: e}

			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			got, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})
			fmt.Println(got)
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
