package mistral

import (
	mistralSDK "github.com/gage-technologies/mistral-go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type MistralClient struct {
	sdkClient *mistralSDK.MistralClient
	logger    *zap.Logger
}

func newClient(apiKey string, logger *zap.Logger) MistralClient {
	client := mistralSDK.NewMistralClientDefault(apiKey)
	return MistralClient{sdkClient: client, logger: logger}
}

func getAPIKey(setup *structpb.Struct) string {
	return setup.GetFields()[cfgAPIKey].GetStringValue()
}
