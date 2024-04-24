package operator

import (
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/component/pkg/operator/base64/v0"
	"github.com/instill-ai/component/pkg/operator/end/v0"
	"github.com/instill-ai/component/pkg/operator/image/v0"
	"github.com/instill-ai/component/pkg/operator/json/v0"
	"github.com/instill-ai/component/pkg/operator/start/v0"
	"github.com/instill-ai/component/pkg/operator/text/v0"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

var (
	once    sync.Once
	opStore *OperatorStore
)

// Operator is the derived operator
type OperatorStore struct {
	operatorUIDs   []uuid.UUID
	operatorUIDMap map[uuid.UUID]*operator
	operatorIDMap  map[string]*operator
}

type operator struct {
	op base.IOperator
}

// Init initializes the operator
func Init(logger *zap.Logger, usageHandler base.UsageHandler) *OperatorStore {
	once.Do(func() {
		opStore = &OperatorStore{
			operatorUIDMap: map[uuid.UUID]*operator{},
			operatorIDMap:  map[string]*operator{},
		}
		opStore.Import(start.Init(logger, usageHandler)) // deprecated
		opStore.Import(end.Init(logger, usageHandler))   // deprecated
		opStore.Import(base64.Init(logger, usageHandler))
		opStore.Import(json.Init(logger, usageHandler))
		opStore.Import(image.Init(logger, usageHandler))
		opStore.Import(text.Init(logger, usageHandler))

	})
	return opStore
}

// Imports imports the operator definitions
func (os *OperatorStore) Import(op base.IOperator) {
	o := &operator{op: op}
	os.operatorUIDMap[op.GetUID()] = o
	os.operatorIDMap[op.GetID()] = o
	os.operatorUIDs = append(os.operatorUIDs, op.GetUID())
}

func (os *OperatorStore) CreateExecution(defUID uuid.UUID, sysVars map[string]any, task string) (*base.ExecutionWrapper, error) {
	if op, ok := os.operatorUIDMap[defUID]; ok {
		return op.op.CreateExecution(sysVars, task)
	}
	return nil, fmt.Errorf("operator definition not found")
}

func (os *OperatorStore) GetOperatorDefinitionByUID(defUID uuid.UUID, sysVars map[string]any, component *pipelinePB.OperatorComponent) (*pipelinePB.OperatorDefinition, error) {
	if op, ok := os.operatorUIDMap[defUID]; ok {
		return op.op.GetOperatorDefinition(sysVars, component)
	}
	return nil, fmt.Errorf("operator definition not found")
}

// Get the operator definition by definition id
func (os *OperatorStore) GetOperatorDefinitionByID(defID string, sysVars map[string]any, component *pipelinePB.OperatorComponent) (*pipelinePB.OperatorDefinition, error) {
	if op, ok := os.operatorIDMap[defID]; ok {
		return op.op.GetOperatorDefinition(sysVars, component)
	}
	return nil, fmt.Errorf("operator definition not found")
}

// Get the list of operator definitions under this operator
func (os *OperatorStore) ListOperatorDefinitions(sysVars map[string]any, returnTombstone bool) []*pipelinePB.OperatorDefinition {
	defs := []*pipelinePB.OperatorDefinition{}
	for _, uid := range os.operatorUIDs {
		op := os.operatorUIDMap[uid]
		def, err := op.op.GetOperatorDefinition(sysVars, nil)
		if err == nil {
			if !def.Tombstone || returnTombstone {
				defs = append(defs, def)
			}
		}
	}
	return defs
}
