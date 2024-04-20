package start

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
var con *operator

type operator struct {
	base.BaseOperator
}

type execution struct {
	base.BaseOperatorExecution
}

func Init(l *zap.Logger, u base.UsageHandler) *operator {
	once.Do(func() {
		con = &operator{
			BaseOperator: base.BaseOperator{
				Logger:       l,
				UsageHandler: u,
			},
		}
		err := con.LoadOperatorDefinition(definitionJSON, tasksJSON, nil)
		if err != nil {
			panic(err)
		}
	})
	return con
}

func (o *operator) CreateExecution(sysVars map[string]any, task string) (*base.ExecutionWrapper, error) {
	return &base.ExecutionWrapper{Execution: &execution{
		BaseOperatorExecution: base.BaseOperatorExecution{Operator: o, Task: task},
	}}, nil
}

func (e *execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	return nil, fmt.Errorf("the Airbyte operator has been removed")
}

func (o *operator) Test(sysVars map[string]any, connection *structpb.Struct) error {
	return fmt.Errorf("the Airbyte operator has been removed")
}
