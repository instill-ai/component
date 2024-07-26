package sql

import (
	"fmt"
	"strconv"

	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/types/known/structpb"

	// Import all the SQL drivers
	_ "github.com/denisenkom/go-mssqldb" // SQL Server
	_ "github.com/go-sql-driver/mysql"   // MySQL and MariaDB
	_ "github.com/lib/pq"                // PostgreSQL
	_ "github.com/nakagami/firebirdsql"  // Firebird
	_ "github.com/sijms/go-ora/v2"       // Oracle
)

var engines = map[string]string{
	"PostgreSQL": "postgresql://%s:%s@%s/%s",         // PostgreSQL
	"SQL Server": "sqlserver://%s:%s@%s?database=%s", // SQL Server
	"Oracle":     "oracle://%s:%s@%s/%s",             // Oracle
	"MySQL":      "%s:%s@tcp(%s)/%s",                 // MySQL and MariaDB
	"Firebird":   "firebirdsql://%s:%s@%s/%s",        // Firebird
}

var enginesType = map[string]string{
	"PostgreSQL": "postgres",    // PostgreSQL
	"SQL Server": "sqlserver",   // SQL Server
	"Oracle":     "oracle",      // Oracle
	"MySQL":      "mysql",       // MySQL and MariaDB
	"Firebird":   "firebirdsql", // Firebird
}

type Config struct {
	DBUser     string
	DBPassword string
	DBName     string
	DBHost     string
	DBPort     string
}

func LoadConfig(setup *structpb.Struct) *Config {
	return &Config{
		DBUser:     getUser(setup),
		DBPassword: getPassword(setup),
		DBName:     getDatabaseName(setup),
		DBHost:     getHost(setup),
		DBPort:     getPort(setup),
	}
}

func newClient(setup *structpb.Struct, inputSetup *Engine) SQLClient {
	cfg := LoadConfig(setup)

	DBEndpoint := fmt.Sprintf("%v:%v", cfg.DBHost, cfg.DBPort)

	// Test every engines to find the correct one
	var db *sqlx.DB
	var err error

	// Get the correct engine
	engine := engines[inputSetup.DBEngine]
	engineType := enginesType[inputSetup.DBEngine]

	dsn := fmt.Sprintf(engine,
		cfg.DBUser, cfg.DBPassword, DBEndpoint, cfg.DBName,
	)

	db, err = sqlx.Open(engineType, dsn)
	if err != nil {
		return nil
	}

	return db
}

func getUser(setup *structpb.Struct) string {
	return setup.GetFields()["user"].GetStringValue()
}
func getPassword(setup *structpb.Struct) string {
	return setup.GetFields()["password"].GetStringValue()
}
func getDatabaseName(setup *structpb.Struct) string {
	return setup.GetFields()["database-name"].GetStringValue()
}
func getHost(setup *structpb.Struct) string {
	return setup.GetFields()["host"].GetStringValue()
}
func getPort(setup *structpb.Struct) string {
	return strconv.Itoa(int(setup.GetFields()["port"].GetNumberValue()))
}