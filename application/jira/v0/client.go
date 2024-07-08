package jira

import (
	"context"
	"fmt"

	"github.com/andygrunwald/go-jira"
	"github.com/instill-ai/x/errmsg"
	"google.golang.org/protobuf/types/known/structpb"
)

type Client struct {
	*jira.Client
}

func newClient(_ context.Context, setup *structpb.Struct) (*Client, error) {
	token := getToken(setup)

	if token == "" {
		return nil, errmsg.AddMessage(
			fmt.Errorf("token not provided"),
			"token not provided",
		)
	}

	authTransport := jira.BearerAuthTransport{
		Token: token,
	}
	jiraClient, err := jira.NewClient(authTransport.Client(), baseURL)
	if err != nil {
		return nil, err
	}
	client := &Client{
		jiraClient,
	}
	return client, nil
}

func getToken(setup *structpb.Struct) string {
	return setup.GetFields()["token"].GetStringValue()
}
