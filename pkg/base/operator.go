package base

import (
	"fmt"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1alpha"
)

// `IOperator` define the function interface for all operators.
type IOperator interface {
	IComponent

	// Functions that shared for all operators
	// Add operator definition
	AddOperatorDefinition(uid uuid.UUID, id string, def *pipelinePB.OperatorDefinition) error

	// Get the map of operator definitions under this operator
	GetOperatorDefinitionMap() map[uuid.UUID]*pipelinePB.OperatorDefinition
	// Get the operator definition by definition uid
	GetOperatorDefinitionByUid(defUid uuid.UUID) (*pipelinePB.OperatorDefinition, error)
	// Get the operator definition by definition id
	GetOperatorDefinitionById(defId string) (*pipelinePB.OperatorDefinition, error)
	// Get the list of operator definitions under this operator
	ListOperatorDefinitions() []*pipelinePB.OperatorDefinition
	// Get the list of operator definitions uuids
	ListOperatorDefinitionUids() []uuid.UUID

	// A helper function to check the operator has this definition by uid.
	HasUid(defUid uuid.UUID) bool
}

type BaseOperator struct {
	// Store all the operator defintions in the operator
	definitionMapByUid map[uuid.UUID]*pipelinePB.OperatorDefinition
	definitionMapById  map[string]*pipelinePB.OperatorDefinition

	// Used for ordered
	definitionUids []uuid.UUID

	// Logger
	Logger *zap.Logger
}

func (c *BaseOperator) AddOperatorDefinition(uid uuid.UUID, id string, def *pipelinePB.OperatorDefinition) error {
	if c.definitionMapByUid == nil {
		c.definitionMapByUid = map[uuid.UUID]*pipelinePB.OperatorDefinition{}
	}
	if c.definitionMapById == nil {
		c.definitionMapById = map[string]*pipelinePB.OperatorDefinition{}
	}
	c.definitionUids = append(c.definitionUids, uid)
	c.definitionMapByUid[uid] = def
	c.definitionMapById[id] = def
	return nil
}

func (c *BaseOperator) GetOperatorDefinitionMap() map[uuid.UUID]*pipelinePB.OperatorDefinition {
	return c.definitionMapByUid
}

func (c *BaseOperator) ListOperatorDefinitions() []*pipelinePB.OperatorDefinition {
	definitions := []*pipelinePB.OperatorDefinition{}
	for _, uid := range c.definitionUids {
		val, ok := c.definitionMapByUid[uid]
		if !ok {
			// logger
			c.Logger.Error("get operator defintion error")
		}
		definitions = append(definitions, val)
	}

	return definitions
}

func (c *BaseOperator) GetOperatorDefinitionByUid(defUid uuid.UUID) (*pipelinePB.OperatorDefinition, error) {
	val, ok := c.definitionMapByUid[defUid]
	if !ok {
		return nil, fmt.Errorf("get operator defintion error")
	}
	return val, nil
}

func (c *BaseOperator) GetOperatorDefinitionById(defId string) (*pipelinePB.OperatorDefinition, error) {
	val, ok := c.definitionMapById[defId]
	if !ok {
		return nil, fmt.Errorf("get operator defintion error")
	}
	return val, nil
}

func (c *BaseOperator) ListOperatorDefinitionUids() []uuid.UUID {
	return c.definitionUids
}

func (c *BaseOperator) HasUid(defUid uuid.UUID) bool {
	_, err := c.GetOperatorDefinitionByUid(defUid)
	return err == nil
}
