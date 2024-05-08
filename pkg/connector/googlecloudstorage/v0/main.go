//go:generate compogen readme --connector ./config ./README.mdx
package googlecloudstorage

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
)

const (
	taskUpload = "TASK_UPLOAD"
)

//go:embed config/definition.json
var definitionJSON []byte

//go:embed config/tasks.json
var tasksJSON []byte
var once sync.Once
var con *connector

type connector struct {
	base.BaseConnector
}

type execution struct {
	base.BaseConnectorExecution
}

func Init(bc base.BaseConnector) *connector {
	once.Do(func() {
		con = &connector{BaseConnector: bc}
		err := con.LoadConnectorDefinition(definitionJSON, tasksJSON, nil)
		if err != nil {
			panic(err)
		}
	})
	return con
}

func (c *connector) CreateExecution(sysVars map[string]any, connection *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	return &base.ExecutionWrapper{Execution: &execution{
		BaseConnectorExecution: base.BaseConnectorExecution{Connector: c, SystemVariables: sysVars, Connection: connection, Task: task},
	}}, nil
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

func (e *execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := []*structpb.Struct{}

	client, err := NewClient(getJSONKey(e.Connection))
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
			bucketName := getBucketName(e.Connection)
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

func (c *connector) Test(sysVars map[string]any, connection *structpb.Struct) error {

	client, err := NewClient(getJSONKey(connection))
	if err != nil {
		return fmt.Errorf("error creating GCS client: %v", err)
	}
	if client == nil {
		return fmt.Errorf("GCS client is nil")
	}
	defer client.Close()
	return nil
}
