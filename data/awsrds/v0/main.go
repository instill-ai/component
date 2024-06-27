//go:generate compogen readme ./config ./README.mdx
package awsrds

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds/rdsutils"

	"github.com/instill-ai/component/base"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/types/known/structpb"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/sijms/go-ora/v2"
)

const (
	taskUpsert = "TASK_UPSERT"
	taskSelect = "TASK_SELECT"
	taskDelete = "TASK_DELETE"
	taskQuery  = "TASK_QUERY"
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

type Config struct {
	DBUser               string
	DBPassword           string
	DBName               string
	DBHost               string
	DBPort               string
	Region               string
	DBAWSAccessKeyId     string
	DBAWSSecretAccessKey string
}

func LoadConfig(e *execution) *Config {
	return &Config{
		DBUser:               getUser(e.Setup),
		DBPassword:           getPassword(e.Setup),
		DBName:               getName(e.Setup),
		DBHost:               getHost(e.Setup),
		DBPort:               getPort(e.Setup),
		Region:               getRegion(e.Setup),
		DBAWSAccessKeyId:     getAWSAccessKeyId(e.Setup),
		DBAWSSecretAccessKey: getAWSSecretAccessKey(e.Setup),
	}
}

var engines = []string{
	"postgresql://%s:%s@%s/%s", "sqlserver://%s:%s@%s%s", "oracle://%s:%s@%s/%s", "%s:%s@tcp(%s)/%s",
}

func newClient(e *execution) (*sqlx.DB, error) {
	cfg := LoadConfig(e)

	DBEndpoint := fmt.Sprintf("%v:%v", cfg.DBHost, cfg.DBPort)

	var db *sqlx.DB
	var err error
	for _, engine := range engines {

		dsn := fmt.Sprintf(engine,
			cfg.DBUser, cfg.DBPassword, DBEndpoint, cfg.DBName,
		)

		db, err = sqlx.Open("mysql", dsn)
		if err == nil {
			return db, nil
		}
	}

	return nil, fmt.Errorf("error creating SQL client")
}

func newClientIAM(e *execution) (*sqlx.DB, error) {
	cfg := LoadConfig(e)
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.Region),
		Credentials: credentials.NewStaticCredentials(
			cfg.DBAWSAccessKeyId, cfg.DBAWSSecretAccessKey, ""),
	})
	if err != nil {
		return nil, err
	}

	DBEndpoint := fmt.Sprintf("%s:%s", cfg.DBHost, cfg.DBPort)

	authToken, err := rdsutils.BuildAuthToken(
		DBEndpoint,
		cfg.Region,
		cfg.DBUser,
		sess.Config.Credentials,
	)
	if err != nil {
		return nil, err
	}

	var db *sqlx.DB
	for _, engine := range engines {
		dsn := fmt.Sprintf(engine+"?tls=true&tls=skip-verify&allowCleartextPasswords=true",
			cfg.DBUser, authToken, DBEndpoint, cfg.DBName,
		)

		db, err = sqlx.Open("mysql", dsn)
		if err == nil {
			return db, nil
		}
	}

	return nil, fmt.Errorf("error creating SQL client")
}

func getUser(setup *structpb.Struct) string {
	return setup.GetFields()["user"].GetStringValue()
}
func getPassword(setup *structpb.Struct) string {
	return setup.GetFields()["password"].GetStringValue()
}
func getName(setup *structpb.Struct) string {
	return setup.GetFields()["name"].GetStringValue()
}
func getHost(setup *structpb.Struct) string {
	return setup.GetFields()["host"].GetStringValue()
}
func getPort(setup *structpb.Struct) string {
	port := setup.GetFields()["port"].GetNumberValue()
	portStr := strconv.FormatFloat(port, 'f', -1, 64)
	return portStr
}
func getRegion(setup *structpb.Struct) string {
	return setup.GetFields()["region"].GetStringValue()
}
func getAWSAccessKeyId(setup *structpb.Struct) string {
	return setup.GetFields()["aws-access-key-id"].GetStringValue()
}
func getAWSSecretAccessKey(setup *structpb.Struct) string {
	return setup.GetFields()["aws-secret-access-key"].GetStringValue()
}

