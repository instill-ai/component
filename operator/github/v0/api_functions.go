package github

import (
	"context"

	"github.com/google/go-github/v62/github"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)


type GetPullRequestsResp struct {
	PullRequests   []PullRequests `json:"pull-requests"`
}
type PullRequests struct {
	Title	 	string 			`string:"title"`
	Body 		string 			`string:"body"`
}
type GitHubClient struct {
	client *github.Client
	owner string
	repository string
}

func (githubClient *GitHubClient) setTargetRepo(setup *structpb.Struct) {
	githubClient.owner = setup.GetFields()["owner"].GetStringValue()
	githubClient.repository = setup.GetFields()["repository"].GetStringValue()
}
func (githubClient *GitHubClient) getPRs(props *structpb.Struct) (*structpb.Struct, error) {
	githubClient.setTargetRepo(props)
	prs, _, err := githubClient.client.PullRequests.List(context.Background(), githubClient.owner, githubClient.repository, nil)
	if err != nil {
		return nil, err
	}

	var prResp GetPullRequestsResp
	prResp.PullRequests = make([]PullRequests, len(prs))
	for idx, pr := range prs {
		prResp.PullRequests[idx].Title = pr.GetTitle()
		prResp.PullRequests[idx].Body = pr.GetBody()
	}

	out, err := base.ConvertToStructpb(prResp)
	if err != nil {
		return nil, err
	}
	return out, nil
}
