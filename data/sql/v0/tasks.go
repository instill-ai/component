package sql

import (
	"fmt"
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
	Limit     int            `json:"limit"`
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

type CreateTableInput struct {
	TableName string            `json:"table-name"`
	Columns   map[string]string `json:"columns"`
}

type CreateTableOutput struct {
	Status string `json:"status"`
}

type DropTableInput struct {
	TableName string `json:"table-name"`
}

type DropTableOutput struct {
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
	sqlStatementCols := "SELECT * FROM " + tableName

	rows, _ := e.client.Queryx(sqlStatementCols)
	defer rows.Close()

	sqlStatement := "UPDATE " + tableName + " SET "
	values := make(map[string]interface{})

	columns, _ := rows.Columns()

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

	var whereClauses []string
	for col, criteriaValue := range criteria {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = :%s_criteria", col, col))
		values[col+"_criteria"] = criteriaValue
	}

	sqlStatement += " WHERE " + strings.Join(whereClauses, " AND ")

	return sqlStatement, values
}

// limit can be empty, but it will have default value 0
func buildSQLStatementSelect(tableName string, criteria *map[string]any, limit int) string {
	sqlStatement := "SELECT "
	var where []string
	var columns []string

	for criteriaKey, criteriaValue := range *criteria {
		if criteriaValue != nil {
			switch criteriaValue.(type) {
			case string:
				where = append(where, fmt.Sprintf("%s = '%v'", criteriaKey, criteriaValue))
			case map[string]any:
				where = append(where, fmt.Sprintf("%s = '%v'", criteriaKey, criteriaValue))
			default:
				where = append(where, fmt.Sprintf("%s = %v", criteriaKey, criteriaValue))
			}
		}

		columns = append(columns, criteriaKey)
	}

	var notAll string
	if limit == 0 {
		notAll = ""
	} else {
		notAll = fmt.Sprintf(" LIMIT %d", limit)
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
	sqlStatement := "DELETE FROM " + tableName + " WHERE "
	var where []string
	values := make(map[string]any)

	for criteriaKey, criteriaValue := range *criteria {
		where = append(where, fmt.Sprintf("%s = :%s", criteriaKey, criteriaKey))
		values[criteriaKey] = criteriaValue
	}

	sqlStatement += strings.Join(where, " AND ")

	return sqlStatement, values
}

// columns is a map of column name and column type and handled in json format to prevent sql injection
func buildSQLStatementCreateTable(tableName string, columns map[string]string) (string, map[string]any) {
	sqlStatement := "CREATE TABLE " + tableName + " ("
	var columnDefs []string
	values := make(map[string]any)

	for colName, colType := range columns {
		columnDefs = append(columnDefs, fmt.Sprintf("%s %s", colName, colType))
		values[colName] = colType
	}

	sqlStatement += strings.Join(columnDefs, ", ") + ");"
	return sqlStatement, values
}

func buildSQLStatementDropTable(tableName string) (string, map[string]any) {
	sqlStatement := "DROP TABLE " + tableName + ";"
	values := map[string]any{"table_name": tableName}
	return sqlStatement, values
}

func (e *execution) insert(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct InsertInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	sqlStatement, values := buildSQLStatementInsert(inputStruct.TableName, &inputStruct.Data)

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

// Queryx is used since we need not only status but also result return
func (e *execution) selects(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct SelectInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	sqlStatement := buildSQLStatementSelect(inputStruct.TableName, &inputStruct.Criteria, inputStruct.Limit)

	rows, err := e.client.Queryx(sqlStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]any

	for rows.Next() {
		rowMap := make(map[string]any)

		err := rows.MapScan(rowMap)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		for key, value := range rowMap {
			switch v := value.(type) {
			case []byte:
				rowMap[key] = string(v)
			}
		}

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

func (e *execution) createTable(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct CreateTableInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	sqlStatement, values := buildSQLStatementCreateTable(inputStruct.TableName, inputStruct.Columns)

	_, err = e.client.NamedExec(sqlStatement, values)

	if err != nil {
		return nil, err
	}

	outputStruct := CreateTableOutput{
		Status: "Successfully created table",
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (e *execution) dropTable(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct DropTableInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	sqlStatement, values := buildSQLStatementDropTable(inputStruct.TableName)

	_, err = e.client.NamedExec(sqlStatement, values)

	if err != nil {
		return nil, err
	}

	outputStruct := DropTableOutput{
		Status: "Successfully dropped table",
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil
}
