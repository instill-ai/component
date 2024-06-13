//go:generate compogen readme ./config ./README.mdx
package bigquery

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"sync"

	"cloud.google.com/go/bigquery"
	pb "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
)

const (
	taskInsert = "TASK_INSERT"
	taskRead   = "TASK_READ"
)

var instillUpstreamTypes = []string{"value", "reference", "template"}

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
			fmt.Println("input", input)
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

type TableColumns struct {
	TableName string
	Columns   []Column
}

type Column struct {
	Name string
	Type string
}

func (c *component) GetDefinition(sysVars map[string]any, compConfig *base.ComponentConfig) (*pb.ComponentDefinition, error) {

	ctx := context.Background()
	oriDef, err := c.Component.GetDefinition(nil, nil)
	if err != nil {
		return nil, err
	}

	if sysVars == nil && compConfig == nil {
		return oriDef, nil
	}

	def := proto.Clone(oriDef).(*pb.ComponentDefinition)
	client, err := NewClient(compConfig.Setup["json_key"].(string), compConfig.Setup["project_id"].(string))
	if err != nil || client == nil {
		return nil, fmt.Errorf("error creating BigQuery client: %v", err)
	}
	defer client.Close()

	myDataset := client.Dataset(compConfig.Setup["dataset_id"].(string))
	tables, err := constructTableColumns(myDataset, ctx, compConfig)
	if err != nil {
		return nil, err
	}

	tableProperties, err := constructTableProperties(tables)
	if err != nil {
		return nil, err
	}

	// TODO: chuang8511, remove table from definition.json and make it dynamic.
	// It will be changed before 2024-06-26.
	tableProperty := tableProperties[0]
	for _, sch := range def.Spec.ComponentSpecification.Fields["oneOf"].GetListValue().Values {
		data := sch.GetStructValue().Fields["properties"].GetStructValue().Fields["input"].GetStructValue().Fields["properties"].GetStructValue().Fields["data"].GetStructValue()
		if data != nil {
			data.Fields["properties"] = structpb.NewStructValue(tableProperty)
		}
	}

	for _, dataSpec := range def.Spec.DataSpecifications {
		dataInput := dataSpec.Input.Fields["properties"].GetStructValue().Fields["data"].GetStructValue()
		if dataInput != nil {
			dataInput.Fields["properties"] = structpb.NewStructValue(tableProperty)
		}
		dataOutput := dataSpec.Output.Fields["properties"].GetStructValue().Fields["data"].GetStructValue()

		if dataOutput != nil {
			aPieceData := dataOutput.Fields["items"].GetStructValue()
			if aPieceData != nil {
				aPieceData.Fields["properties"] = structpb.NewStructValue(tableProperty)
			}

		}
	}

	return def, nil
}

func constructTableColumns(myDataset *bigquery.Dataset, ctx context.Context, compConfig *base.ComponentConfig) ([]TableColumns, error) {
	tableIT := myDataset.Tables(ctx)
	tables := []TableColumns{}
	for {
		table, err := tableIT.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		tableName := table.TableID
		tableDetail := myDataset.Table(tableName)
		metadata, err := tableDetail.Metadata(ctx)
		if err != nil {
			return nil, err
		}
		schema := metadata.Schema
		columns := []Column{}
		for _, field := range schema {
			columns = append(columns, Column{Name: field.Name, Type: string(field.Type)})
		}

		// TODO: chuang8511, remove table from definition.json and make it dynamic.
		if compConfig.Setup["table_name"].(string) == tableName {
			tables = append(tables, TableColumns{TableName: tableName, Columns: columns})
		}
	}
	if len(tables) == 0 {
		return nil, fmt.Errorf("table name is not found in the dataset")
	}
	return tables, nil
}

func constructTableProperties(tables []TableColumns) ([]*structpb.Struct, error) {
	tableProperties := make([]*structpb.Struct, len(tables))
	transformer := base.InstillDynamicFormatTransformer{}

	for idx, table := range tables {
		propertiesMap := make(map[string]interface{})
		for idx, column := range table.Columns {
			kebabCaseColumnName := transformer.ConvertToKebab(column.Name)
			propertiesMap[kebabCaseColumnName] = map[string]interface{}{
				"title":                column.Name,
				"instillUIOrder":       idx,
				"description":          "Column " + column.Name + " of table " + table.TableName,
				"instillFormat":        getInstillAcceptFormat(column.Type),
				"instillUpstreamTypes": instillUpstreamTypes,
				"required":             []string{},
				"type":                 getInstillAcceptFormat(column.Type),
			}
		}
		propertyStructPB, err := base.ConvertToStructpb(propertiesMap)
		if err != nil {
			return nil, err
		}

		tableProperties[idx] = propertyStructPB
	}
	return tableProperties, nil
}

func getInstillAcceptFormat(tableType string) string {
	switch tableType {
	case "STRING":
		return "string"
	case "INTEGER":
		return "integer"
	case "BOOLEAN":
		return "boolean"
	default:
		return "string"
	}
}
