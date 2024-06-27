package cohere

import (
	"context"
	"fmt"

	cohereSDK "github.com/cohere-ai/cohere-go/v2"
	cohereClientSDK "github.com/cohere-ai/cohere-go/v2/client"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type cohereClient struct {
	sdkClient *cohereClientSDK.Client
	logger    *zap.Logger
}

func newClient(apiKey string, logger *zap.Logger) *cohereClient {
	fmt.Printf("### API key: %s.\n", apiKey)
	client := cohereClientSDK.NewClient(cohereClientSDK.WithToken(apiKey))
	return &cohereClient{sdkClient: client, logger: logger}
}

func (cl *cohereClient) generateTextChat(request cohereSDK.ChatRequest) (*cohereSDK.NonStreamedChatResponse, error) {
	fmt.Println("### Generating text chat")

	resp, err := cl.sdkClient.Chat(
		context.TODO(),
		&request,
	)
	if err != nil {
		fmt.Println("### An error has occurred")
		return resp, err
	}
	fmt.Printf("### Text chat generated: %s\n", resp.Text)
	return resp, nil
}

func getAPIKey(setup *structpb.Struct) string {
	return setup.GetFields()[cfgAPIKey].GetStringValue()
}
