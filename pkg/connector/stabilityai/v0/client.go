package stabilityai

import (
	"github.com/instill-ai/component/pkg/connector/util/httpclient"
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
