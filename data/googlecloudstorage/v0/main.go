//go:generate compogen readme ./config ./README.mdx
package googlecloudstorage

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
)

const (
	taskUpload       = "TASK_UPLOAD"
	taskReadObjects  = "TASK_READ_OBJECTS"
	taskCreateBucket = "TASK_CREATE_BUCKET"
)

//go:embed config/definition.json
var definitionJSON []byte

//go:embed config/setup.json
var setupJSON []byte

//go:embed config/tasks.json
var tasksJSON []byte

var once sync.Once
var comp *component

type component struct {
	base.Component
}

type execution struct {
	base.ComponentExecution
}

func Init(bc base.Component) *component {
	once.Do(func() {
		comp = &component{Component: bc}
		err := comp.LoadDefinition(definitionJSON, setupJSON, tasksJSON, nil)
		if err != nil {
			panic(err)
		}
	})
	return comp
}

func (c *component) CreateExecution(x base.ComponentExecution) (base.IExecution, error) {
	return &execution{
		ComponentExecution: x,
	}, nil
}

func NewClient(jsonKey string) (*storage.Client, error) {
	return storage.NewClient(context.Background(), option.WithCredentialsJSON([]byte(jsonKey)))
}

func getJSONKey(setup *structpb.Struct) string {
	return setup.GetFields()["json-key"].GetStringValue()
}

func (e *execution) Execute(ctx context.Context, in base.InputReader, out base.OutputWriter) error {
	inputs, err := in.Read(ctx)
	if err != nil {
		return err
	}
	outputs := []*structpb.Struct{}

	client, err := NewClient(getJSONKey(e.Setup))
	if err != nil || client == nil {
		return fmt.Errorf("error creating GCS client: %v", err)
	}
	defer client.Close()
	for _, input := range inputs {
		var output *structpb.Struct
		bucketName := input.GetFields()["bucket-name"].GetStringValue()
		switch e.Task {
		case taskUpload, "":
			objectName := input.GetFields()["object-name"].GetStringValue()
			data := input.GetFields()["data"].GetStringValue()
			err = uploadToGCS(client, bucketName, objectName, data)
			if err != nil {
				return err
			}

			gsutilURI := fmt.Sprintf("gs://%s/%s", bucketName, objectName)
			authenticatedURL := fmt.Sprintf("https://storage.cloud.google.com/%s/%s?authuser=1", bucketName, objectName)
			publicURL := ""

			// Check whether the object is public or not
			publicAccess, err := isObjectPublic(client, bucketName, objectName)
			if err != nil {
				return err
			}
			if publicAccess {
				publicURL = fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, objectName)
			}

			output = &structpb.Struct{Fields: map[string]*structpb.Value{
				"authenticated-url": {Kind: &structpb.Value_StringValue{StringValue: authenticatedURL}},
				"gsutil-uri":        {Kind: &structpb.Value_StringValue{StringValue: gsutilURI}},
				"public-url":        {Kind: &structpb.Value_StringValue{StringValue: publicURL}},
				"public-access":     {Kind: &structpb.Value_BoolValue{BoolValue: publicAccess}},
				"status":            {Kind: &structpb.Value_StringValue{StringValue: "success"}}}}

		case taskReadObjects:
			inputStruct := ReadInput{
				BucketName: bucketName,
			}

			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return err
			}
			outputStruct, err := readObjects(inputStruct, client, ctx)
			if err != nil {
				return err
			}
			output, err = base.ConvertToStructpb(outputStruct)

			if err != nil {
				return err
			}

		case taskCreateBucket:
			inputStruct := CreateBucketInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return err
			}

			outputStruct, err := createBucket(inputStruct, client, ctx)
			if err != nil {
				return err
			}

			output, err = base.ConvertToStructpb(outputStruct)
			if err != nil {
				return err
			}
		}

		outputs = append(outputs, output)
	}
	return out.Write(ctx, outputs)
}

func (c *component) Test(sysVars map[string]any, setup *structpb.Struct) error {

	client, err := NewClient(getJSONKey(setup))
	if err != nil {
		return fmt.Errorf("error creating GCS client: %v", err)
	}
	if client == nil {
		return fmt.Errorf("GCS client is nil")
	}
	defer client.Close()
	return nil
}
