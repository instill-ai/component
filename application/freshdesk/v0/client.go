package freshdesk

import (
	"github.com/instill-ai/component/internal/util/httpclient"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func newClient(setup *structpb.Struct, logger *zap.Logger) *FreshdeskClient {
	c := httpclient.New("Freshdesk", basePath+"/"+version,
		httpclient.WithLogger(logger),
		httpclient.WithEndUserError(new(errBody)),
	)

	c.SetAuthToken(getToken(setup))

	w := &FreshdeskClient{httpclient: c}

	return w
}

type errBody struct {
	Description string `json:"description"`
	Errors      []struct {
		Field   string `json:"field"`
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"errors"`
}

func (e errBody) Message() string {
	var errReturn string
	for index, err := range e.Errors {
		if index > 0 {
			errReturn += ", "
		}

		errReturn += err.Message
		if err.Field != "" {
			errReturn += " " + err.Field
		}
		if err.Code != "" {
			errReturn += " " + err.Code
		}
	}

	return errReturn
}

func getToken(setup *structpb.Struct) string {
	return setup.GetFields()["token"].GetStringValue()
}

type FreshdeskClient struct {
	httpclient *httpclient.Client
}

// api functions

type FreshdeskInterface interface {
}
