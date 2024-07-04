package github

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v62/github"
	"github.com/instill-ai/x/errmsg"
	"golang.org/x/oauth2"
	"google.golang.org/protobuf/types/known/structpb"
)

type Client struct {
	client     GitHubClient
	owner      string
	repository string
}
type GitHubClient struct {
	*github.Client
	Repositories RepositoriesService
	PullRequests PullRequestService
	Issues       IssuesService
}

func newClient(ctx context.Context, setup *structpb.Struct) Client {
	token := getToken(setup)

	var oauth2Client *http.Client
	if token != "" {
		tokenSource := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		oauth2Client = oauth2.NewClient(ctx, tokenSource)
	}
	client := github.NewClient(oauth2Client)
	githubClient := Client{
		client: GitHubClient{
			Client:       client,
			Repositories: client.Repositories,
			PullRequests: client.PullRequests,
			Issues:       client.Issues,
		},
	}
	return githubClient
}

func (githubClient *Client) setTargetRepo(setup *structpb.Struct) error {
	githubClient.owner = setup.GetFields()["owner"].GetStringValue()
	githubClient.repository = setup.GetFields()["repository"].GetStringValue()

	if githubClient.owner == "" || githubClient.repository == "" {
		return errmsg.AddMessage(
			fmt.Errorf("owner or repository not provided"),
			"owner or repository not provided",
		)
	}
	return nil
}

func getToken(setup *structpb.Struct) string {
	return setup.GetFields()["token"].GetStringValue()
}
