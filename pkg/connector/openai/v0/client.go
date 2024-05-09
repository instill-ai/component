package openai

import (
	"github.com/instill-ai/component/pkg/connector/util/httpclient"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func newClient(config *structpb.Struct, logger *zap.Logger) *httpclient.Client {
	c := httpclient.New("OpenAI", getBasePath(config),
		httpclient.WithLogger(logger),
		httpclient.WithEndUserError(new(errBody)),
	)

	c.SetAuthToken(getAPIKey(config))

	org := getOrg(config)
	if org != "" {
		c.SetHeader("OpenAI-Organization", org)
	}

	return c
}

type errBody struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (e errBody) Message() string {
	return e.Error.Message
}

// getBasePath returns OpenAI's API URL. This configuration param allows us to
// override the API the connector will point to. It isn't meant to be exposed
// to users. Rather, it can serve to test the logic against a fake server.
// TODO instead of having the API value hardcoded in the codebase, it should be
// read from a config file or environment variable.
func getBasePath(config *structpb.Struct) string {
	v, ok := config.GetFields()["base_path"]
	if !ok {
		return host
	}
	return v.GetStringValue()
}

func getAPIKey(config *structpb.Struct) string {
	return config.GetFields()[cfgAPIKey].GetStringValue()
}

func getOrg(config *structpb.Struct) string {
	val, ok := config.GetFields()[cfgOrganization]
	if !ok {
		return ""
	}
	return val.GetStringValue()
}
