package github

import (
	"context"

	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
	"google.golang.org/protobuf/types/known/structpb"
)

func newClient(ctx context.Context, setup *structpb.Struct) *github.Client {
	token := getToken(setup)
	if token == "" {
		// no token provided, use unauthenticated client
		return github.NewClient(nil)
	}
	// create a new client with the provided token
	tokenSource := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: token},
    )
    oauth2Client := oauth2.NewClient(ctx, tokenSource)
	return github.NewClient(oauth2Client)
}

// Need to confirm where the map is
func getToken(setup *structpb.Struct) string {
	return setup.GetFields()["token"].GetStringValue()
}
