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

	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
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

func initMgmtPublicServiceClient(serverURL string) (mgmtPB.MgmtPublicServiceClient, *grpc.ClientConn) {
	var clientDialOpts grpc.DialOption

	if strings.HasPrefix(serverURL, "https://") {
		clientDialOpts = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true}))
	} else {
		clientDialOpts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	serverURL = stripProtocolFromURL(serverURL)
	clientConn, err := grpc.NewClient(serverURL, clientDialOpts)
	if err != nil {
		return nil, nil
	}

	return mgmtPB.NewMgmtPublicServiceClient(clientConn), clientConn
}

func stripProtocolFromURL(url string) string {
	index := strings.Index(url, "://")
	if index > 0 {
		return url[strings.Index(url, "://")+3:]
	}
	return url
}

func trigger(gRPCClient modelPB.ModelPublicServiceClient, vars map[string]any, modelName string, taskInputs []*modelPB.TaskInput) ([]*modelPB.TaskOutput, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	ctx = metadata.NewOutgoingContext(ctx, getRequestMetadata(vars))

	nameSplits := strings.Split(modelName, "/")

	if strings.HasPrefix(modelName, "user") {
		req := modelPB.TriggerUserModelRequest{
			Name:       strings.Join(nameSplits[0:4], "/"),
			TaskInputs: taskInputs,
			Version:    nameSplits[5],
		}

		res, err := gRPCClient.TriggerUserModel(ctx, &req)
		if err != nil || res == nil {
			return nil, err
		}
		return res.TaskOutputs, nil
	} else {
		req := modelPB.TriggerOrganizationModelRequest{
			Name:       strings.Join(nameSplits[0:4], "/"),
			TaskInputs: taskInputs,
			Version:    nameSplits[5],
		}

		res, err := gRPCClient.TriggerOrganizationModel(ctx, &req)
		if err != nil || res == nil {
			return nil, err
		}
		return res.TaskOutputs, nil
	}

}
