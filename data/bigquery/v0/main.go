//go:generate compogen readme ./config ./README.mdx
package bigquery

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"sync"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
)

const (
	taskInsert = "TASK_INSERT"
	taskRead   = "TASK_READ"
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

func (c *component) CreateExecution(sysVars map[string]any, setup *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	return &base.ExecutionWrapper{Execution: &execution{
		ComponentExecution: base.ComponentExecution{Component: c, SystemVariables: sysVars, Setup: setup, Task: task},
	}}, nil
}

func NewClient(jsonKey, projectID string) (*bigquery.Client, error) {
	return bigquery.NewClient(context.Background(), projectID, option.WithCredentialsJSON([]byte(jsonKey)))
}

func getJSONKey(setup *structpb.Struct) string {
	return setup.GetFields()["json-key"].GetStringValue()
}
func getProjectID(setup *structpb.Struct) string {
	return setup.GetFields()["project-id"].GetStringValue()
}
func getDatasetID(setup *structpb.Struct) string {
	return setup.GetFields()["dataset-id"].GetStringValue()
}
func getTableName(setup *structpb.Struct) string {
	return setup.GetFields()["table-name"].GetStringValue()
}

func (e *execution) Execute(ctx context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := []*structpb.Struct{}

	client, err := NewClient(getJSONKey(e.Setup), getProjectID(e.Setup))
	if err != nil || client == nil {
		return nil, fmt.Errorf("error creating BigQuery client: %v", err)
	}
	defer client.Close()

	for _, input := range inputs {
		var output *structpb.Struct
		switch e.Task {
		case taskInsert, "":
			datasetID := getDatasetID(e.Setup)
			tableName := getTableName(e.Setup)
			tableRef := client.Dataset(datasetID).Table(tableName)
			metaData, err := tableRef.Metadata(context.Background())
			if err != nil {
				return nil, err
			}
			valueSaver, err := getDataSaver(input, metaData.Schema)
			if err != nil {
				return nil, err
			}
			err = insertDataToBigQuery(getProjectID(e.Setup), datasetID, tableName, valueSaver, client)
			if err != nil {
				return nil, err
			}
			output = &structpb.Struct{Fields: map[string]*structpb.Value{"status": {Kind: &structpb.Value_StringValue{StringValue: "success"}}}}
		case taskRead:

			inputStruct := ReadInput{
				ProjectID: getProjectID(e.Setup),
				DatasetID: getDatasetID(e.Setup),
				TableName: getTableName(e.Setup),
				Client:    client,
			}
			fmt.Println("inputStruct", inputStruct)
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			outputStruct, err := readDataFromBigQuery(inputStruct)
			if err != nil {
				return nil, err
			}
			output, err = base.ConvertToStructpb(outputStruct)
			if err != nil {
				return nil, err
			}

		default:
			return nil, fmt.Errorf("unsupported task: %s", e.Task)
		}
		outputs = append(outputs, output)
	}
	return outputs, nil
}

func (c *component) Test(sysVars map[string]any, setup *structpb.Struct) error {

	client, err := NewClient(getJSONKey(setup), getProjectID(setup))
	if err != nil || client == nil {
		return fmt.Errorf("error creating BigQuery client: %v", err)
	}
	defer client.Close()
	if client.Project() == getProjectID(setup) {
		return nil
	}
	return errors.New("project ID does not match")
}
