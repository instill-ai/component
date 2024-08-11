//go:generate compogen readme ./config ./README.mdx
package sql

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"sync"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/x/errmsg"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	TaskInsert      = "TASK_INSERT"
	TaskInsertMany  = "TASK_INSERT_MANY"
	TaskUpdate      = "TASK_UPDATE"
	TaskSelect      = "TASK_SELECT"
	TaskDelete      = "TASK_DELETE"
	TaskCreateTable = "TASK_CREATE_TABLE"
	TaskDropTable   = "TASK_DROP_TABLE"
)

//go:embed config/definition.json
var definitionJSON []byte

//go:embed config/setup.json
var setupJSON []byte

//go:embed config/tasks.json
var tasksJSON []byte

var once sync.Once
var comp *component

type SQLClient interface {
	NamedExec(query string, arg interface{}) (sql.Result, error)
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
}

type component struct {
	base.Component
}

type execution struct {
	base.ComponentExecution

	execute func(*structpb.Struct) (*structpb.Struct, error)
	client  SQLClient
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

func (c *component) CreateExecution(x base.ComponentExecution) (base.IExecution, error) {
	e := &execution{
		ComponentExecution: x,
	}

	switch x.Task {
	case TaskInsert:
		e.execute = e.insert
	case TaskUpdate:
		e.execute = e.update
	case TaskSelect:
		e.execute = e.selects
	case TaskDelete:
		e.execute = e.delete
	case TaskCreateTable:
		e.execute = e.createTable
	case TaskDropTable:
		e.execute = e.dropTable
	case TaskInsertMany:
		e.execute = e.insertMany
	default:
		return nil, errmsg.AddMessage(
			fmt.Errorf("not supported task: %s", x.Task),
			fmt.Sprintf("%s task is not supported.", x.Task),
		)
	}
	return e, nil
}

type Engine struct {
	DBEngine string `json:"engine"`
}

// newClient being setup here in the Execute since engine is part of the input,
// therefore, every new inputs will create a new connection
func (e *execution) Execute(ctx context.Context, in base.InputReader, out base.OutputWriter) error {
	inputs, err := in.Read(ctx)
	if err != nil {
		return err
	}
	outputs := make([]*structpb.Struct, len(inputs))

	for i, input := range inputs {
		var inputStruct Engine
		err := base.ConvertFromStructpb(input, &inputStruct)
		if err != nil {
			return err
		}

		if e.client == nil {
			e.client = newClient(e.Setup, &inputStruct)
		}

		output, err := e.execute(input)
		if err != nil {
			return err
		}

		outputs[i] = output
	}

	return out.Write(ctx, outputs)
}
