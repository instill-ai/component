package base

import (
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid"
	"google.golang.org/protobuf/encoding/protojson"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

// IOperator is the interface that all operators need to implement
type IOperator interface {
	IComponent

	// Functions that shared for all operators
	// Load operator definitions from json files, the additionalJSONBytes is only needed when you reference in-memory json file
	LoadOperatorDefinition(definitionJSON []byte, tasksJSON []byte, additionalJSONBytes map[string][]byte) error
	// Add definition
	AddOperatorDefinition(def *pipelinePB.OperatorDefinition) error
	// Get the operator definition by definition uid
	GetOperatorDefinitionByUID(defUID uuid.UUID, component *pipelinePB.OperatorComponent) (*pipelinePB.OperatorDefinition, error)
	// Get the operator definition by definition id
	GetOperatorDefinitionByID(defID string, component *pipelinePB.OperatorComponent) (*pipelinePB.OperatorDefinition, error)
	// Get the list of operator definitions under this operator
	ListOperatorDefinitions() []*pipelinePB.OperatorDefinition
}

// Operator is the base struct for all operators
type Operator struct {
	Component
}

// LoadOperatorDefinition loads the operator definitions from json files
func (o *Operator) LoadOperatorDefinition(definitionJSONBytes []byte, tasksJSONBytes []byte, additionalJSONBytes map[string][]byte) error {
	var err error
	var definitionJSON any

	err = json.Unmarshal(definitionJSONBytes, &definitionJSON)
	if err != nil {
		return err
	}
	renderedTasksJSON, nil := renderTaskJSON(tasksJSONBytes, additionalJSONBytes)
	if err != nil {
		return nil
	}

	err = o.Component.loadTasks(renderedTasksJSON)
	if err != nil {
		return err
	}

	availableTasks := []string{}
	for _, availableTask := range definitionJSON.(map[string]interface{})["available_tasks"].([]interface{}) {
		availableTasks = append(availableTasks, availableTask.(string))
	}

	def := &pipelinePB.OperatorDefinition{}
	err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(definitionJSONBytes, def)
	if err != nil {
		return err
	}

	def.Tasks = o.generateComponentTasks(availableTasks)

	def.Spec.ComponentSpecification, err = o.generateComponentSpec(def.Title, def.Tasks)
	if err != nil {
		return err
	}

	def.Spec.DataSpecifications, err = o.generateDataSpecs(def.Title, availableTasks)
	if err != nil {
		return err
	}

	err = o.addDefinition(def)
	if err != nil {
		return err
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
	opDefs := make([]*pipelinePB.OperatorDefinition, 0, len(compDefs))
	for _, d := range compDefs {
		od := d.(*pipelinePB.OperatorDefinition)
		if !od.Tombstone {
			opDefs = append(opDefs, od)
		}
	}

	return opDefs
}

// GetOperatorDefinitionByUID returns the operator definition by definition uid
func (o *Operator) GetOperatorDefinitionByUID(defUID uuid.UUID, _ /*component*/ *pipelinePB.OperatorComponent) (*pipelinePB.OperatorDefinition, error) {
	def, err := o.Component.getDefinitionByUID(defUID)
	if err != nil {
		return nil, err
	}

	od, ok := def.(*pipelinePB.OperatorDefinition)
	if !ok {
		return nil, fmt.Errorf("invalid type for operator definition UID")
	}

	return od, nil
}

// GetOperatorDefinitionByID returns the operator definition by definition id
func (o *Operator) GetOperatorDefinitionByID(defID string, _ /*component*/ *pipelinePB.OperatorComponent) (*pipelinePB.OperatorDefinition, error) {
	def, err := o.Component.getDefinitionByID(defID)
	if err != nil {
		return nil, err
	}

	od, ok := def.(*pipelinePB.OperatorDefinition)
	if !ok {
		return nil, fmt.Errorf("invalid type for operator definition ID")
	}

	return od, nil
}
