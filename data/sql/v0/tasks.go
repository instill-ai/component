package sql

import (
	"fmt"
	"regexp"
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
	UpdateData map[string]any `json:"update-data"`
	Filter     string         `json:"filter"`
	TableName  string         `json:"table-name"`
}

type UpdateOutput struct {
	Status string `json:"status"`
}

type SelectInput struct {
	Filter    string   `json:"filter"`
	TableName string   `json:"table-name"`
	Limit     int      `json:"limit"`
	Columns   []string `json:"columns"`
}

type SelectOutput struct {
	Rows   []map[string]any `json:"rows"`
	Status string           `json:"status"`
}

type DeleteInput struct {
	Filter    string `json:"filter"`
	TableName string `json:"table-name"`
}

type DeleteOutput struct {
	Status string `json:"status"`
}

type CreateTableInput struct {
	TableName        string            `json:"table-name"`
	ColumnsStructure map[string]string `json:"columns-structure"`
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

func isValidWhereClause(whereClause string) error {
	// Extended regex pattern for logical operators and additional conditions
	regex := `^(?:\w+ (?:=|!=|>|<|>=|<=|LIKE|MATCH|IS NULL|IS NOT NULL|BETWEEN|IN|EXISTS|NOT|REGEXP|RLIKE|IS DISTINCT FROM|IS NOT DISTINCT FROM|COALESCE\(.*\)|NULLIF\(.*\)) (?:[\w'%]+|\d+|\([\w\s,']+\)|(?:CASE .* END)|(?:\w+\s+\w+)))(?: (?:AND|OR) (?:\w+ (?:=|!=|>|<|>=|<=|LIKE|MATCH|IS NULL|IS NOT NULL|BETWEEN|IN|EXISTS|NOT|REGEXP|RLIKE|IS DISTINCT FROM|IS NOT DISTINCT FROM|COALESCE\(.*\)|NULLIF\(.*\)) (?:[\w'%]+|\d+|\([\w\s,']+\)|(?:CASE .* END)|(?:\w+\s+\w+))))*$`
	matched, err := regexp.MatchString(regex, whereClause)
	if err != nil || !matched {
		return err
	}
	return nil
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

func buildSQLStatementUpdate(tableName string, updateData map[string]any, filter string) (string, map[string]any) {
	sqlStatement := "UPDATE " + tableName + " SET "
	values := make(map[string]any)

	var setClauses []string
	for col, updateValue := range updateData {
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", col, col))
		values[col] = updateValue
	}

	sqlStatement += strings.Join(setClauses, ", ")

	if filter != "" {
		sqlStatement += " WHERE " + filter
	}

	return sqlStatement, values
}

// limit can be empty, but it will have default value 0
// columns can be empty, if empty it will select all columns
func buildSQLStatementSelect(tableName string, filter string, limit int, columns []string) string {
	sqlStatement := "SELECT "

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
	if filter != "" {
		sqlStatement += " WHERE " + filter
	}
	sqlStatement += notAll

	return sqlStatement
}

func buildSQLStatementDelete(tableName string, filter string) string {
	sqlStatement := "DELETE FROM " + tableName

	if filter != "" {
		sqlStatement += " WHERE " + filter
	}

	return sqlStatement
}

// columns is a map of column name and column type and handled in json format to prevent sql injection
func buildSQLStatementCreateTable(tableName string, columnsStructure map[string]string) (string, map[string]any) {
	sqlStatement := "CREATE TABLE " + tableName + " ("
	var columnDefs []string
	values := make(map[string]any)

	for colName, colType := range columnsStructure {
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
	err = isValidWhereClause(inputStruct.Filter)
	if err != nil {
		return nil, err
	}

	sqlStatement, values := buildSQLStatementUpdate(inputStruct.TableName, inputStruct.UpdateData, inputStruct.Filter)

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
	err = isValidWhereClause(inputStruct.Filter)
	if err != nil {
		return nil, err
	}

	sqlStatement := buildSQLStatementSelect(inputStruct.TableName, inputStruct.Filter, inputStruct.Limit, inputStruct.Columns)

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
	err = isValidWhereClause(inputStruct.Filter)
	if err != nil {
		return nil, err
	}

	sqlStatement := buildSQLStatementDelete(inputStruct.TableName, inputStruct.Filter)

	_, err = e.client.Queryx(sqlStatement)

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

	sqlStatement, values := buildSQLStatementCreateTable(inputStruct.TableName, inputStruct.ColumnsStructure)

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
