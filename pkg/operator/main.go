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
	"github.com/instill-ai/component/pkg/operator/pdf/v0"
	"github.com/instill-ai/component/pkg/operator/start/v0"
	"github.com/instill-ai/component/pkg/operator/text/v0"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

var (
	once    sync.Once
	opStore *Store
)

// Store holds in-memory information about the initialized operators.
type Store struct {
	operatorUIDs   []uuid.UUID
	operatorUIDMap map[uuid.UUID]*operator
	operatorIDMap  map[string]*operator
}

type operator struct {
	op base.IOperator
}

// Init initializes the different operator components and loads their
// information to memory.
func Init(logger *zap.Logger) *Store {
	baseOp := base.Operator{Logger: logger}

	once.Do(func() {
		opStore = &Store{
			operatorUIDMap: map[uuid.UUID]*operator{},
			operatorIDMap:  map[string]*operator{},
		}
		opStore.Import(start.Init(baseOp)) // deprecated
		opStore.Import(end.Init(baseOp))   // deprecated
		opStore.Import(base64.Init(baseOp))
		opStore.Import(json.Init(baseOp))
		opStore.Import(image.Init(baseOp))
		opStore.Import(text.Init(baseOp))
		opStore.Import(pdf.Init(baseOp))

	})
	return opStore
}

// Import loads the operator definitions into memory.
func (os *Store) Import(op base.IOperator) {
	o := &operator{op: op}
	os.operatorUIDMap[op.GetUID()] = o
	os.operatorIDMap[op.GetID()] = o
	os.operatorUIDs = append(os.operatorUIDs, op.GetUID())
}

// CreateExecution initializes the execution of a operator given its UID.
func (os *Store) CreateExecution(defUID uuid.UUID, sysVars map[string]any, task string) (*base.ExecutionWrapper, error) {
	if op, ok := os.operatorUIDMap[defUID]; ok {
		return op.op.CreateExecution(sysVars, task)
	}
	return nil, fmt.Errorf("operator definition not found")
}

// GetOperatorDefinitionByUID returns a operator definition by its UID.
func (os *Store) GetOperatorDefinitionByUID(defUID uuid.UUID, sysVars map[string]any, component *pipelinePB.OperatorComponent) (*pipelinePB.OperatorDefinition, error) {
	if op, ok := os.operatorUIDMap[defUID]; ok {
		return op.op.GetOperatorDefinition(sysVars, component)
	}
	return nil, fmt.Errorf("operator definition not found")
}

// GetOperatorDefinitionByID returns a operator definition by its ID.
func (os *Store) GetOperatorDefinitionByID(defID string, sysVars map[string]any, component *pipelinePB.OperatorComponent) (*pipelinePB.OperatorDefinition, error) {
	if op, ok := os.operatorIDMap[defID]; ok {
		return op.op.GetOperatorDefinition(sysVars, component)
	}
	return nil, fmt.Errorf("operator definition not found")
}

// ListOperatorDefinitions returns all the loaded operator definitions.
func (os *Store) ListOperatorDefinitions(sysVars map[string]any, returnTombstone bool) []*pipelinePB.OperatorDefinition {
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
