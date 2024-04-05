//go:generate compogen readme --connector ./config ./README.mdx
package googlecloudstorage

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"cloud.google.com/go/storage"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

const (
	taskUpload = "TASK_UPLOAD"
)

//go:embed config/definition.json
var definitionJSON []byte

//go:embed config/tasks.json
var tasksJSON []byte
var once sync.Once
var connector base.IConnector

type Connector struct {
	base.Connector
}

type Execution struct {
	base.Execution
}

func Init(logger *zap.Logger) base.IConnector {
	once.Do(func() {
		connector = &Connector{
			Connector: base.Connector{
				Component: base.Component{Logger: logger},
			},
		}
		err := connector.LoadConnectorDefinition(definitionJSON, tasksJSON, nil)
		if err != nil {
			logger.Fatal(err.Error())
		}
	})
	return connector
}

func (c *Connector) CreateExecution(defUID uuid.UUID, task string, config *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	e := &Execution{}
	e.Execution = base.CreateExecutionHelper(e, c, defUID, task, config, logger)
	return e, nil
}

func NewClient(jsonKey string) (*storage.Client, error) {
	return storage.NewClient(context.Background(), option.WithCredentialsJSON([]byte(jsonKey)))
}

func getBucketName(config *structpb.Struct) string {
	return config.GetFields()["bucket_name"].GetStringValue()
}

func getJSONKey(config *structpb.Struct) string {
	return config.GetFields()["json_key"].GetStringValue()
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := []*structpb.Struct{}

	client, err := NewClient(getJSONKey(e.Config))
	if err != nil || client == nil {
		return nil, fmt.Errorf("error creating GCS client: %v", err)
	}
	defer client.Close()
	for _, input := range inputs {
		var output *structpb.Struct
		switch e.Task {
		case taskUpload, "":
			objectName := input.GetFields()["object_name"].GetStringValue()
			data := input.GetFields()["data"].GetStringValue()
			bucketName := getBucketName(e.Config)
			err = uploadToGCS(client, bucketName, objectName, data)
			if err != nil {
				return nil, err
			}

			gsutilURI := fmt.Sprintf("gs://%s/%s", bucketName, objectName)
			authenticatedURL := fmt.Sprintf("https://storage.cloud.google.com/%s/%s?authuser=1", bucketName, objectName)
			publicURL := ""

			// Check whether the object is public or not
			publicAccess, err := isObjectPublic(client, bucketName, objectName)
			if err != nil {
				return nil, err
			}
			if publicAccess {
				publicURL = fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, objectName)
			}

			output = &structpb.Struct{Fields: map[string]*structpb.Value{
				"authenticated_url": {Kind: &structpb.Value_StringValue{StringValue: authenticatedURL}},
				"gsutil_uri":        {Kind: &structpb.Value_StringValue{StringValue: gsutilURI}},
				"public_url":        {Kind: &structpb.Value_StringValue{StringValue: publicURL}},
				"public_access":     {Kind: &structpb.Value_BoolValue{BoolValue: publicAccess}},
				"status":            {Kind: &structpb.Value_StringValue{StringValue: "success"}}}}
		}
		outputs = append(outputs, output)
	}
	return outputs, nil
}

func (c *Connector) Test(defUID uuid.UUID, config *structpb.Struct, logger *zap.Logger) (pipelinePB.Connector_State, error) {

	client, err := NewClient(getJSONKey(config))
	if err != nil {
		return pipelinePB.Connector_STATE_ERROR, fmt.Errorf("error creating GCS client: %v", err)
	}
	if client == nil {
		return pipelinePB.Connector_STATE_DISCONNECTED, fmt.Errorf("GCS client is nil")
	}
	defer client.Close()
	return pipelinePB.Connector_STATE_CONNECTED, nil
}
