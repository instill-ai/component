package stabilityai

import (
	"github.com/instill-ai/component/internal/util/httpclient"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func newClient(config *structpb.Struct, logger *zap.Logger) *httpclient.Client {
	c := httpclient.New("Stability AI", getBasePath(config),
		httpclient.WithLogger(logger),
		httpclient.WithEndUserError(new(errBody)),
	)

	c.SetAuthToken(getAPIKey(config))

	return c
}

type errBody struct {
	Msg string `json:"message"`
}

func (e errBody) Message() string {
	return e.Msg
}

// getBasePath returns Stability AI's API URL. This configuration param allows
// us to override the API the connector will point to. It isn't meant to be
// exposed to users. Rather, it can serve to test the logic against a fake
// server.
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
