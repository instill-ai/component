package sql

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

type InsertInput struct {
	Data      map[string]any `json:"data"`
	TableName string         `json:"table-name"`
}

type InsertOutput struct {
	Status string `json:"status"`
}

type UpdateInput struct {
	Update    map[string]any `json:"update"`
	Criteria  map[string]any `json:"criteria"`
	TableName string         `json:"table-name"`
}

type UpdateOutput struct {
	Status string `json:"status"`
}

type SelectInput struct {
	Criteria  map[string]any `json:"criteria"`
	TableName string         `json:"table-name"`
	From      int            `json:"from"`
	To        int            `json:"to"`
}

type SelectOutput struct {
	Rows   []map[string]any `json:"rows"`
	Status string           `json:"status"`
}

type DeleteInput struct {
	Criteria  map[string]any `json:"criteria"`
	TableName string         `json:"table-name"`
}

type DeleteOutput struct {
	Status string `json:"status"`
}

func buildSQLStatementInsert(tableName string, data *map[string]any) (string, map[string]any) {
	sqlStatement := "INSERT INTO " + tableName + " ("
	var columns []string
	var placeholders []string
	values := make(map[string]any)

	for dataKey, dataValue := range *data {
		columns = append(columns, dataKey)
		placeholders = append(placeholders, ":"+dataKey)
		values[dataKey] = dataValue
	}

	sqlStatement += strings.Join(columns, ", ") + ") VALUES (" + strings.Join(placeholders, ", ") + ")"

	return sqlStatement, values
}

func buildSQLStatementUpdate(tableName string, updateData map[string]interface{}, criteria map[string]interface{}, e execution) (string, map[string]interface{}) {
	// Take all columns
	sqlStatementCols := "SELECT * FROM " + tableName

	// Prepare and execute the statement
	rows, _ := e.client.Queryx(sqlStatementCols)
	defer rows.Close()

	sqlStatement := "UPDATE " + tableName + " SET "
	values := make(map[string]interface{})

	// Get column names from the query result
	columns, _ := rows.Columns()

	// Build SET clauses
	var setClauses []string
	for _, col := range columns {
		if updateValue, found := updateData[col]; found {
			setClauses = append(setClauses, fmt.Sprintf("%s = :%s", col, col))
			values[col] = updateValue
		} else {
			setClauses = append(setClauses, fmt.Sprintf("%s = NULL", col))
		}
	}

	sqlStatement += strings.Join(setClauses, ", ")

	// Build WHERE clauses
	var whereClauses []string
	for col, criteriaValue := range criteria {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = :%s_criteria", col, col))
		values[col+"_criteria"] = criteriaValue
	}

	sqlStatement += " WHERE " + strings.Join(whereClauses, " AND ")

	return sqlStatement, values
}

func buildSQLStatementSelect(tableName string, criteria *map[string]any, to int, from int) string {
	// Begin constructing SQL statement
	sqlStatement := "SELECT "
	var where []string
	var columns []string

	for criteriaKey, criteriaValue := range *criteria {
		if criteriaValue != nil {
			switch criteriaValue.(type) {
			case string:
				// If the value is a string, quote it in the SQL statement
				where = append(where, fmt.Sprintf("%s = '%v'", criteriaKey, criteriaValue))
			case map[string]any:
				// If the value is json, quote it in the SQL statement
				where = append(where, fmt.Sprintf("%s = '%v'", criteriaKey, criteriaValue))
			default:
				// If the value is a number or bool, use it directly without quotes
				where = append(where, fmt.Sprintf("%s = %v", criteriaKey, criteriaValue))
			}
		}

		columns = append(columns, criteriaKey)
	}

	var notAll string
	if to == 0 && from == 0 {
		notAll = ""
	} else {
		notAll = " LIMIT " + strconv.Itoa(to-from+1) + " OFFSET " + strconv.Itoa(from-1)
	}

	if len(columns) > 0 {
		sqlStatement += strings.Join(columns, ", ")
	} else {
		sqlStatement += "*"
	}

	sqlStatement += " FROM " + tableName
	if len(where) > 0 {
		sqlStatement += " WHERE " + strings.Join(where, " AND ")
	}
	sqlStatement += notAll

	return sqlStatement
}

func buildSQLStatementDelete(tableName string, criteria *map[string]any) (string, map[string]any) {
	// Begin constructing SQL statement
	sqlStatement := "DELETE FROM " + tableName + " WHERE "
	var where []string
	values := make(map[string]any) // Initialize the map

	for criteriaKey, criteriaValue := range *criteria {
		where = append(where, fmt.Sprintf("%s = :%s", criteriaKey, criteriaKey))
		values[criteriaKey] = criteriaValue
	}

	sqlStatement += strings.Join(where, " AND ")

	return sqlStatement, values
}

func (e *execution) insert(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct InsertInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	sqlStatement, values := buildSQLStatementInsert(inputStruct.TableName, &inputStruct.Data)

	// Prepare and execute the statement using NamedExec
	_, err = e.client.NamedExec(sqlStatement, values)

	if err != nil {
		return nil, err
	}

	outputStruct := InsertOutput{
		Status: "Successfully inserted rows",
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (e *execution) update(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct UpdateInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	sqlStatement, values := buildSQLStatementUpdate(inputStruct.TableName, inputStruct.Update, inputStruct.Criteria, *e)

	// Prepare and execute the statement using NamedExec
	_, err = e.client.NamedExec(sqlStatement, values)

	if err != nil {
		return nil, err
	}

	outputStruct := UpdateOutput{
		Status: "Successfully updated rows",
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (e *execution) selects(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct SelectInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	sqlStatement := buildSQLStatementSelect(inputStruct.TableName, &inputStruct.Criteria, inputStruct.To, inputStruct.From)

	// Prepare and execute the statement
	rows, err := e.client.Queryx(sqlStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Prepare the result slice of maps
	var result []map[string]any

	// Iterate over the rows
	for rows.Next() {
		// Create a map to hold the row data
		rowMap := make(map[string]any)

		// Load the row data into the map
		err := rows.MapScan(rowMap)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		// Convert each value in the map to the appropriate type
		for key, value := range rowMap {
			switch v := value.(type) {
			case []byte:
				// Convert byte slices to strings
				rowMap[key] = string(v)
			}
		}

		// Add the row map to the result slice
		result = append(result, rowMap)
	}

	outputStruct := SelectOutput{
		Rows:   result,
		Status: "Successfully selected rows",
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (e *execution) delete(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct DeleteInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	sqlStatement, values := buildSQLStatementDelete(inputStruct.TableName, &inputStruct.Criteria)

	// Prepare and execute the statement using NamedExec
	_, err = e.client.NamedExec(sqlStatement, values)

	if err != nil {
		return nil, err
	}

	outputStruct := DeleteOutput{
		Status: "Successfully deleted rows",
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil
}
