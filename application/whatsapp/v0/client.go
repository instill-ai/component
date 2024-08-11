package whatsapp

import (
	"github.com/instill-ai/component/internal/util/httpclient"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func newClient(setup *structpb.Struct, logger *zap.Logger) *WhatsAppClient {
	c := httpclient.New("WhatsApp", basePath+"/"+version,
		httpclient.WithLogger(logger),
		httpclient.WithEndUserError(new(errBody)),
	)

	c.SetAuthToken(getToken(setup))

	w := &WhatsAppClient{httpclient: c}

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

type WhatsAppClient struct {
	httpclient *httpclient.Client
}

// api functions

type WhatsAppInterface interface {
	SendMessageAPI(req interface{}, res interface{}, PhoneNumberID string) error
}

func (c *WhatsAppClient) SendMessageAPI(req interface{}, resp interface{}, PhoneNumberID string) error {
	httpReq := c.httpclient.R().SetBody(req).SetResult(resp)
	if _, err := httpReq.Post("/" + PhoneNumberID + "/messages"); err != nil {
		return err
	}
	return nil
}
