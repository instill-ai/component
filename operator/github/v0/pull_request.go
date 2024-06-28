package github

import (
	"context"

	"fmt"

	"github.com/google/go-github/v62/github"
	"github.com/instill-ai/component/base"
	"github.com/instill-ai/x/errmsg"
	"google.golang.org/protobuf/types/known/structpb"
)

type PullRequest struct {
	ID          		int64           	`json:"id,omitempty"`
	Number      		int             	`json:"number,omitempty"`
	State       		string          	`json:"state,omitempty"`
	Title	 			string 				`json:"title,omitempty"`
	Body 				string 				`json:"body,omitempty"`
	DiffURL 			string 				`json:"diff_url,omitempty"`
	CommitsURL 			string 				`json:"commits_url,omitempty"`
	Commits 			[]Commit 			`json:"commits,omitempty"`
	Head 				string 				`json:"head,omitempty"`
	Base 				string 				`json:"base,omitempty"`
	CommentsNum 		int 				`json:"comments_num,omitempty"`
	CommitsNum 			int 				`json:"commits_num,omitempty"`
	ReviewCommentsNum 	int		 			`json:"review_comments_num,omitempty"`
}
func (githubClient *GitHubClient) extractPullRequestInformation(originalPr *github.PullRequest) (PullRequest, error) {
	resp := PullRequest{
		ID: originalPr.GetID(),
		Number: originalPr.GetNumber(),
		State: originalPr.GetState(),
		Title: originalPr.GetTitle(),
		Body: originalPr.GetBody(),
		DiffURL: originalPr.GetDiffURL(),
		Head: originalPr.GetHead().GetSHA(),
		Base: originalPr.GetBase().GetSHA(),
		CommentsNum: originalPr.GetComments(),
		CommitsNum: originalPr.GetCommits(),
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
		fmt.Println("=====================================")
		fmt.Println("commits: ",resp.Commits)
		fmt.Println("=====================================")
	}
	return resp, nil
}

type GetAllPullRequestsInput struct {
	Owner string `json:"owner"`
	Repository string `json:"repository"`
	State string `json:"state"`
	Sort string `json:"sort"`
	Direction string `json:"direction"`
}
type GetAllPullRequestsResp struct {
	PullRequests   []PullRequest `json:"pull-requests"`
}

func (githubClient *GitHubClient) getAllPullRequestsTask(props *structpb.Struct) (*structpb.Struct, error) {
	err := githubClient.setTargetRepo(props)
	if err != nil {
		return nil, err
	}

	opts := &github.PullRequestListOptions{
		State: props.GetFields()["state"].GetStringValue(),
		Sort: props.GetFields()["sort"].GetStringValue(),
		Direction: props.GetFields()["direction"].GetStringValue(),
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
	Owner string `json:"owner"`
	Repository string `json:"repository"`
	PrNumber float64 `json:"pr_number"`
}
type GetPullRequestResp struct {
	PullRequest		PullRequest `json:"pull-request"`
}

func (githubClient *GitHubClient) getPullRequestTask(props *structpb.Struct) (*structpb.Struct, error) {
	err := githubClient.setTargetRepo(props)
	if err != nil {
		return nil, err
	}

	number := props.GetFields()["pr_number"].GetNumberValue()
	var pullRequest *github.PullRequest
	if number > 0 {
		pr, _, err := githubClient.client.PullRequests.Get(context.Background(), githubClient.owner, githubClient.repository, int(number))
		if err != nil {
			// err includes the rate limit, 404 not found, etc.
			// if the connection is not authorized, it's easy to get rate limit error in large scale usage.
			return nil, err
		}
		pullRequest = pr
	}else {
		// Get the latest PR
		opts := &github.PullRequestListOptions{
			State: "all",
			Sort: "created",
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
	}

	var prResp GetPullRequestResp
	prResp.PullRequest, err = githubClient.extractPullRequestInformation(pullRequest)
	if err != nil {
		return nil, err
	}
	fmt.Println("===========================")
	fmt.Println("prResp.PullRequest: ",prResp.PullRequest)
	fmt.Println("===========================")
	out, err := base.ConvertToStructpb(prResp)
	if err != nil {
		return nil, err
	}
	return out, nil
}
