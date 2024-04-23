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

// IConnector is the interface that all connectors need to implement
type IConnector interface {
	IComponent

	LoadConnectorDefinition(definitionJSON []byte, tasksJSON []byte, additionalJSONBytes map[string][]byte) error
	GetConnectorDefinition(sysVars map[string]any, component *pipelinePB.ConnectorComponent) (*pipelinePB.ConnectorDefinition, error)

	CreateExecution(sysVars map[string]any, connection *structpb.Struct, task string) (*ExecutionWrapper, error)
	Test(sysVars map[string]any, connection *structpb.Struct) error

	IsCredentialField(target string) bool
}

// Connector is the base struct for all connectors
type BaseConnector struct {
	Logger       *zap.Logger
	UsageHandler UsageHandler

	taskInputSchemas  map[string]string
	taskOutputSchemas map[string]string

	definition       *pipelinePB.ConnectorDefinition
	credentialFields []string
}

type IConnectorExecution interface {
	IExecution

	GetConnector() IConnector
	GetConnection() *structpb.Struct
}

type BaseConnectorExecution struct {
	Connector       IConnector
	SystemVariables map[string]any
	Connection      *structpb.Struct
	Task            string
}

func (c *BaseConnector) GetID() string {
	return c.definition.Id
}

func (c *BaseConnector) GetUID() uuid.UUID {
	return uuid.FromStringOrNil(c.definition.Uid)
}

func (c *BaseConnector) GetLogger() *zap.Logger {
	return c.Logger
}
func (c *BaseConnector) GetUsageHandler() UsageHandler {
	return c.UsageHandler
}
func (c *BaseConnector) GetConnectorDefinition(sysVars map[string]any, component *pipelinePB.ConnectorComponent) (*pipelinePB.ConnectorDefinition, error) {
	return c.definition, nil
}

func (c *BaseConnector) GetTaskInputSchemas() map[string]string {
	return c.taskInputSchemas
}
func (c *BaseConnector) GetTaskOutputSchemas() map[string]string {
	return c.taskOutputSchemas
}

// LoadConnectorDefinition loads the connector definitions from json files
func (c *BaseConnector) LoadConnectorDefinition(definitionJSONBytes []byte, tasksJSONBytes []byte, additionalJSONBytes map[string][]byte) error {
	var err error
	var definitionJSON any

	c.credentialFields = []string{}

	err = json.Unmarshal(definitionJSONBytes, &definitionJSON)
	if err != nil {
		return err
	}
	renderedTasksJSON, nil := RenderJSON(tasksJSONBytes, additionalJSONBytes)
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
	c.definition.Spec.ComponentSpecification, err = generateComponentSpec(c.definition.Title, tasks, taskStructs)
	if err != nil {
		return err
	}

	b, err := protojson.Marshal(c.definition.Spec)
	if err != nil {
		return err
	}
	connectionSpecification := &structpb.Struct{}
	err = protojson.Unmarshal(b, connectionSpecification)
	if err != nil {
		return err
	}

	connection, err := c.refineResourceSpec(connectionSpecification)
	if err != nil {
		return err
	}

	connectionPropStruct := &structpb.Struct{Fields: map[string]*structpb.Value{}}
	connectionPropStruct.Fields["connection"] = structpb.NewStructValue(connection)
	c.definition.Spec.ComponentSpecification.Fields["properties"] = structpb.NewStructValue(connectionPropStruct)

	c.definition.Spec.DataSpecifications, err = generateDataSpecs(taskStructs)
	if err != nil {
		return err
	}

	c.initCredentialField(c.definition)

	return nil

}

func (c *BaseConnector) refineResourceSpec(resourceSpec *structpb.Struct) (*structpb.Struct, error) {

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

// IsCredentialField checks if the target field is credential field
func (c *BaseConnector) IsCredentialField(target string) bool {
	for _, field := range c.credentialFields {
		if target == field {
			return true
		}
	}
	return false
}

// ListCredentialField lists the credential fields by definition id
func (c *BaseConnector) ListCredentialField() ([]string, error) {
	return c.credentialFields, nil
}

func (c *BaseConnector) initCredentialField(def *pipelinePB.ConnectorDefinition) {
	if c.credentialFields == nil {
		c.credentialFields = []string{}
	}
	credentialFields := []string{}
	connection := def.Spec.GetComponentSpecification().GetFields()["properties"].GetStructValue().GetFields()["connection"].GetStructValue()
	credentialFields = c.traverseCredentialField(connection.GetFields()["properties"], "", credentialFields)
	if l, ok := connection.GetFields()["oneOf"]; ok {
		for _, v := range l.GetListValue().Values {
			credentialFields = c.traverseCredentialField(v.GetStructValue().GetFields()["properties"], "", credentialFields)
		}
	}
	c.credentialFields = credentialFields
}

func (c *BaseConnector) traverseCredentialField(input *structpb.Value, prefix string, credentialFields []string) []string {
	for key, v := range input.GetStructValue().GetFields() {
		if isCredential, ok := v.GetStructValue().GetFields()["instillCredentialField"]; ok {
			if isCredential.GetBoolValue() || isCredential.GetStringValue() == "true" {
				credentialFields = append(credentialFields, fmt.Sprintf("%s%s", prefix, key))
			}
		}
		if tp, ok := v.GetStructValue().GetFields()["type"]; ok {
			if tp.GetStringValue() == "object" {
				if l, ok := v.GetStructValue().GetFields()["oneOf"]; ok {
					for _, v := range l.GetListValue().Values {
						credentialFields = c.traverseCredentialField(v.GetStructValue().GetFields()["properties"], fmt.Sprintf("%s%s.", prefix, key), credentialFields)
					}
				}
				credentialFields = c.traverseCredentialField(v.GetStructValue().GetFields()["properties"], fmt.Sprintf("%s%s.", prefix, key), credentialFields)
			}

		}
	}

	return credentialFields
}

func (e *BaseConnectorExecution) GetTask() string {
	return e.Task
}
func (e *BaseConnectorExecution) GetConnector() IConnector {
	return e.Connector
}
func (e *BaseConnectorExecution) GetConnection() *structpb.Struct {
	return e.Connection
}
func (e *BaseConnectorExecution) GetSystemVariables() map[string]any {
	return e.SystemVariables
}
func (e *BaseConnectorExecution) GetLogger() *zap.Logger {
	return e.Connector.GetLogger()
}
func (e *BaseConnectorExecution) GetUsageHandler() UsageHandler {
	return e.Connector.GetUsageHandler()
}
func (e *BaseConnectorExecution) GetTaskInputSchema() string {
	return e.Connector.GetTaskInputSchemas()[e.Task]
}
func (e *BaseConnectorExecution) GetTaskOutputSchema() string {
	return e.Connector.GetTaskOutputSchemas()[e.Task]
}
