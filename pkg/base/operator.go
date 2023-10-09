package base

import (
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid"
	"google.golang.org/protobuf/encoding/protojson"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1alpha"
)

// `IOperator` define the function interface for all operators.
type IOperator interface {
	IComponent

	// Functions that shared for all operators
	// Load operator definitions from json files
	LoadOperatorDefinitions(definitionsJson []byte, tasksJson []byte) error
	// Add definition
	AddOperatorDefinition(def *pipelinePB.OperatorDefinition) error
	// Get the operator definition by definition uid
	GetOperatorDefinitionByUID(defUID uuid.UUID) (*pipelinePB.OperatorDefinition, error)
	// Get the operator definition by definition id
	GetOperatorDefinitionByID(defID string) (*pipelinePB.OperatorDefinition, error)
	// Get the list of operator definitions under this operator
	ListOperatorDefinitions() []*pipelinePB.OperatorDefinition
}

type Operator struct {
	Component
}

func (o *Operator) LoadOperatorDefinitions(definitionsJsonBytes []byte, tasksJsonBytes []byte) error {
	var err error
	definitionsJsonList := &[]interface{}{}

	err = json.Unmarshal(definitionsJsonBytes, definitionsJsonList)
	if err != nil {
		return err
	}
	err = o.Component.loadTasks(tasksJsonBytes)
	if err != nil {
		return err
	}

	for _, definitionJson := range *definitionsJsonList {
		availableTasks := []string{}
		for _, availableTask := range definitionJson.(map[string]interface{})["available_tasks"].([]interface{}) {
			availableTasks = append(availableTasks, availableTask.(string))
		}
		definitionJsonBytes, err := json.Marshal(definitionJson)
		if err != nil {
			return err
		}
		def := &pipelinePB.OperatorDefinition{}
		err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(definitionJsonBytes, def)
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

func (o *Operator) AddOperatorDefinition(def *pipelinePB.OperatorDefinition) error {
	def.Name = fmt.Sprintf("operator-definitions/%s", def.Id)
	err := o.addDefinition(def)
	if err != nil {
		return err
	}
	return nil
}

func (o *Operator) ListOperatorDefinitions() []*pipelinePB.OperatorDefinition {
	compDefs := o.Component.listDefinitions()
	defs := []*pipelinePB.OperatorDefinition{}
	for _, compDef := range compDefs {
		defs = append(defs, compDef.(*pipelinePB.OperatorDefinition))
	}
	return defs
}

func (o *Operator) GetOperatorDefinitionByUID(defUID uuid.UUID) (*pipelinePB.OperatorDefinition, error) {
	def, err := o.Component.getDefinitionByUID(defUID)
	if err != nil {
		return nil, err
	}
	return def.(*pipelinePB.OperatorDefinition), nil
}

func (o *Operator) GetOperatorDefinitionByID(defID string) (*pipelinePB.OperatorDefinition, error) {
	def, err := o.Component.getDefinitionByID(defID)
	if err != nil {
		return nil, err
	}
	return def.(*pipelinePB.OperatorDefinition), nil
}
