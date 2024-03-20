package pkg

import (
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/component/pkg/operator/base64/v0"
	"github.com/instill-ai/component/pkg/operator/end/v0"
	"github.com/instill-ai/component/pkg/operator/image/v0"
	"github.com/instill-ai/component/pkg/operator/json/v0"
	"github.com/instill-ai/component/pkg/operator/start/v0"
	"github.com/instill-ai/component/pkg/operator/text/v0"
)

var (
	once     sync.Once
	operator base.IOperator
)

// Operator is the derived operator
type Operator struct {
	base.Operator
	operatorUIDMap map[uuid.UUID]base.IOperator
}

// Init initializes the operator
func Init(logger *zap.Logger) base.IOperator {
	once.Do(func() {
		operator = &Operator{
			Operator:       base.Operator{Component: base.Component{Logger: logger}},
			operatorUIDMap: map[uuid.UUID]base.IOperator{},
		}
		operator.(*Operator).ImportDefinitions(base64.Init(logger))
		operator.(*Operator).ImportDefinitions(start.Init(logger))
		operator.(*Operator).ImportDefinitions(end.Init(logger))
		operator.(*Operator).ImportDefinitions(json.Init(logger))
		operator.(*Operator).ImportDefinitions(image.Init(logger))
		operator.(*Operator).ImportDefinitions(text.Init(logger))

	})
	return operator
}

// ImportDefinitions imports the operator definitions
func (o *Operator) ImportDefinitions(op base.IOperator) {
	for _, v := range op.ListOperatorDefinitions() {
		err := o.AddOperatorDefinition(v)
		if err != nil {
			panic(err)
		}
		o.operatorUIDMap[uuid.FromStringOrNil(v.Uid)] = op
	}
}

// CreateExecution creates the derived execution
func (o *Operator) CreateExecution(defUID uuid.UUID, task string, config *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	return o.operatorUIDMap[defUID].CreateExecution(defUID, task, config, logger)
}
