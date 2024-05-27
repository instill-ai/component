package restapi

import (
	"github.com/instill-ai/component/pkg/util/httpclient"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type TaskInput struct {
	EndpointURL string      `json:"endpoint_url"`
	Body        interface{} `json:"body,omitempty"`
}

type TaskOutput struct {
	StatusCode int                 `json:"status_code"`
	Body       interface{}         `json:"body"`
	Header     map[string][]string `json:"header"`
}

func newClient(config *structpb.Struct, logger *zap.Logger) (*httpclient.Client, error) {
	c := httpclient.New("REST API", "",
		httpclient.WithLogger(logger),
	)

	auth, err := getAuthentication(config)
	if err != nil {
		return nil, err
	}

	if err := auth.setAuthInClient(c); err != nil {
		return nil, err
	}

	return c, nil
}
