package bigquery

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"sync"

	"cloud.google.com/go/bigquery"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

const (
	taskInsert = "TASK_INSERT"
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

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := []*structpb.Struct{}

	client, err := NewClient(getJSONKey(e.Config), getProjectID(e.Config))
	if err != nil || client == nil {
		return nil, fmt.Errorf("error creating BigQuery client: %v", err)
	}
	defer client.Close()

	for _, input := range inputs {
		var output *structpb.Struct
		switch e.Task {
		case taskInsert, "":
			datasetID := getDatasetID(e.Config)
			tableName := getTableName(e.Config)
			tableRef := client.Dataset(datasetID).Table(tableName)
			metaData, err := tableRef.Metadata(context.Background())
			if err != nil {
				return nil, err
			}
			valueSaver, err := getDataSaver(input, metaData.Schema)
			if err != nil {
				return nil, err
			}
			err = insertDataToBigQuery(getProjectID(e.Config), datasetID, tableName, valueSaver, client)
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

func (c *Connector) Test(defUID uuid.UUID, config *structpb.Struct, logger *zap.Logger) (pipelinePB.Connector_State, error) {

	client, err := NewClient(getJSONKey(config), getProjectID(config))
	if err != nil || client == nil {
		return pipelinePB.Connector_STATE_ERROR, fmt.Errorf("error creating BigQuery client: %v", err)
	}
	defer client.Close()
	if client.Project() == getProjectID(config) {
		return pipelinePB.Connector_STATE_CONNECTED, nil
	}
	return pipelinePB.Connector_STATE_DISCONNECTED, errors.New("project ID does not match")
}
