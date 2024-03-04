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

	// Functions that need to be implemented for all connectors
	// Test connection
	Test(defUID uuid.UUID, config *structpb.Struct, logger *zap.Logger) (pipelinePB.Connector_State, error)

	// Functions that shared for all connectors
	// Load connector definitions from json files
	LoadConnectorDefinitions(definitionsJSON []byte, tasksJSON []byte, additionalJSONBytes map[string][]byte) error
	// Add definition
	AddConnectorDefinition(def *pipelinePB.ConnectorDefinition) error
	// Get the connector definition by definition uid
	GetConnectorDefinitionByUID(defUID uuid.UUID, resourceConfig *structpb.Struct, componentConfig *structpb.Struct) (*pipelinePB.ConnectorDefinition, error)
	// Get the connector definition by definition id
	GetConnectorDefinitionByID(defID string, resourceConfig *structpb.Struct, componentConfig *structpb.Struct) (*pipelinePB.ConnectorDefinition, error)
	// Get the list of connector definitions under this connector
	ListConnectorDefinitions() []*pipelinePB.ConnectorDefinition

	// List the CredentialFields by definition id
	ListCredentialField(defID string) ([]string, error)
	// A helper function to check the target field a.b.c is credential
	IsCredentialField(defID string, target string) bool
}

// Connector is the base struct for all connectors
type Connector struct {
	Component

	// TODO: we can store the instillCredentialFields here when LoadConnectorDefinitions
	credentialFields map[string][]string
}

// LoadConnectorDefinitions loads the connector definitions from json files
func (c *Connector) LoadConnectorDefinitions(definitionsJSONBytes []byte, tasksJSONBytes []byte, additionalJSONBytes map[string][]byte) error {
	var err error
	definitionsJSONList := &[]interface{}{}
	c.credentialFields = map[string][]string{}

	err = json.Unmarshal(definitionsJSONBytes, definitionsJSONList)
	if err != nil {
		return err
	}
	renderedTasksJSON, nil := renderTaskJSON(tasksJSONBytes, additionalJSONBytes)
	if err != nil {
		return nil
	}

	err = c.Component.loadTasks(renderedTasksJSON)
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
		def := &pipelinePB.ConnectorDefinition{}
		err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(definitionJSONBytes, def)
		if err != nil {
			return err
		}
		def.Spec.ResourceSpecification, err = c.refineResourceSpec(def.Spec.ResourceSpecification)
		if err != nil {
			return err
		}

		def.Tasks = c.generateComponentTasks(availableTasks)

		def.Spec.ComponentSpecification, err = c.generateComponentSpec(def.Title, def.Tasks)
		if err != nil {
			return err
		}

		def.Spec.OpenapiSpecifications, err = c.generateOpenAPISpecs(def.Title, availableTasks)
		if err != nil {
			return err
		}

		err = c.AddConnectorDefinition(def)
		if err != nil {
			return err
		}

	}

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

// AddConnectorDefinition adds a connector definition to the connector
func (c *Connector) AddConnectorDefinition(def *pipelinePB.ConnectorDefinition) error {
	def.Name = fmt.Sprintf("connector-definitions/%s", def.Id)
	err := c.addDefinition(def)
	if err != nil {
		return err
	}
	c.initCredentialField(def.Id)
	return nil
}

// ListConnectorDefinitions lists all the connector definitions
func (c *Connector) ListConnectorDefinitions() []*pipelinePB.ConnectorDefinition {
	compDefs := c.Component.listDefinitions()
	connDefs := make([]*pipelinePB.ConnectorDefinition, 0, len(compDefs))
	for _, d := range compDefs {
		cd := d.(*pipelinePB.ConnectorDefinition)
		if !cd.Tombstone {
			connDefs = append(connDefs, cd)
		}
	}

	return connDefs
}

// GetConnectorDefinitionByUID gets the connector definition by definition uid
func (c *Connector) GetConnectorDefinitionByUID(defUID uuid.UUID, resourceConfig *structpb.Struct, componentConfig *structpb.Struct) (*pipelinePB.ConnectorDefinition, error) {
	def, err := c.Component.getDefinitionByUID(defUID)
	if err != nil {
		return nil, err
	}
	return def.(*pipelinePB.ConnectorDefinition), nil
}

// GetConnectorDefinitionByID gets the connector definition by definition id
func (c *Connector) GetConnectorDefinitionByID(defID string, resourceConfig *structpb.Struct, componentConfig *structpb.Struct) (*pipelinePB.ConnectorDefinition, error) {
	def, err := c.Component.getDefinitionByID(defID)
	if err != nil {
		return nil, err
	}
	return def.(*pipelinePB.ConnectorDefinition), nil
}

// IsCredentialField checks if the target field is credential field
func (c *Connector) IsCredentialField(defID string, target string) bool {
	for _, field := range c.credentialFields[defID] {
		if target == field {
			return true
		}
	}
	return false
}

// ListCredentialField lists the credential fields by definition id
func (c *Connector) ListCredentialField(defID string) ([]string, error) {
	return c.credentialFields[defID], nil
}

func (c *Connector) initCredentialField(defID string) {
	if c.credentialFields == nil {
		c.credentialFields = map[string][]string{}
	}
	credentialFields := []string{}
	credentialFields = c.traverseCredentialField(c.definitionMapByID[defID].(*pipelinePB.ConnectorDefinition).Spec.GetResourceSpecification().GetFields()["properties"], "", credentialFields)
	if l, ok := c.definitionMapByID[defID].(*pipelinePB.ConnectorDefinition).Spec.GetResourceSpecification().GetFields()["oneOf"]; ok {
		for _, v := range l.GetListValue().Values {
			credentialFields = c.traverseCredentialField(v.GetStructValue().GetFields()["properties"], "", credentialFields)
		}
	}
	c.credentialFields[defID] = credentialFields
}

func (c *Connector) traverseCredentialField(input *structpb.Value, prefix string, credentialFields []string) []string {
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
