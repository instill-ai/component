package bigquery

import (
	"context"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

type ReadInput struct {
	ProjectID string
	DatasetID string
	TableName string
	Client    *bigquery.Client
	// If SelectColumns is empty, all columns will be selected
	SelectColumns  []string
	QueryParameter map[string]any
}

type ReadOutput struct {
	Data []map[string]any
}

func queryBuilder(input ReadInput) string {
	sql := "SELECT "

	if len(input.SelectColumns) == 0 {
		sql = sql + "*"
	}

	for idx, column := range input.SelectColumns {
		sql += column
		if idx < len(input.SelectColumns)-1 {
			sql += ", "
		}
	}

	sql += " FROM " + input.ProjectID + "." + input.DatasetID + "." + input.TableName

	keys := make([]string, len(input.QueryParameter))
	for k := range input.QueryParameter {
		keys = append(keys, k)
	}

	if len(input.QueryParameter) > 0 {
		sql += " WHERE "
		for idx, key := range keys {
			sql += key + " = @" + key
			if idx < len(keys)-1 {
				sql += " AND "
			}
		}
	}

	return sql
}

func readDataFromBigQuery(input ReadInput) (ReadOutput, error) {

	ctx := context.Background()
	client := input.Client

	sql := queryBuilder(input)
	q := client.Query(sql)
	var queryParameter []bigquery.QueryParameter
	for key, value := range input.QueryParameter {
		queryParameter = append(queryParameter, bigquery.QueryParameter{Name: key, Value: value})
	}

	q.Parameters = queryParameter
	it, err := q.Read(ctx)
	if err != nil {
		return ReadOutput{}, err
	}

	result := []map[string]any{}
	for {
		var values []bigquery.Value
		err := it.Next(&values)

		if err == nil {
			data := map[string]any{}
			for idx, value := range values {
				data[input.SelectColumns[idx]] = value
			}
			result = append(result, data)
		}

		if err == iterator.Done {
			break
		}
		if err != nil {
			return ReadOutput{}, err
		}

	}

	return ReadOutput{Data: result}, nil
}
