package slack

import (
	"github.com/slack-go/slack"
	"google.golang.org/protobuf/types/known/structpb"
)

func newClient(config *structpb.Struct) *slack.Client {
	return slack.New(getToken(config))
}

// Need to confirm where the map is
func getToken(config *structpb.Struct) string {
	return config.GetFields()["token"].GetStringValue()
}
