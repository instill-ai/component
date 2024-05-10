package end

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
)

//go:embed config/definition.json
var definitionJSON []byte

//go:embed config/tasks.json
var tasksJSON []byte

var once sync.Once
var op *operator

type operator struct {
	base.Operator
}

type execution struct {
	base.OperatorExecution
}

func Init(bo base.Operator) *operator {
	once.Do(func() {
		op = &operator{Operator: bo}
		err := op.LoadOperatorDefinition(definitionJSON, tasksJSON, nil)
		if err != nil {
			panic(err)
		}
	})
	return op
}

func (o *operator) CreateExecution(sysVars map[string]any, task string) (*base.ExecutionWrapper, error) {
	return &base.ExecutionWrapper{Execution: &execution{
		OperatorExecution: base.OperatorExecution{Operator: o, SystemVariables: sysVars, Task: task},
	}}, nil
}

func (e *execution) Execute(_ context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	return nil, fmt.Errorf("the Airbyte operator has been removed")
}

func (o *operator) Test(sysVars map[string]any, connection *structpb.Struct) error {
	return fmt.Errorf("the Airbyte operator has been removed")
}
