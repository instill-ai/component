package base

import (
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

// `IConnector` define the function interface for all connectors.
type IConnector interface {

	// Functions that shared for all connectors
	// Add connector definition
	AddConnectorDefinition(uid uuid.UUID, id string, def *connectorPB.ConnectorDefinition) error

	// Get the map of connector definitions under this connector
	GetConnectorDefinitionMap() map[uuid.UUID]*connectorPB.ConnectorDefinition
	// Get the connector definition by definition uid
	GetConnectorDefinitionByUid(defUid uuid.UUID) (*connectorPB.ConnectorDefinition, error)
	// Get the connector definition by definition id
	GetConnectorDefinitionById(defId string) (*connectorPB.ConnectorDefinition, error)
	// Get the list of connector definitions under this connector
	ListConnectorDefinitions() []*connectorPB.ConnectorDefinition
	// Get the list of connector definitions uuids
	ListConnectorDefinitionUids() []uuid.UUID
	// List the CredentialFields by definition id
	ListCredentialField(defId string) []string

	// A helper function to check the connector has this definition by uid.
	HasUid(defUid uuid.UUID) bool
	// A helper function to check the target field a.b.c is credential
	IsCredentialField(defId string, target string) bool

	// Functions that need to be implenmented in connector implenmentation
	// Create a connection by defition uid and connector configuration
	CreateConnection(defUid uuid.UUID, connConfig *structpb.Struct, logger *zap.Logger) (IConnection, error)
}

type BaseConnector struct {
	// Store all the connector defintions in the connector
	definitionMapByUid map[uuid.UUID]*connectorPB.ConnectorDefinition
	definitionMapById  map[string]*connectorPB.ConnectorDefinition

	// Used for ordered
	definitionUids []uuid.UUID

	// Logger
	Logger *zap.Logger
}

type IConnection interface {
	// Functions that shared for all connectors
	// Validate the input and output format
	ValidateInput(data []*structpb.Struct, task string) error
	ValidateOutput(data []*structpb.Struct, task string) error

	// Functions that need to be implenmented in connector implenmentation
	// Execute
	Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error)
	// Test connection
	Test() (connectorPB.ConnectorResource_State, error)
}

type BaseConnection struct {
	// Logger for connection
	Logger     *zap.Logger
	DefUid     uuid.UUID
	Definition *connectorPB.ConnectorDefinition
	Config     *structpb.Struct
}

func (c *BaseConnector) AddConnectorDefinition(uid uuid.UUID, id string, def *connectorPB.ConnectorDefinition) error {
	if c.definitionMapByUid == nil {
		c.definitionMapByUid = map[uuid.UUID]*connectorPB.ConnectorDefinition{}
	}
	if c.definitionMapById == nil {
		c.definitionMapById = map[string]*connectorPB.ConnectorDefinition{}
	}
	c.definitionUids = append(c.definitionUids, uid)
	c.definitionMapByUid[uid] = def
	c.definitionMapById[id] = def
	return nil
}

func (c *BaseConnector) GetConnectorDefinitionMap() map[uuid.UUID]*connectorPB.ConnectorDefinition {
	return c.definitionMapByUid
}

func (c *BaseConnector) ListConnectorDefinitions() []*connectorPB.ConnectorDefinition {
	definitions := []*connectorPB.ConnectorDefinition{}
	for _, uid := range c.definitionUids {
		val, ok := c.definitionMapByUid[uid]
		if !ok {
			// logger
			c.Logger.Error("get connector defintion error")
		}
		definitions = append(definitions, val)
	}

	return definitions
}

func (c *BaseConnector) GetConnectorDefinitionByUid(defUid uuid.UUID) (*connectorPB.ConnectorDefinition, error) {
	val, ok := c.definitionMapByUid[defUid]
	if !ok {
		return nil, fmt.Errorf("get connector defintion error")
	}
	return val, nil
}

