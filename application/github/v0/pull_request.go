package github

import (
	"context"

	"fmt"

	"github.com/google/go-github/v62/github"
	"github.com/instill-ai/component/base"
	"github.com/instill-ai/x/errmsg"
	"google.golang.org/protobuf/types/known/structpb"
)

type PullRequestService interface {
	List(context.Context, string, string, *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error)
	Get(context.Context, string, string, int) (*github.PullRequest, *github.Response, error)
	ListComments(context.Context, string, string, int, *github.PullRequestListCommentsOptions) ([]*github.PullRequestComment, *github.Response, error)
	CreateComment(context.Context, string, string, int, *github.PullRequestComment) (*github.PullRequestComment, *github.Response, error)
	ListCommits(context.Context, string, string, int, *github.ListOptions) ([]*github.RepositoryCommit, *github.Response, error)
}

type PullRequest struct {
	ID                int64    `json:"id"`
	Number            int      `json:"number"`
	State             string   `json:"state"`
	Title             string   `json:"title"`
	Body              string   `json:"body"`
	DiffURL           string   `json:"diff_url,omitempty"`
	CommitsURL        string   `json:"commits_url,omitempty"`
	Commits           []Commit `json:"commits"`
	Head              string   `json:"head"`
	Base              string   `json:"base"`
	CommentsNum       int      `json:"comments_num"`
	CommitsNum        int      `json:"commits_num"`
	ReviewCommentsNum int      `json:"review_comments_num"`
}

func (githubClient *Client) extractPullRequestInformation(originalPr *github.PullRequest) (PullRequest, error) {
	resp := PullRequest{
		ID:                originalPr.GetID(),
		Number:            originalPr.GetNumber(),
		State:             originalPr.GetState(),
		Title:             originalPr.GetTitle(),
		Body:              originalPr.GetBody(),
		DiffURL:           originalPr.GetDiffURL(),
		Head:              originalPr.GetHead().GetSHA(),
		Base:              originalPr.GetBase().GetSHA(),
		CommentsNum:       originalPr.GetComments(),
		CommitsNum:        originalPr.GetCommits(),
		ReviewCommentsNum: originalPr.GetReviewComments(),
	}
	if originalPr.GetCommitsURL() != "" {
		commits, _, err := githubClient.client.PullRequests.ListCommits(context.Background(), githubClient.owner, githubClient.repository, resp.Number, nil)
		if err != nil {
			return PullRequest{}, err
		}
		resp.Commits = make([]Commit, len(commits))
		for idx, commit := range commits {
			resp.Commits[idx] = githubClient.extractCommitInformation(commit)
		}
	}
	return resp, nil
}

type GetAllPullRequestsInput struct {
	RepoInfo
	State     string `json:"state"`
	Sort      string `json:"sort"`
	Direction string `json:"direction"`
}
type GetAllPullRequestsResp struct {
	PullRequests []PullRequest `json:"pull_requests"`
}

func (githubClient *Client) getAllPullRequestsTask(props *structpb.Struct) (*structpb.Struct, error) {

	var inputStruct GetAllPullRequestsInput
	err := base.ConvertFromStructpb(props, &inputStruct)
	if err != nil {
		return nil, err
	}
	err = githubClient.setTargetRepo(inputStruct)
	if err != nil {
		return nil, err
	}

	opts := &github.PullRequestListOptions{
		State:     inputStruct.State,
		Sort:      inputStruct.Sort,
		Direction: inputStruct.Direction,
	}
	prs, _, err := githubClient.client.PullRequests.List(context.Background(), githubClient.owner, githubClient.repository, opts)
	if err != nil {
		return nil, err
	}
	PullRequests := make([]PullRequest, len(prs))
	for idx, pr := range prs {
		PullRequests[idx], err = githubClient.extractPullRequestInformation(pr)
		if err != nil {
			return nil, err
		}
	}

	var prResp GetAllPullRequestsResp
	prResp.PullRequests = PullRequests
	out, err := base.ConvertToStructpb(prResp)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type GetPullRequestInput struct {
	RepoInfo
	PrNumber float64 `json:"pr_number"`
}
type GetPullRequestResp struct {
	PullRequest PullRequest `json:"pull_request"`
}

func (githubClient *Client) getPullRequestTask(props *structpb.Struct) (*structpb.Struct, error) {

	var inputStruct GetPullRequestInput
	err := base.ConvertFromStructpb(props, &inputStruct)
	if err != nil {
		return nil, err
	}
	err = githubClient.setTargetRepo(inputStruct)
	if err != nil {
		return nil, err
	}
	number := inputStruct.PrNumber
	var pullRequest *github.PullRequest
	if number > 0 {
		pr, _, err := githubClient.client.PullRequests.Get(context.Background(), githubClient.owner, githubClient.repository, int(number))
		if err != nil {
			// err includes the rate limit, 404 not found, etc.
			// if the connection is not authorized, it's easy to get rate limit error in large scale usage.
			return nil, err
		}
		pullRequest = pr
	} else {
		// Get the latest PR
		opts := &github.PullRequestListOptions{
			State:     "all",
			Sort:      "created",
			Direction: "desc",
		}
		prs, _, err := githubClient.client.PullRequests.List(context.Background(), githubClient.owner, githubClient.repository, opts)
		if err != nil {
			// err includes the rate limit.
			// if the connection is not authorized, it's easy to get rate limit error in large scale usage.
			return nil, err
		}
		if len(prs) == 0 {
			return nil, errmsg.AddMessage(
				fmt.Errorf("no pull request found"),
				"No pull request found",
			)
		}
		pullRequest = prs[0]
		// Some fields are not included in the list API, so we need to get the PR again.
		pr, _, err := githubClient.client.PullRequests.Get(context.Background(), githubClient.owner, githubClient.repository, *pullRequest.Number)
		if err != nil {
			// err includes the rate limit, 404 not found, etc.
			// if the connection is not authorized, it's easy to get rate limit error in large scale usage.
			return nil, err
		}
		pullRequest = pr
	}

	var prResp GetPullRequestResp
	prResp.PullRequest, err = githubClient.extractPullRequestInformation(pullRequest)
	if err != nil {
		return nil, err
	}
	out, err := base.ConvertToStructpb(prResp)
	if err != nil {
		return nil, err
	}
	return out, nil
}
