package instill

import (
	"context"
	"crypto/tls"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

const maxPayloadSize int = 1024 * 1024 * 32

// initModelPublicServiceClient initialises a ModelPublicServiceClient instance
func initModelPublicServiceClient(serverURL string) (modelPB.ModelPublicServiceClient, *grpc.ClientConn) {
	var clientDialOpts grpc.DialOption

	if strings.HasPrefix(serverURL, "https://") {
		clientDialOpts = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true}))
	} else {
		clientDialOpts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	serverURL = stripProtocolFromURL(serverURL)
	clientConn, err := grpc.NewClient(serverURL, clientDialOpts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxPayloadSize), grpc.MaxCallSendMsgSize(maxPayloadSize)))
	if err != nil {
		return nil, nil
	}

	return modelPB.NewModelPublicServiceClient(clientConn), clientConn
}

func stripProtocolFromURL(url string) string {
	index := strings.Index(url, "://")
	if index > 0 {
		return url[strings.Index(url, "://")+3:]
	}
	return url
}

func trigger(gRPCClient modelPB.ModelPublicServiceClient, vars map[string]any, nsID string, modelID string, version string, taskInputs []*modelPB.TaskInput) ([]*modelPB.TaskOutput, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	ctx = metadata.NewOutgoingContext(ctx, getRequestMetadata(vars))

	res, err := gRPCClient.TriggerNamespaceModel(ctx, &modelPB.TriggerNamespaceModelRequest{
		NamespaceId: nsID,
		ModelId:     modelID,
		Version:     version,
		TaskInputs:  taskInputs,
	})
	if err != nil || res == nil {
		return nil, err
	}
	return res.TaskOutputs, nil
}
