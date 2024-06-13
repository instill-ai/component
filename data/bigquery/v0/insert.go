package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

type DataSaver struct {
	Schema  bigquery.Schema
	DataMap map[string]bigquery.Value
}

func (v DataSaver) Save() (row map[string]bigquery.Value, insertID string, err error) {
	return v.DataMap, bigquery.NoDedupeID, nil
}

func insertDataToBigQuery(projectID, datasetID, tableName string, valueSaver DataSaver, client *bigquery.Client) error {
	ctx := context.Background()
	tableRef := client.Dataset(datasetID).Table(tableName)
	inserter := tableRef.Inserter()
	if err := inserter.Put(ctx, valueSaver); err != nil {
		return fmt.Errorf("error inserting data: %v", err)
	}
	fmt.Printf("Data inserted into %s.%s.%s.\n", projectID, datasetID, tableName)
	return nil
}

func getDataSaver(input *structpb.Struct, schema bigquery.Schema) (DataSaver, error) {
	inputObj := input.GetFields()["data"].GetStructValue()
	dataMap := map[string]bigquery.Value{}
	transformer := base.InstillDynamicFormatTransformer{}
	for _, sc := range schema {
		kebabName := transformer.ConvertToKebab(sc.Name)
		dataMap[sc.Name] = inputObj.GetFields()[kebabName].AsInterface()
	}
	return DataSaver{Schema: schema, DataMap: dataMap}, nil
}
