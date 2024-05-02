package slack

import (
	"fmt"

	"github.com/instill-ai/component/pkg/connector/util/httpclient"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	host            = "https://slack.com/api"
	channelListPath = "/conversations.list"
	sendMessagePath = "/chat.postMessage"
)

func newClient(config *structpb.Struct, logger *zap.Logger) *httpclient.Client {
	c := httpclient.New("Slack", getBasePath(config),
		httpclient.WithLogger(logger),
		httpclient.WithEndUserError(new(errBody)),
	)

	c.SetAuthToken(getAPIKey(config))

	return c
}

type errBody struct {
	Error string `json:"error"`
}

func (e errBody) Message() string {
	return e.Error
}

// Need to confirm where the map is
func getAPIKey(config *structpb.Struct) string {
	return config.GetFields()["api_key"].GetStringValue()
}

func getBasePath(config *structpb.Struct) string {
	v, ok := config.GetFields()["base_path"]
	if !ok {
		return host
	}
	return v.GetStringValue()
}

// TODO: to be refactor with DI for API calling part
func fetchChannelInfo(c *httpclient.Client) (*[]SlackChannel, error) {
	resp := SlackChannelApiResp{}
	req := c.R().SetResult(&resp)

	if _, err := req.Get(channelListPath); err != nil {
		return nil, err
	}

	return &resp.Channels, nil
}

func postMessageToSlackChannel(c *httpclient.Client, params SendingData) error {
	resp := SendingMessageResp{}
	req := c.R().SetBody(params).SetResult(&resp)

	if _, err := req.Post(sendMessagePath); err != nil {
		return err
	}

	if !resp.OK {
		if resp.Needed == "" {
			err := fmt.Errorf("fail to send message because of [%v]", resp.Error)
			return err
		} else {
			err := fmt.Errorf("fail to send message because you need [%v] for your scope", resp.Needed)
			return err
		}

	}

	return nil
}
