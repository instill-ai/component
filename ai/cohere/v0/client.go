package cohere

import (
	"context"
	"fmt"
	"sync"

	cohereSDK "github.com/cohere-ai/cohere-go/v2"
	cohereClientSDK "github.com/cohere-ai/cohere-go/v2/client"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type cohereClient struct {
	sdkClient *cohereClientSDK.Client
	logger    *zap.Logger
	lock      sync.Mutex
}

func newClient(apiKey string, logger *zap.Logger) *cohereClient {
	fmt.Printf("### API key: %s\n", apiKey)
	client := cohereClientSDK.NewClient(cohereClientSDK.WithToken(apiKey))
	return &cohereClient{sdkClient: client, logger: logger, lock: sync.Mutex{}}
}

func (cl *cohereClient) generateTextChat(request cohereSDK.ChatRequest) (cohereSDK.NonStreamedChatResponse, error) {
	respPtr, err := cl.sdkClient.Chat(
		context.TODO(),
		&request,
	)
	if err != nil {
		panic(err)
	}
	resp := cohereSDK.NonStreamedChatResponse{
		Text:         respPtr.Text,
		GenerationId: respPtr.GenerationId,
		Citations:    respPtr.Citations,
	}
	return resp, nil
}

func getAPIKey(setup *structpb.Struct) string {
	return setup.GetFields()[cfgAPIKey].GetStringValue()
}
