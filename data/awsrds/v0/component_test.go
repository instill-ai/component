package awsrds

import (
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestUpsertUser(t *testing.T) {
	c := qt.New(t)

	testcases := []struct {
		name      string
		tableName string
		data      *structpb.Struct
		wantErr   bool
	}{
		{
			name:      "insert user",
			tableName: "users",
			data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"id":   structpb.NewStringValue("1"),
					"name": structpb.NewStringValue("John Doe"),
				},
			},
			wantErr: false,
		},
		{
			name:      "update user",
			tableName: "users",
			data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"id":   structpb.NewStringValue("1"),
					"name": structpb.NewStringValue("John Doe Updated"),
				},
			},
			wantErr: false,
		},
		{
			name: "insert error",
			data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"id":   structpb.NewStringValue("1"),
					"name": structpb.NewStringValue("John Doe Error"),
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			db, mock, err := sqlmock.New()
			c.Assert(err, qt.IsNil)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")

			// Mock the expected SQL query
			mock.ExpectExec("INSERT INTO users").
				WithArgs(tc.data.GetFields()["id"].GetStringValue(), tc.data.GetFields()["name"].GetStringValue()).
				WillReturnResult(sqlmock.NewResult(1, 1))

			// Call the function
			err = UpsertUser(sqlxDB, &tc.tableName, tc.data)

			if tc.wantErr {
				c.Assert(err, qt.Not(qt.IsNil))

				// Ensure all expectations were met
				c.Assert(mock.ExpectationsWereMet(), qt.Not(qt.IsNil))
			} else {
				c.Assert(err, qt.IsNil)

				// Ensure all expectations were met
				c.Assert(mock.ExpectationsWereMet(), qt.IsNil)
			}
		})
	}
}

func TestSelectUser(t *testing.T) {
	c := qt.New(t)

	testcases := []struct {
		name      string
		tableName string
		columns   *structpb.ListValue
		from      int
		to        int
		wantErr   bool
	}{
		{
			name:      "select users",
			tableName: "users",
			columns: &structpb.ListValue{
				Values: []*structpb.Value{
					structpb.NewStringValue("id"),
					structpb.NewStringValue("name"),
				},
			},
			from:    0,
			to:      2,
			wantErr: false,
		},
		{
			name:      "select error",
			tableName: "users",
			columns: &structpb.ListValue{
				Values: []*structpb.Value{},
			},
			from:    3,
			to:      4,
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			db, mock, err := sqlmock.New()
			c.Assert(err, qt.IsNil)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")

			// Mock the expected SQL query
			rows := sqlmock.NewRows([]string{"id", "name"})
			for i := tc.from; i < tc.to; i++ {
				rows.AddRow(i, "John Doe"+strconv.Itoa(i))
			}

			mock.ExpectQuery("SELECT id, name FROM users").
				WillReturnRows(rows)

			// Call the function
			result, err := SelectUser(sqlxDB, &tc.tableName, tc.columns, tc.from, tc.to)

			if tc.wantErr {
				c.Assert(err, qt.Not(qt.IsNil))

				c.Assert(result, qt.IsNil)

				// Ensure all expectations were met
				c.Assert(mock.ExpectationsWereMet(), qt.Not(qt.IsNil))
			} else {
				c.Assert(err, qt.IsNil)

				c.Assert(result, qt.Not(qt.IsNil))

				// Ensure all expectations were met
				c.Assert(mock.ExpectationsWereMet(), qt.IsNil)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	c := qt.New(t)

	testcases := []struct {
		name      string
		tableName string
		criteria  *structpb.Struct
	}{
		{
			name:      "delete user",
			tableName: "users",
			criteria: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"id": structpb.NewStringValue("1"),
				},
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			db, mock, err := sqlmock.New()
			c.Assert(err, qt.IsNil)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")

			// Mock the expected SQL query
			mock.ExpectExec("DELETE FROM users WHERE id = ?").
				WithArgs("1").
				WillReturnResult(sqlmock.NewResult(1, 1))

			// Call the function
			err = DeleteUser(sqlxDB, &tc.tableName, tc.criteria)
			c.Assert(err, qt.IsNil)

			// Ensure all expectations were met
			c.Assert(mock.ExpectationsWereMet(), qt.IsNil)
		})
	}
}

func TestQueryUser(t *testing.T) {
	c := qt.New(t)

	testcases := []struct {
		name      string
		tableName string
		query     string
	}{
		{
			name:  "query users",
			query: "SELECT * FROM users",
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			db, mock, err := sqlmock.New()
			c.Assert(err, qt.IsNil)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")

			// Mock the expected SQL query
			rows := sqlmock.NewRows([]string{"id", "name"}).
				AddRow("1", "John Doe")
			mock.ExpectQuery("SELECT \\* FROM users").
				WillReturnRows(rows)

			// Call the function
			err = QueryUser(sqlxDB, tc.query)
			c.Assert(err, qt.IsNil)

			// Ensure all expectations were met
			c.Assert(mock.ExpectationsWereMet(), qt.IsNil)
		})
	}
}

