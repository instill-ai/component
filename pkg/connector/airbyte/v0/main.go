package airbyte

import (
	_ "embed"
	"fmt"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
)

//go:embed config/definition.json
var definitionJSON []byte

//go:embed config/tasks.json
var tasksJSON []byte

var once sync.Once
var con *connector

type connector struct {
	base.BaseConnector
}

type execution struct {
	base.BaseConnectorExecution
}

func Init(l *zap.Logger, u base.UsageHandler) *connector {
	once.Do(func() {
		con = &connector{
			BaseConnector: base.BaseConnector{
				Logger:       l,
				UsageHandler: u,
			},
		}
		err := con.LoadConnectorDefinition(definitionJSON, tasksJSON, nil)
		if err != nil {
			panic(err)
		}
	})
	return con
}

func (c *connector) CreateExecution(sysVars map[string]any, connection *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	return &base.ExecutionWrapper{Execution: &execution{
		BaseConnectorExecution: base.BaseConnectorExecution{Connector: c, SystemVariables: sysVars, Connection: connection, Task: task},
	}}, nil
}

func (e *execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	return nil, fmt.Errorf("the Airbyte connector has been removed")
}

func (c *connector) Test(sysVars map[string]any, connection *structpb.Struct) error {
	return fmt.Errorf("the Airbyte connector has been removed")
}
