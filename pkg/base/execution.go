package base

import (
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

// IExecution is the interface that all executions need to implement
type IExecution interface {
	// Functions that shared for all connectors
	// Validate the input and output format
	Validate(data []*structpb.Struct, jsonSchema string) error

	// Execute
	GetTask() string
	GetConfig() *structpb.Struct
	GetUID() uuid.UUID

	// Execute
	ExecuteWithValidation(inputs []*structpb.Struct) ([]*structpb.Struct, error)
	// execute
	Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error)
}

// Execution is the base struct for all executions
type Execution struct {
	Logger             *zap.Logger
	Component          IComponent
	ComponentExecution IExecution
	UID                uuid.UUID
	Config             *structpb.Struct
	Task               string
}

// GetUID returns the uid of the execution
func (e *Execution) GetUID() uuid.UUID {
	return e.UID
}

// GetTask returns the task of the execution
func (e *Execution) GetTask() string {
	return e.Task
}

// GetConfig returns the config of the execution
func (e *Execution) GetConfig() *structpb.Struct {
	return e.Config
}

// Validate the input and output format
func (e *Execution) Validate(data []*structpb.Struct, jsonSchema string) error {
	sch, err := jsonschema.CompileString("schema.json", jsonSchema)
	if err != nil {
		return err
	}
	for idx := range data {
		var v interface{}
		jsonData, err := protojson.Marshal(data[idx])
		if err != nil {
			return err
		}

		if err := json.Unmarshal(jsonData, &v); err != nil {
			return err
		}

		if err = sch.Validate(v); err != nil {
			return err
		}
	}
	return nil
}

// ExecuteWithValidation executes the execution with validation
func (e *Execution) ExecuteWithValidation(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	task := e.GetTask()
	if task == "" {
		keys := make([]string, 0, len(e.Component.GetTaskInputSchemas()))
		for k := range e.Component.GetTaskInputSchemas() {
			keys = append(keys, k)
		}
		if len(keys) != 1 {
			return nil, fmt.Errorf("must specify a task")
		}
		task = keys[0]
	}

	if _, ok := e.Component.GetTaskInputSchemas()[task]; !ok {
		return nil, fmt.Errorf("no task %s", e.GetTask())
	}

	if err := e.Validate(inputs, e.Component.GetTaskInputSchemas()[task]); err != nil {
		return nil, err
	}

	outputs, err := e.ComponentExecution.Execute(inputs)
	if err != nil {
		return nil, err
	}

	if err := e.Validate(outputs, e.Component.GetTaskOutputSchemas()[task]); err != nil {
		return nil, err
	}
	return outputs, err
}

// CreateExecutionHelper creates a new execution
func CreateExecutionHelper(e IExecution, comp IComponent, defUID uuid.UUID, task string, config *structpb.Struct, logger *zap.Logger) Execution {

	baseExecution := Execution{
		Logger:             logger,
		Component:          comp,
		ComponentExecution: e,
		UID:                defUID,
		Config:             config,
		Task:               task,
	}

	return baseExecution
}