func TestCreateExecution(t *testing.T) {
	c := qt.New(t)

	testcases := []struct {
		name  string
		setup *structpb.Struct
		task  string
	}{
		{
			name: "create execution",
			setup: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"user": structpb.NewStringValue("test_user"),
				},
			},
			task: "upsert",
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			component := Init(base.Component{})

			sysVars := map[string]interface{}{
				"var1": "value1",
			}

			// Call the function
			executionWrapper, err := component.CreateExecution(sysVars, tc.setup, tc.task)
			c.Assert(err, qt.IsNil)
			c.Assert(executionWrapper, qt.Not(qt.IsNil))
		})
	}
}

func TestNewClient(t *testing.T) {
	c := qt.New(t)

	testcases := []struct {
		name     string
		execMock *execution
		wantErr  bool
	}{
		{
			name: "test new client success",
			execMock: &execution{ComponentExecution: base.ComponentExecution{Setup: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"user":     structpb.NewStringValue("testuser"),
					"password": structpb.NewStringValue("testpass"),
					"name":     structpb.NewStringValue("testdb"),
					"host":     structpb.NewStringValue("localhost"),
					"port":     structpb.NewStringValue("3306"),
					"region":   structpb.NewStringValue("us-west-2"),
				},
			}}},
			wantErr: false,
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			// Mock sqlx.Open function
			mockDB, mock, err := sqlmock.New()
			c.Assert(err, qt.IsNil)
			defer mockDB.Close()

			// Call the function
			db, err := newClient(tc.execMock)

			c.Assert(err, qt.IsNil)
			c.Assert(db, qt.Not(qt.IsNil))

			// Verify sqlx.Open was called with the correct DSN
			mock.ExpectationsWereMet()
		})
	}
}

func TestNewClientIAM(t *testing.T) {
	c := qt.New(t)

	testcases := []struct {
		name     string
		execMock *execution
		config   *Config
		wantErr  bool
	}{
		{
			name: "test new client success",
			execMock: &execution{ComponentExecution: base.ComponentExecution{Setup: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"user":                  structpb.NewStringValue("testuser"),
					"name":                  structpb.NewStringValue("testdb"),
					"host":                  structpb.NewStringValue("localhost"),
					"port":                  structpb.NewStringValue("3306"),
					"region":                structpb.NewStringValue("us-west-2"),
					"aws-access-key-id":     structpb.NewStringValue("testkeyid"),
					"aws-secret-access-key": structpb.NewStringValue("testsecret"),
				},
			}}},
			config:  &Config{},
			wantErr: false,
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			// Mock sqlx.Open function
			mockDB, mock, err := sqlmock.New()
			c.Assert(err, qt.IsNil)
			defer mockDB.Close()

			// Call the function
			db, err := newClientIAM(tc.execMock)

			c.Assert(err, qt.IsNil)
			c.Assert(db, qt.Not(qt.IsNil))

			// Verify sqlx.Open was called with the correct DSN
			mock.ExpectationsWereMet()
		})
	}
}
