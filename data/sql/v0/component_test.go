package sql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"github.com/jmoiron/sqlx"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type MockSQLClient struct{}

func (m *MockSQLClient) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	mock.ExpectQuery("SELECT (.+) FROM (users) WHERE (id = (.+) AND name = (.+) AND email = (.+) LIMIT (.+) OFFSET (.+))").
		WithArgs("1", "john", "john@example.com", 1, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).AddRow("1", "john", "john@example.com"))

	return sqlxDB.Queryx("SELECT id, name, email FROM users WHERE id = ? AND name = ? AND email = ? LIMIT ? OFFSET ?", "1", "john", "john@example.com", 1, 0)
}

func (m *MockSQLClient) NamedExec(query string, arg interface{}) (sql.Result, error) {
	mockDB, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	defer mockDB.Close()

	if strings.Contains(query, "INSERT") {
		fmt.Print(arg)
		arg = map[string]interface{}{
			"id":   "1",
			"name": "John Doe",
		}

		mock.ExpectExec("INSERT INTO users \\(id, name\\) VALUES \\(\\?, \\?\\)").
			WithArgs("1", "John Doe").WillReturnResult(sqlmock.NewResult(1, 1))

		return sqlxDB.NamedExec("INSERT INTO users (id, name) VALUES (:id, :name)", arg)
	} else if strings.Contains(query, "INSERT INTO usersMany") {
		arg = []map[string]interface{}{
			{"id": "1", "name": "John Doe"},
		}

		mock.ExpectExec("INSERT INTO usersMany \\(id, name\\) VALUES \\(\\?, \\?\\), \\(\\?, \\?\\)").
			WithArgs("1", "John Doe", "2", "Jane Doe").WillReturnResult(sqlmock.NewResult(1, 1))

		return sqlxDB.NamedExec("INSERT INTO usersMany (id, name) VALUES (:id, :name), (:id, :name)", arg)
	} else if strings.Contains(query, "DELETE") {
		arg = map[string]interface{}{
			"id":   "1",
			"name": "john",
		}

		mock.ExpectExec("DELETE FROM users WHERE id = \\? AND name = \\?").
			WithArgs("1", "john").WillReturnResult(sqlmock.NewResult(1, 1))

		return sqlxDB.NamedExec("DELETE FROM users WHERE id = :id AND name = :name", arg)
	} else if strings.Contains(query, "UPDATE") {
		arg = map[string]interface{}{
			"id":   "1",
			"name": "John Doe Updated",
		}

		mock.ExpectExec("UPDATE users SET id = \\?, name = \\? WHERE id = \\? AND name = \\?").
			WithArgs("1", "John Doe Updated", "1", "John Doe Updated").WillReturnResult(sqlmock.NewResult(1, 1))

		return sqlxDB.NamedExec("UPDATE users SET id = :id, name = :name WHERE id = :id AND name = :name", arg)
	} else if strings.Contains(query, "CREATE") {
		arg = map[string]interface{}{
			"id":   "INT",
			"name": "VARCHAR(255)",
		}

		mock.ExpectExec("CREATE TABLE users \\(id INT, name VARCHAR\\(255\\)\\)").
			WillReturnResult(sqlmock.NewResult(1, 1))

		return sqlxDB.NamedExec("CREATE TABLE users (id INT, name VARCHAR(255))", arg)
	} else if strings.Contains(query, "DROP") {
		arg = map[string]interface{}{}

		mock.ExpectExec("DROP TABLE users").
			WillReturnResult(sqlmock.NewResult(1, 1))

		return sqlxDB.NamedExec("DROP TABLE users", arg)
	}

	return nil, nil
}

func TestComponent_ExecuteInsertTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    InsertInput
		wantResp InsertOutput
		wantErr  string
	}{
		{
			name: "insert user",
			input: InsertInput{
				Data: map[string]any{
					"id":   "1",
					"name": "John Doe",
				},
				TableName: "users",
			},
			wantResp: InsertOutput{
				Status: "Successfully inserted 1 row",
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"user":     "test_user",
				"password": "test_pass",
				"name":     "test_db",
				"host":     "localhost",
				"port":     "3306",
				"region":   "us-west-2",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskInsert},
				client:             &MockSQLClient{},
			}
			e.execute = e.insert
			exec := &base.ExecutionWrapper{Execution: e}

			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			got, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})

			if tc.wantErr != "" {
				c.Assert(err, qt.ErrorMatches, tc.wantErr)
				return
			}

			wantJSON, err := json.Marshal(tc.wantResp)
			c.Assert(err, qt.IsNil)
			c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
		})
	}
}

func TestComponent_ExecuteInsertManyTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    InsertManyInput
		wantResp InsertManyOutput
		wantErr  string
	}{
		{
			name: "insert many users",
			input: InsertManyInput{
				ArrayData: []map[string]any{
					{"id": "1", "name": "John Doe"},
				},
				TableName: "users",
			},
			wantResp: InsertManyOutput{
				Status: "Successfully inserted 1 rows",
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"user":     "test_user",
				"password": "test_pass",
				"name":     "test_db",
				"host":     "localhost",
				"port":     "3306",
				"region":   "us-west-2",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskInsertMany},
				client:             &MockSQLClient{},
			}
			e.execute = e.insertMany
			exec := &base.ExecutionWrapper{Execution: e}

			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			got, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})

			if tc.wantErr != "" {
				c.Assert(err, qt.ErrorMatches, tc.wantErr)
				return
			}

			wantJSON, err := json.Marshal(tc.wantResp)
			c.Assert(err, qt.IsNil)
			c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
		})
	}
}

