package base

import (
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid"
	"google.golang.org/protobuf/encoding/protojson"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1alpha"
)

// IOperator is the interface that all operators need to implement
type IOperator interface {
	IComponent

	// Functions that shared for all operators
	// Load operator definitions from json files
	LoadOperatorDefinitions(definitionsJSON []byte, tasksJSON []byte) error
	// Add definition
	AddOperatorDefinition(def *pipelinePB.OperatorDefinition) error
	// Get the operator definition by definition uid
	GetOperatorDefinitionByUID(defUID uuid.UUID) (*pipelinePB.OperatorDefinition, error)
	// Get the operator definition by definition id
	GetOperatorDefinitionByID(defID string) (*pipelinePB.OperatorDefinition, error)
	// Get the list of operator definitions under this operator
	ListOperatorDefinitions() []*pipelinePB.OperatorDefinition
}

// Operator is the base struct for all operators
type Operator struct {
	Component
}

// LoadOperatorDefinitions loads the operator definitions from json files
func (o *Operator) LoadOperatorDefinitions(definitionsJSONBytes []byte, tasksJSONBytes []byte) error {
	var err error
	definitionsJSONList := &[]interface{}{}

	err = json.Unmarshal(definitionsJSONBytes, definitionsJSONList)
	if err != nil {
		return err
	}
	err = o.Component.loadTasks(tasksJSONBytes)
	if err != nil {
		return err
	}

	for _, definitionJSON := range *definitionsJSONList {
		availableTasks := []string{}
		for _, availableTask := range definitionJSON.(map[string]interface{})["available_tasks"].([]interface{}) {
			availableTasks = append(availableTasks, availableTask.(string))
		}
		definitionJSONBytes, err := json.Marshal(definitionJSON)
		if err != nil {
			return err
		}
		def := &pipelinePB.OperatorDefinition{}
		err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(definitionJSONBytes, def)
		if err != nil {
			return err
		}

		def.Spec.ComponentSpecification, err = o.generateComponentSpec(def.Title, availableTasks)
		if err != nil {
			return err
		}

		def.Spec.OpenapiSpecifications, err = o.generateOpenAPISpecs(def.Title, availableTasks)
		if err != nil {
			return err
		}

		err = o.addDefinition(def)
		if err != nil {
			return err
		}
	}

	return nil
}

// AddOperatorDefinition adds a operator definition to the operator
func (o *Operator) AddOperatorDefinition(def *pipelinePB.OperatorDefinition) error {
	def.Name = fmt.Sprintf("operator-definitions/%s", def.Id)
	err := o.addDefinition(def)
	if err != nil {
		return err
	}
	return nil
}

// ListOperatorDefinitions returns the list of operator definitions under this operator
func (o *Operator) ListOperatorDefinitions() []*pipelinePB.OperatorDefinition {
	compDefs := o.Component.listDefinitions()
	defs := []*pipelinePB.OperatorDefinition{}
	for _, compDef := range compDefs {
		defs = append(defs, compDef.(*pipelinePB.OperatorDefinition))
	}
	return defs
}

// GetOperatorDefinitionByUID returns the operator definition by definition uid
func (o *Operator) GetOperatorDefinitionByUID(defUID uuid.UUID) (*pipelinePB.OperatorDefinition, error) {
	def, err := o.Component.getDefinitionByUID(defUID)
	if err != nil {
		return nil, err
	}
	return def.(*pipelinePB.OperatorDefinition), nil
}

// GetOperatorDefinitionByID returns the operator definition by definition id
func (o *Operator) GetOperatorDefinitionByID(defID string) (*pipelinePB.OperatorDefinition, error) {
	def, err := o.Component.getDefinitionByID(defID)
	if err != nil {
		return nil, err
	}
	return def.(*pipelinePB.OperatorDefinition), nil
}
