package bigquery

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"sync"

	"cloud.google.com/go/bigquery"
	"go.uber.org/zap"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
)

const (
	taskInsert = "TASK_INSERT"
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

func Init(l *zap.Logger, u base.UsageHandler) *connector {
	once.Do(func() {
		con = &connector{
			BaseConnector: base.BaseConnector{
				Logger:       l,
				UsageHandler: u,
			},
		}
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

func NewClient(jsonKey, projectID string) (*bigquery.Client, error) {
	return bigquery.NewClient(context.Background(), projectID, option.WithCredentialsJSON([]byte(jsonKey)))
}

func getJSONKey(config *structpb.Struct) string {
	return config.GetFields()["json_key"].GetStringValue()
}
func getProjectID(config *structpb.Struct) string {
	return config.GetFields()["project_id"].GetStringValue()
}
func getDatasetID(config *structpb.Struct) string {
	return config.GetFields()["dataset_id"].GetStringValue()
}
func getTableName(config *structpb.Struct) string {
	return config.GetFields()["table_name"].GetStringValue()
}

func (e *execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := []*structpb.Struct{}

	client, err := NewClient(getJSONKey(e.Connection), getProjectID(e.Connection))
	if err != nil || client == nil {
		return nil, fmt.Errorf("error creating BigQuery client: %v", err)
	}
	defer client.Close()

	for _, input := range inputs {
		var output *structpb.Struct
		switch e.Task {
		case taskInsert, "":
			datasetID := getDatasetID(e.Connection)
			tableName := getTableName(e.Connection)
			tableRef := client.Dataset(datasetID).Table(tableName)
			metaData, err := tableRef.Metadata(context.Background())
			if err != nil {
				return nil, err
			}
			valueSaver, err := getDataSaver(input, metaData.Schema)
			if err != nil {
				return nil, err
			}
			err = insertDataToBigQuery(getProjectID(e.Connection), datasetID, tableName, valueSaver, client)
			if err != nil {
				return nil, err
			}
			output = &structpb.Struct{Fields: map[string]*structpb.Value{"status": {Kind: &structpb.Value_StringValue{StringValue: "success"}}}}
		default:
			return nil, fmt.Errorf("unsupported task: %s", e.Task)
		}
		outputs = append(outputs, output)
	}
	return outputs, nil
}

func (c *connector) Test(sysVars map[string]any, connection *structpb.Struct) error {

	client, err := NewClient(getJSONKey(connection), getProjectID(connection))
	if err != nil || client == nil {
		return fmt.Errorf("error creating BigQuery client: %v", err)
	}
	defer client.Close()
	if client.Project() == getProjectID(connection) {
		return nil
	}
	return errors.New("project ID does not match")
}
