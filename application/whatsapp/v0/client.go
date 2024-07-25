package whatsapp

import (
	"github.com/instill-ai/component/internal/util/httpclient"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func newClient(setup *structpb.Struct, logger *zap.Logger) *WhatsappClient {
	c := httpclient.New("WhatsApp", basePath+"/"+version,
		httpclient.WithLogger(logger),
		httpclient.WithEndUserError(new(errBody)),
	)

	c.SetAuthToken(getToken(setup))

	w := &WhatsappClient{httpclient: c}

	return w
}

type errBody struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (e errBody) Message() string {
	return e.Error.Message
}

func getToken(setup *structpb.Struct) string {
	return setup.GetFields()["token"].GetStringValue()
}

type WhatsappClient struct {
	httpclient *httpclient.Client
}

// api functions

type WhatsappInterface interface {
	SendMessageAPI(req interface{}, res interface{}, PhoneNumberId string) (interface{}, error)
}

func (c *WhatsappClient) SendMessageAPI(req interface{}, resp interface{}, PhoneNumberId string) (interface{}, error) {
	httpReq := c.httpclient.R().SetBody(req).SetResult(resp)
	if _, err := httpReq.Post("/" + PhoneNumberId + "/messages"); err != nil {
		return nil, err
	}
	return resp, nil
}
