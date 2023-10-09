package base

import (
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

// `IConnector` define the function interface for all connectors.
type IConnector interface {
	IComponent

	// Functions that need to be implemented for all connectors
	// Test connection
	Test(defUID uuid.UUID, config *structpb.Struct, logger *zap.Logger) (connectorPB.ConnectorResource_State, error)

	// Functions that shared for all connectors
	// Load connector definitions from json files
	LoadConnectorDefinitions(definitionsJson []byte, tasksJson []byte) error
	// Add definition
	AddConnectorDefinition(def *connectorPB.ConnectorDefinition) error
	// Get the connector definition by definition uid
	GetConnectorDefinitionByUID(defUID uuid.UUID) (*connectorPB.ConnectorDefinition, error)
	// Get the connector definition by definition id
	GetConnectorDefinitionByID(defID string) (*connectorPB.ConnectorDefinition, error)
	// Get the list of connector definitions under this connector
	ListConnectorDefinitions() []*connectorPB.ConnectorDefinition

	// List the CredentialFields by definition id
	ListCredentialField(defID string) ([]string, error)
	// A helper function to check the target field a.b.c is credential
	IsCredentialField(defID string, target string) bool
}

type Connector struct {
	Component

	// TODO: we can store the credential_fields here when LoadConnectorDefinitions
	credentialFields map[string][]string
}

func (c *Connector) LoadConnectorDefinitions(definitionsJsonBytes []byte, tasksJsonBytes []byte) error {
	var err error
	definitionsJsonList := &[]interface{}{}
	c.credentialFields = map[string][]string{}

	err = json.Unmarshal(definitionsJsonBytes, definitionsJsonList)
	if err != nil {
		return err
	}
	err = c.Component.loadTasks(tasksJsonBytes)
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
		def := &connectorPB.ConnectorDefinition{}
		err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(definitionJsonBytes, def)
		if err != nil {
			return err
		}

		def.Spec.ComponentSpecification, err = c.generateComponentSpec(def.Title, availableTasks)
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

func (c *Connector) AddConnectorDefinition(def *connectorPB.ConnectorDefinition) error {
	def.Name = fmt.Sprintf("connector-definitions/%s", def.Id)
	err := c.addDefinition(def)
	if err != nil {
		return err
	}
	c.initCredentialField(def.Id)
	return nil
}

func (c *Connector) ListConnectorDefinitions() []*connectorPB.ConnectorDefinition {
	compDefs := c.Component.listDefinitions()
	defs := []*connectorPB.ConnectorDefinition{}
	for _, compDef := range compDefs {
		defs = append(defs, compDef.(*connectorPB.ConnectorDefinition))
	}
	return defs
}

func (c *Connector) GetConnectorDefinitionByUID(defUID uuid.UUID) (*connectorPB.ConnectorDefinition, error) {
	def, err := c.Component.getDefinitionByUID(defUID)
	if err != nil {
		return nil, err
	}
	return def.(*connectorPB.ConnectorDefinition), nil
}

func (c *Connector) GetConnectorDefinitionByID(defID string) (*connectorPB.ConnectorDefinition, error) {
	def, err := c.Component.getDefinitionByID(defID)
	if err != nil {
		return nil, err
	}
	return def.(*connectorPB.ConnectorDefinition), nil
}

func (c *Connector) IsCredentialField(defID string, target string) bool {
	for _, field := range c.credentialFields[defID] {
		if target == field {
			return true
		}
	}
	return false
}

func (c *Connector) ListCredentialField(defID string) ([]string, error) {
	return c.credentialFields[defID], nil
}

func (c *Connector) initCredentialField(defID string) {
	if c.credentialFields == nil {
		c.credentialFields = map[string][]string{}
	}
	credentialFields := []string{}
	credentialFields = c.traverseCredentialField(c.definitionMapByID[defID].(*connectorPB.ConnectorDefinition).Spec.GetResourceSpecification().GetFields()["properties"], "", credentialFields)
	c.credentialFields[defID] = credentialFields
}

func (c *Connector) traverseCredentialField(input *structpb.Value, prefix string, credentialFields []string) []string {
	for key, v := range input.GetStructValue().GetFields() {
		if isCredential, ok := v.GetStructValue().GetFields()["credential_field"]; ok {
			if isCredential.GetBoolValue() || isCredential.GetStringValue() == "true" {
				credentialFields = append(credentialFields, fmt.Sprintf("%s%s", prefix, key))
			}
		}
		if type_, ok := v.GetStructValue().GetFields()["type"]; ok {
			if type_.GetStringValue() == "object" {
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