func (c *BaseConnector) GetConnectorDefinitionById(defId string) (*connectorPB.ConnectorDefinition, error) {

	val, ok := c.definitionMapById[defId]
	if !ok {
		return nil, fmt.Errorf("get connector defintion error")
	}
	return val, nil
}

func (c *BaseConnector) ListConnectorDefinitionUids() []uuid.UUID {
	return c.definitionUids
}

func (c *BaseConnector) HasUid(defUid uuid.UUID) bool {
	_, err := c.GetConnectorDefinitionByUid(defUid)
	return err == nil
}

func (conn *BaseConnection) ValidateInput(data []*structpb.Struct, task string) error {
	schema, err := conn.getInputSchema(task)
	if err != nil {
		return err
	}
	return conn.validate(data, string(schema))
}

func (conn *BaseConnection) ValidateOutput(data []*structpb.Struct, task string) error {
	schema, err := conn.getOutputSchema(task)
	if err != nil {
		return err
	}
	return conn.validate(data, string(schema))

}
func (conn *BaseConnection) validate(data []*structpb.Struct, jsonSchema string) error {
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

func (conn *BaseConnection) getInputSchema(task string) ([]byte, error) {

	if _, ok := conn.Definition.Spec.OpenapiSpecifications.GetFields()[task]; !ok {
		return nil, fmt.Errorf("task %s not exist", task)
	}
	walk := conn.Definition.Spec.OpenapiSpecifications.GetFields()[task]
	for _, key := range []string{"paths", "/execute", "post", "requestBody", "content", "application/json", "schema", "properties", "inputs", "items"} {
		walk = walk.GetStructValue().Fields[key]
	}
	walkBytes, err := protojson.Marshal(walk)
	return walkBytes, err
}

func (conn *BaseConnection) getOutputSchema(task string) ([]byte, error) {
	if _, ok := conn.Definition.Spec.OpenapiSpecifications.GetFields()[task]; !ok {
		return nil, fmt.Errorf("task %s not exist", task)
	}
	walk := conn.Definition.Spec.OpenapiSpecifications.GetFields()[task]
	for _, key := range []string{"paths", "/execute", "post", "responses", "200", "content", "application/json", "schema", "properties", "outputs", "items"} {
		walk = walk.GetStructValue().Fields[key]
	}
	walkBytes, err := protojson.Marshal(walk)
	return walkBytes, err
}

func (c *BaseConnector) IsCredentialField(defId string, target string) bool {
	for _, field := range c.ListCredentialField(defId) {
		if target == field {
			return true
		}
	}
	return false
}

func (c *BaseConnector) ListCredentialField(defId string) []string {
	credentialFields := []string{}
	credentialFields = c.listCredentialField(c.definitionMapById[defId].Spec.GetResourceSpecification().GetFields()["properties"], "", credentialFields)
	return credentialFields
}

func (c *BaseConnector) listCredentialField(input *structpb.Value, prefix string, credentialFields []string) []string {
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
						credentialFields = c.listCredentialField(v.GetStructValue().GetFields()["properties"], fmt.Sprintf("%s%s.", prefix, key), credentialFields)
					}
				}
				credentialFields = c.listCredentialField(v.GetStructValue().GetFields()["properties"], fmt.Sprintf("%s%s.", prefix, key), credentialFields)
			}

		}
	}

	return credentialFields
}

func ConvertFromStructpb(from *structpb.Struct, to interface{}) error {
	inputJson, err := protojson.Marshal(from)
	if err != nil {
		return err
	}

	err = json.Unmarshal(inputJson, to)
	if err != nil {
		return err
	}
	return nil
}

func ConvertToStructpb(from interface{}) (*structpb.Struct, error) {
	to := &structpb.Struct{}
	outputJson, err := json.Marshal(from)
	if err != nil {
		return nil, err
	}

	err = protojson.Unmarshal(outputJson, to)
	if err != nil {
		return nil, err
	}
	return to, nil
}