func TestComponent_ExecuteUpdateTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    UpdateInput
		wantResp UpdateOutput
		wantErr  string
	}{
		{
			name: "update user",
			input: UpdateInput{
				Filter: "id = 1 AND name = 'John Doe'",
				UpdateData: map[string]any{
					"id":   "1",
					"name": "John Doe Updated",
				},
				TableName: "users",
			},
			wantResp: UpdateOutput{
				Status: "Successfully updated 1 rows",
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"user":     "test_user",
				"password": "test_pass",
				"name":     "test_db",
				"host":     "localhost",
				"port":     "3306",
				"region":   "us-west-2",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskInsert},
				client:             &MockSQLClient{},
			}
			e.execute = e.update
			exec := &base.ExecutionWrapper{Execution: e}

			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			got, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})

			if tc.wantErr != "" {
				c.Assert(err, qt.ErrorMatches, tc.wantErr)
				return
			}

			wantJSON, err := json.Marshal(tc.wantResp)
			c.Assert(err, qt.IsNil)
			c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
		})
	}
}

func TestComponent_ExecuteSelectTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    SelectInput
		wantResp SelectOutput
		wantErr  string
	}{
		{
			name: "select users",
			input: SelectInput{
				Filter:    "id = 1 AND name = 'john' AND email = 'john@example.com'",
				TableName: "users",
				Limit:     0,
			},
			wantResp: SelectOutput{
				Status: "Successfully selected 1 rows",
				Rows: []map[string]any{
					{"id": "1", "name": "john", "email": "john@example.com"},
				},
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"user":     "test_user",
				"password": "test_pass",
				"name":     "test_db",
				"host":     "localhost",
				"port":     "3306",
				"region":   "us-west-2",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskSelect},
				client:             &MockSQLClient{},
			}
			e.execute = e.selects
			exec := &base.ExecutionWrapper{Execution: e}

			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			got, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})

			if tc.wantErr != "" {
				c.Assert(err, qt.ErrorMatches, tc.wantErr)
				return
			}

			wantJSON, err := json.Marshal(tc.wantResp)
			c.Assert(err, qt.IsNil)
			c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
		})
	}
}

func TestComponent_ExecuteDeleteTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    DeleteInput
		wantResp DeleteOutput
		wantErr  string
	}{
		{
			name: "delete user",
			input: DeleteInput{
				Filter:    "id = 1 AND name = 'john'",
				TableName: "users",
			},
			wantResp: DeleteOutput{
				Status: "Successfully deleted 1 rows",
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"user":     "test_user",
				"password": "test_pass",
				"name":     "test_db",
				"host":     "localhost",
				"port":     "3306",
				"region":   "us-west-2",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskDelete},
				client:             &MockSQLClient{},
			}
			e.execute = e.delete
			exec := &base.ExecutionWrapper{Execution: e}

			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			got, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})

			if tc.wantErr != "" {
				c.Assert(err, qt.ErrorMatches, tc.wantErr)
				return
			}

			wantJSON, err := json.Marshal(tc.wantResp)
			c.Assert(err, qt.IsNil)
			c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
		})
	}
}

func TestComponent_ExecuteCreateTableTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    CreateTableInput
		wantResp CreateTableOutput
		wantErr  string
	}{
		{
			name: "create table",
			input: CreateTableInput{
				ColumnsStructure: map[string]string{
					"id":   "INT",
					"name": "VARCHAR(255)",
				},
				TableName: "users",
			},
			wantResp: CreateTableOutput{
				Status: "Successfully created 1 table",
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"user":     "test_user",
				"password": "test_pass",
				"name":     "test_db",
				"host":     "localhost",
				"port":     "3306",
				"region":   "us-west-2",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskCreateTable},
				client:             &MockSQLClient{},
			}
			e.execute = e.createTable
			exec := &base.ExecutionWrapper{Execution: e}

			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			got, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})

			if tc.wantErr != "" {
				c.Assert(err, qt.ErrorMatches, tc.wantErr)
				return
			}

			wantJSON, err := json.Marshal(tc.wantResp)
			c.Assert(err, qt.IsNil)
			c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
		})
	}
}

func TestComponent_ExecuteDropTableTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    DropTableInput
		wantResp DropTableOutput
		wantErr  string
	}{
		{
			name: "drop table",
			input: DropTableInput{
				TableName: "users",
			},
			wantResp: DropTableOutput{
				Status: "Successfully dropped 1 table",
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"user":     "test_user",
				"password": "test_pass",
				"name":     "test_db",
				"host":     "localhost",
				"port":     "3306",
				"region":   "us-west-2",
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskDropTable},
				client:             &MockSQLClient{},
			}
			e.execute = e.dropTable
			exec := &base.ExecutionWrapper{Execution: e}

			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			got, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})

			if tc.wantErr != "" {
				c.Assert(err, qt.ErrorMatches, tc.wantErr)
				return
			}

			wantJSON, err := json.Marshal(tc.wantResp)
			c.Assert(err, qt.IsNil)
			c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
		})
	}
}
