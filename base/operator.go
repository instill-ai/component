package base

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/gofrs/uuid"
	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

// IOperator defines the methods of an operator component.
type IOperator interface {
	IComponent

	LoadOperatorDefinition(definitionJSON []byte, tasksJSON []byte, additionalJSONBytes map[string][]byte) error

	// Note: Some content in the definition JSON schema needs to be generated
	// by sysVars or component setting.
	GetOperatorDefinition(sysVars map[string]any, component *pipelinePB.OperatorComponent) (*pipelinePB.OperatorDefinition, error)

	CreateExecution(sysVars map[string]any, task string) (*ExecutionWrapper, error)
}

// Operator implements the common operator methods.
type Operator struct {
	Logger *zap.Logger

	taskInputSchemas  map[string]string
	taskOutputSchemas map[string]string

	definition *pipelinePB.OperatorDefinition
}

func (o *Operator) GetID() string {
	return o.definition.Id
}
func (o *Operator) GetUID() uuid.UUID {
	return uuid.FromStringOrNil(o.definition.Uid)
}
func (o *Operator) GetLogger() *zap.Logger {
	return o.Logger
}
func (o *Operator) GetTaskInputSchemas() map[string]string {
	return o.taskInputSchemas
}
func (o *Operator) GetTaskOutputSchemas() map[string]string {
	return o.taskOutputSchemas
}

func (o *Operator) GetOperatorDefinition(sysVars map[string]any, component *pipelinePB.OperatorComponent) (*pipelinePB.OperatorDefinition, error) {
	return o.definition, nil
}

// LoadOperatorDefinition loads the operator definitions from json files
func (o *Operator) LoadOperatorDefinition(definitionJSONBytes []byte, tasksJSONBytes []byte, additionalJSONBytes map[string][]byte) error {
	var err error
	var definitionJSON any

	err = json.Unmarshal(definitionJSONBytes, &definitionJSON)
	if err != nil {
		return err
	}
	renderedTasksJSON, err := RenderJSON(tasksJSONBytes, additionalJSONBytes)
	if err != nil {
		return nil
	}

	availableTasks := []string{}
	for _, availableTask := range definitionJSON.(map[string]interface{})["available_tasks"].([]interface{}) {
		availableTasks = append(availableTasks, availableTask.(string))
	}

	tasks, taskStructs, err := loadTasks(availableTasks, renderedTasksJSON)
	if err != nil {
		return err
	}

	o.taskInputSchemas = map[string]string{}
	o.taskOutputSchemas = map[string]string{}
	for k := range taskStructs {
		var s []byte
		s, err = protojson.Marshal(taskStructs[k].Fields["input"].GetStructValue())
		if err != nil {
			return err
		}
		o.taskInputSchemas[k] = string(s)

		s, err = protojson.Marshal(taskStructs[k].Fields["output"].GetStructValue())
		if err != nil {
			return err
		}
		o.taskOutputSchemas[k] = string(s)
	}

	o.definition = &pipelinePB.OperatorDefinition{}
	err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(definitionJSONBytes, o.definition)
	if err != nil {
		return err
	}

	o.definition.Name = fmt.Sprintf("operator-definitions/%s", o.definition.Id)
	o.definition.Tasks = tasks
	o.definition.Spec.ComponentSpecification, err = generateComponentSpec(o.definition.Title, tasks, taskStructs)
	if err != nil {
		return err
	}
	o.definition.Spec.DataSpecifications, err = generateDataSpecs(taskStructs)
	if err != nil {
		return err
	}

	return nil
}

// UsageHandlerCreator returns a no-op usage handler initializer. For the
// moment there are no use cases for usage collection in operators.
func (o *Operator) UsageHandlerCreator() UsageHandlerCreator {
	return NewNoopUsageHandler
}

// OperatorExecution implements the common methods for operator execution.
type OperatorExecution struct {
	Operator        IOperator
	SystemVariables map[string]any
	Task            string
}

func (e *OperatorExecution) GetTask() string {
	return e.Task
}
func (e *OperatorExecution) GetOperator() IOperator {
	return e.Operator
}
func (e *OperatorExecution) GetSystemVariables() map[string]any {
	return e.SystemVariables
}
func (e *OperatorExecution) GetLogger() *zap.Logger {
	return e.Operator.GetLogger()
}
func (e *OperatorExecution) GetTaskInputSchema() string {
	return e.Operator.GetTaskInputSchemas()[e.Task]
}
func (e *OperatorExecution) GetTaskOutputSchema() string {
	return e.Operator.GetTaskOutputSchemas()[e.Task]
}

// UsesSecret indicates wether the operator execution is configured with
// global secrets. Components should override this method when they have the
// ability to be executed with global secrets.
func (e *OperatorExecution) UsesSecret() bool {
	return false
}

// UsageHandlerCreator returns a function to initialize a UsageHandler.
func (e *OperatorExecution) UsageHandlerCreator() UsageHandlerCreator {
	return e.Operator.UsageHandlerCreator()
}
