package base

import (
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

// IConnector defines the methods of a connector component.
type IConnector interface {
	IComponent

	LoadConnectorDefinition(definitionJSON []byte, tasksJSON []byte, additionalJSONBytes map[string][]byte) error

	// Note: Some content in the definition JSON schema needs to be generated
	// by sysVars or component setting.
	GetConnectorDefinition(sysVars map[string]any, component *pipelinePB.ConnectorComponent) (*pipelinePB.ConnectorDefinition, error)

	CreateExecution(sysVars map[string]any, connection *structpb.Struct, task string) (*ExecutionWrapper, error)
	Test(sysVars map[string]any, connection *structpb.Struct) error

	IsSecretField(target string) bool
}

// Connector implements the common connector methods.
type Connector struct {
	Logger *zap.Logger

	taskInputSchemas  map[string]string
	taskOutputSchemas map[string]string

	definition   *pipelinePB.ConnectorDefinition
	secretFields []string
}

func (c *Connector) GetID() string {
	return c.definition.Id
}

func (c *Connector) GetUID() uuid.UUID {
	return uuid.FromStringOrNil(c.definition.Uid)
}

func (c *Connector) GetLogger() *zap.Logger {
	return c.Logger
}
func (c *Connector) GetConnectorDefinition(sysVars map[string]any, component *pipelinePB.ConnectorComponent) (*pipelinePB.ConnectorDefinition, error) {
	return c.definition, nil
}

func (c *Connector) GetTaskInputSchemas() map[string]string {
	return c.taskInputSchemas
}
func (c *Connector) GetTaskOutputSchemas() map[string]string {
	return c.taskOutputSchemas
}

// LoadConnectorDefinition loads the connector definitions from json files
func (c *Connector) LoadConnectorDefinition(definitionJSONBytes []byte, tasksJSONBytes []byte, additionalJSONBytes map[string][]byte) error {
	var err error
	var definitionJSON any

	c.secretFields = []string{}

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

	c.taskInputSchemas = map[string]string{}
	c.taskOutputSchemas = map[string]string{}
	for k := range taskStructs {
		var s []byte
		s, err = protojson.Marshal(taskStructs[k].Fields["input"].GetStructValue())
		if err != nil {
			return err
		}
		c.taskInputSchemas[k] = string(s)

		s, err = protojson.Marshal(taskStructs[k].Fields["output"].GetStructValue())
		if err != nil {
			return err
		}
		c.taskOutputSchemas[k] = string(s)
	}

	c.definition = &pipelinePB.ConnectorDefinition{}
	err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(definitionJSONBytes, c.definition)
	if err != nil {
		return err
	}

	c.definition.Name = fmt.Sprintf("connector-definitions/%s", c.definition.Id)
	c.definition.Tasks = tasks
	if c.definition.Spec == nil {
		c.definition.Spec = &pipelinePB.ConnectorSpec{}
	}
	c.definition.Spec.ComponentSpecification, err = generateComponentSpec(c.definition.Title, tasks, taskStructs)
	if err != nil {
		return err
	}

	raw := &structpb.Struct{}
	err = protojson.Unmarshal(definitionJSONBytes, raw)
	if err != nil {
		return err
	}
	// TODO: Avoid using structpb traversal here.
	if _, ok := raw.Fields["spec"]; ok {
		if v, ok := raw.Fields["spec"].GetStructValue().Fields["connection_specification"]; ok {
			connection, err := c.refineResourceSpec(v.GetStructValue())
			if err != nil {
				return err
			}
			connectionPropStruct := &structpb.Struct{Fields: map[string]*structpb.Value{}}
			connectionPropStruct.Fields["connection"] = structpb.NewStructValue(connection)
			c.definition.Spec.ComponentSpecification.Fields["properties"] = structpb.NewStructValue(connectionPropStruct)
		}
	}

	c.definition.Spec.DataSpecifications, err = generateDataSpecs(taskStructs)
	if err != nil {
		return err
	}

	c.initSecretField(c.definition)

	return nil

}

func (c *Connector) refineResourceSpec(resourceSpec *structpb.Struct) (*structpb.Struct, error) {

	spec := proto.Clone(resourceSpec).(*structpb.Struct)
	if _, ok := spec.Fields["instillShortDescription"]; !ok {
		spec.Fields["instillShortDescription"] = structpb.NewStringValue(spec.Fields["description"].GetStringValue())
	}

	if _, ok := spec.Fields["properties"]; ok {
		for k, v := range spec.Fields["properties"].GetStructValue().AsMap() {
			s, err := structpb.NewStruct(v.(map[string]interface{}))
			if err != nil {
				return nil, err
			}
			converted, err := c.refineResourceSpec(s)
			if err != nil {
				return nil, err
			}
			spec.Fields["properties"].GetStructValue().Fields[k] = structpb.NewStructValue(converted)

		}
	}
	if _, ok := spec.Fields["patternProperties"]; ok {
		for k, v := range spec.Fields["patternProperties"].GetStructValue().AsMap() {
			s, err := structpb.NewStruct(v.(map[string]interface{}))
			if err != nil {
				return nil, err
			}
			converted, err := c.refineResourceSpec(s)
			if err != nil {
				return nil, err
			}
			spec.Fields["patternProperties"].GetStructValue().Fields[k] = structpb.NewStructValue(converted)

		}
	}
	for _, target := range []string{"allOf", "anyOf", "oneOf"} {
		if _, ok := spec.Fields[target]; ok {
			for idx, item := range spec.Fields[target].GetListValue().AsSlice() {
				s, err := structpb.NewStruct(item.(map[string]interface{}))
				if err != nil {
					return nil, err
				}
				converted, err := c.refineResourceSpec(s)
				if err != nil {
					return nil, err
				}
				spec.Fields[target].GetListValue().AsSlice()[idx] = structpb.NewStructValue(converted)
			}
		}
	}

	return spec, nil
}

// IsSecretField checks if the target field is secret field
func (c *Connector) IsSecretField(target string) bool {
	for _, field := range c.secretFields {
		if target == field {
			return true
		}
	}
	return false
}

// ListSecretFields lists the secret fields by definition id
func (c *Connector) ListSecretFields() ([]string, error) {
	return c.secretFields, nil
}

func (c *Connector) initSecretField(def *pipelinePB.ConnectorDefinition) {
	if c.secretFields == nil {
		c.secretFields = []string{}
	}
	secretFields := []string{}
	connection := def.Spec.GetComponentSpecification().GetFields()["properties"].GetStructValue().GetFields()["connection"].GetStructValue()
	secretFields = c.traverseSecretField(connection.GetFields()["properties"], "", secretFields)
	if l, ok := connection.GetFields()["oneOf"]; ok {
		for _, v := range l.GetListValue().Values {
			secretFields = c.traverseSecretField(v.GetStructValue().GetFields()["properties"], "", secretFields)
		}
	}
	c.secretFields = secretFields
}

func (c *Connector) traverseSecretField(input *structpb.Value, prefix string, secretFields []string) []string {
	for key, v := range input.GetStructValue().GetFields() {
		if isSecret, ok := v.GetStructValue().GetFields()["instillSecret"]; ok {
			if isSecret.GetBoolValue() || isSecret.GetStringValue() == "true" {
				secretFields = append(secretFields, fmt.Sprintf("%s%s", prefix, key))
			}
		}
		if tp, ok := v.GetStructValue().GetFields()["type"]; ok {
			if tp.GetStringValue() == "object" {
				if l, ok := v.GetStructValue().GetFields()["oneOf"]; ok {
					for _, v := range l.GetListValue().Values {
						secretFields = c.traverseSecretField(v.GetStructValue().GetFields()["properties"], fmt.Sprintf("%s%s.", prefix, key), secretFields)
					}
				}
				secretFields = c.traverseSecretField(v.GetStructValue().GetFields()["properties"], fmt.Sprintf("%s%s.", prefix, key), secretFields)
			}

		}
	}

	return secretFields
}

// UsageHandlerCreator returns a function to initialize a UsageHandler.
func (c *Connector) UsageHandlerCreator() UsageHandlerCreator {
	return NewNoopUsageHandler
}

// ConnectorExecution implements the common methods for connector
// execution.
type ConnectorExecution struct {
	Connector       IConnector
	SystemVariables map[string]any
	Connection      *structpb.Struct
	Task            string
}

func (e *ConnectorExecution) GetTask() string {
	return e.Task
}
func (e *ConnectorExecution) GetConnector() IConnector {
	return e.Connector
}
func (e *ConnectorExecution) GetConnection() *structpb.Struct {
	return e.Connection
}
func (e *ConnectorExecution) GetSystemVariables() map[string]any {
	return e.SystemVariables
}
func (e *ConnectorExecution) GetLogger() *zap.Logger {
	return e.Connector.GetLogger()
}
func (e *ConnectorExecution) GetTaskInputSchema() string {
	return e.Connector.GetTaskInputSchemas()[e.Task]
}
func (e *ConnectorExecution) GetTaskOutputSchema() string {
	return e.Connector.GetTaskOutputSchemas()[e.Task]
}

// UsesSecret indicates wether the connector execution is configured with
// global secrets. Components should override this method when they have the
// ability to be executed with global secrets.
func (e *ConnectorExecution) UsesSecret() bool {
	return false
}

// UsageHandlerCreator returns a function to initialize a UsageHandler.
func (e *ConnectorExecution) UsageHandlerCreator() UsageHandlerCreator {
	return e.Connector.UsageHandlerCreator()
}