func (e *execution) Execute(ctx context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := []*structpb.Struct{}

	var client *sqlx.DB
	var err error

	if e.Setup.GetFields()["password"] != nil {
		client, err = newClient(e)
	} else if e.Setup.GetFields()["password"] == nil && e.Setup.GetFields()["aws-access-key-id"] != nil && e.Setup.GetFields()["aws-secret-access-key"] != nil {
		client, err = newClientIAM(e)
	} else {
		return nil, fmt.Errorf("some fields are missing, please fill in the required fields")
	}

	if err != nil {
		return nil, fmt.Errorf("error creating SQL client: %v", err)
	}
	defer client.Close()

	for _, input := range inputs {
		var output *structpb.Struct

		switch e.Task {
		case taskUpsert:
			tableName := input.GetFields()["table-name"].GetStringValue()

			// Unmarshal the JSON data from the protobuf into structpb.Struct
			data := input.GetFields()["data"].GetStructValue()

			err := UpsertUser(client, &tableName, data)
			if err != nil {
				return nil, err
			}

			output = &structpb.Struct{Fields: map[string]*structpb.Value{
				"status": {Kind: &structpb.Value_StringValue{StringValue: "Successfully executed SQL statement"}}}}
		case taskSelect:
			tableName := input.GetFields()["table-name"].GetStringValue()

			// Unmarshal the JSON data from the protobuf into structpb.Struct
			columns := input.GetFields()["columns"].GetListValue()
			from := input.GetFields()["from"].GetNumberValue()
			to := input.GetFields()["to"].GetNumberValue()

			outputList, err := SelectUser(client, &tableName, columns, int(from), int(to))
			if err != nil {
				return nil, err
			}
			output = &structpb.Struct{Fields: map[string]*structpb.Value{
				"status": {Kind: &structpb.Value_StringValue{StringValue: "Successfully executed SQL statement"}},
				"rows":   {Kind: &structpb.Value_ListValue{ListValue: outputList}},
			}}

		case taskDelete:
			tableName := input.GetFields()["table-name"].GetStringValue()

			// Unmarshal the JSON data from the protobuf into structpb.Struct
			criteria := input.GetFields()["criteria"].GetStructValue()

			err := DeleteUser(client, &tableName, criteria)
			if err != nil {
				return nil, err
			}

			output = &structpb.Struct{Fields: map[string]*structpb.Value{
				"status": {Kind: &structpb.Value_StringValue{StringValue: "Successfully executed SQL statement"}}}}
		case taskQuery:
			query := input.GetFields()["query"].GetStringValue()

			err := QueryUser(client, query)
			if err != nil {
				return nil, err
			}

			output = &structpb.Struct{Fields: map[string]*structpb.Value{
				"status": {Kind: &structpb.Value_StringValue{StringValue: "Successfully executed SQL statement"}}}}
		default:
			return nil, fmt.Errorf("task %s not supported", e.Task)
		}
		outputs = append(outputs, output)
	}
	return outputs, nil
}

func UpsertUser(client *sqlx.DB, tableName *string, data *structpb.Struct) error {
	// Begin constructing SQL statement
	sqlStatement := "INSERT INTO " + *tableName + " ("
	var columns []string
	var placeholders []string
	var updateColumns []string
	values := make(map[string]interface{}) // Initialize the map

	for dataKey, dataValue := range data.GetFields() {
		columns = append(columns, dataKey)
		placeholders = append(placeholders, ":"+dataKey)
		updateColumns = append(updateColumns, fmt.Sprintf("%s = :%s", dataKey, dataKey))
		values[dataKey] = dataValue.AsInterface()
	}

	sqlStatement += strings.Join(columns, ", ") + ") VALUES (" + strings.Join(placeholders, ", ") + ") "
	sqlStatement += "ON DUPLICATE KEY UPDATE " + strings.Join(updateColumns, ", ")

	// Prepare and execute the statement using NamedExec
	_, err := client.NamedExec(sqlStatement, values)

	if err != nil {
		return err
	}

	return nil
}

func SelectUser(client *sqlx.DB, tableName *string, columns *structpb.ListValue, from int, to int) (*structpb.ListValue, error) {
	// Begin constructing SQL statement
	sqlStatement := "SELECT "
	var columnNames []string
	for _, column := range columns.Values {
		columnNames = append(columnNames, column.GetStringValue())
	}

	var notAll string
	if to == 0 && from == 0 {
		notAll = ""
	} else {
		notAll = " LIMIT " + strconv.Itoa(to-from+1) + " OFFSET " + strconv.Itoa(from-1)
	}

	sqlStatement += strings.Join(columnNames, ", ") + " FROM " + *tableName + notAll
	// Prepare and execute the statement
	rows, err := client.Queryx(sqlStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Prepare the result ListValue
	result := &structpb.ListValue{}

	// Iterate over the rows
	for rows.Next() {
		// Create a map to hold the row data
		rowMap := make(map[string]interface{})

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

		// Convert the map to a structpb.Struct
		structValue, err := structpb.NewStruct(rowMap)
		if err != nil {
			return nil, fmt.Errorf("failed to convert map to structpb.Struct: %v", err)
		}

		// Convert the structpb.Struct to a structpb.Value
		structValuePb := structpb.NewStructValue(structValue)

		// Add the structpb.Value to the result ListValue
		result.Values = append(result.Values, structValuePb)
	}

	// Check for errors after iterating over rows
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during row iteration: %v", err)
	}

	return result, nil
}

func DeleteUser(client *sqlx.DB, tableName *string, criteria *structpb.Struct) error {
	// Begin constructing SQL statement
	sqlStatement := "DELETE FROM " + *tableName + " WHERE "
	var where []string
	values := make(map[string]interface{}) // Initialize the map

	for criteriaKey, criteriaValue := range criteria.GetFields() {
		where = append(where, fmt.Sprintf("%s = :%s", criteriaKey, criteriaKey))
		values[criteriaKey] = criteriaValue.AsInterface()
	}

	sqlStatement += strings.Join(where, " AND ")

	// Prepare and execute the statement using NamedExec
	_, err := client.NamedExec(sqlStatement, values)

	if err != nil {
		return err
	}

	return nil
}

func QueryUser(client *sqlx.DB, query string) error {
	// Prepare and execute the statement
	rows, err := client.Queryx(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	return nil
}

func (c *component) Test(sysVars map[string]any, setup *structpb.Struct) error {
	client, err := newClient(&execution{ComponentExecution: base.ComponentExecution{Setup: setup}})
	if err != nil || client == nil {
		return fmt.Errorf("error creating RDS client: %v", err)
	}
	defer client.Close()

	err = client.Ping()
	if err != nil {
		return fmt.Errorf("error pinging RDS database: %v", err)
	}

	return nil
}
