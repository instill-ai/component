package github

import (
	"fmt"

	"github.com/google/go-github/v62/github"
	"github.com/instill-ai/x/errmsg"
	"google.golang.org/protobuf/types/known/structpb"
)


type GitHubClient struct {
	client *github.Client
	owner string
	repository string
}

func (githubClient *GitHubClient) setTargetRepo(setup *structpb.Struct) error {
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
