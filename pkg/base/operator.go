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

// `IOperator` define the function interface for all operators.
type IOperator interface {

	// Functions that shared for all operators
	// Add operator definition
	AddOperatorDefinition(uid uuid.UUID, id string, def *connectorPB.ConnectorDefinition) error

	// Get the map of operator definitions under this operator
	GetOperatorDefinitionMap() map[uuid.UUID]*connectorPB.ConnectorDefinition
	// Get the operator definition by definition uid
	GetOperatorDefinitionByUid(defUid uuid.UUID) (*connectorPB.ConnectorDefinition, error)
	// Get the operator definition by definition id
	GetOperatorDefinitionById(defId string) (*connectorPB.ConnectorDefinition, error)
	// Get the list of operator definitions under this operator
	ListOperatorDefinitions() []*connectorPB.ConnectorDefinition
	// Get the list of operator definitions uuids
	ListOperatorDefinitionUids() []uuid.UUID
	// List the CredentialFields by definition id
	ListCredentialField(defId string) []string

	// A helper function to check the operator has this definition by uid.
	HasUid(defUid uuid.UUID) bool
	// A helper function to check the target field a.b.c is credential
	IsCredentialField(defId string, target string) bool

	// Functions that need to be implemented in operator implenmentation
	// Create a operation by defition uid and operator configuration
	CreateOperation(defUid uuid.UUID, connConfig *structpb.Struct, logger *zap.Logger) (IOperation, error)
}

type BaseOperator struct {
	// Store all the operator defintions in the operator
	definitionMapByUid map[uuid.UUID]*connectorPB.ConnectorDefinition
	definitionMapById  map[string]*connectorPB.ConnectorDefinition

	// Used for ordered
	definitionUids []uuid.UUID

	// Logger
	Logger *zap.Logger
}

type IOperation interface {
	// Functions that shared for all operators
	// Validate the input and output format
	ValidateInput(data []*structpb.Struct, task string) error
	ValidateOutput(data []*structpb.Struct, task string) error

	// Functions that need to be implenmented in operator implenmentation
	// Execute
	Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error)
	// Test operation
	Test() (connectorPB.ConnectorResource_State, error)
}

type BaseOperation struct {
	// Logger for connection
	Logger     *zap.Logger
	DefUid     uuid.UUID
	Definition *connectorPB.ConnectorDefinition
	Config     *structpb.Struct
}

func (c *BaseOperator) AddOperatorDefinition(uid uuid.UUID, id string, def *connectorPB.ConnectorDefinition) error {
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

func (c *BaseOperator) GetOperatorDefinitionMap() map[uuid.UUID]*connectorPB.ConnectorDefinition {
	return c.definitionMapByUid
}

func (c *BaseOperator) ListOperatorDefinitions() []*connectorPB.ConnectorDefinition {
	definitions := []*connectorPB.ConnectorDefinition{}
	for _, uid := range c.definitionUids {
		val, ok := c.definitionMapByUid[uid]
		if !ok {
			// logger
			c.Logger.Error("get operator defintion error")
		}
		definitions = append(definitions, val)
	}

	return definitions
}

func (c *BaseOperator) GetOperatorDefinitionByUid(defUid uuid.UUID) (*connectorPB.ConnectorDefinition, error) {
	val, ok := c.definitionMapByUid[defUid]
	if !ok {
		return nil, fmt.Errorf("get operator defintion error")
	}
	return val, nil
}

func (c *BaseOperator) GetOperatorDefinitionById(defId string) (*connectorPB.ConnectorDefinition, error) {
	val, ok := c.definitionMapById[defId]
	if !ok {
		return nil, fmt.Errorf("get operator defintion error")
	}
	return val, nil
}

func (c *BaseOperator) ListOperatorDefinitionUids() []uuid.UUID {
	return c.definitionUids
}

func (c *BaseOperator) HasUid(defUid uuid.UUID) bool {
	_, err := c.GetOperatorDefinitionByUid(defUid)
	return err == nil
}

func (conn *BaseOperation) ValidateInput(data []*structpb.Struct, task string) error {
	schema, err := conn.getInputSchema(task)
	if err != nil {
		return err
	}
	return conn.validate(data, string(schema))
}

func (conn *BaseOperation) ValidateOutput(data []*structpb.Struct, task string) error {
	schema, err := conn.getOutputSchema(task)
	if err != nil {
		return err
	}
	return conn.validate(data, string(schema))

}
func (conn *BaseOperation) validate(data []*structpb.Struct, jsonSchema string) error {
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

func (conn *BaseOperation) getInputSchema(task string) ([]byte, error) {

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

func (conn *BaseOperation) getOutputSchema(task string) ([]byte, error) {
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

func (c *BaseOperator) IsCredentialField(defId string, target string) bool {
	for _, field := range c.ListCredentialField(defId) {
		if target == field {
			return true
		}
	}
	return false
}

func (c *BaseOperator) ListCredentialField(defId string) []string {
	credentialFields := []string{}
	credentialFields = c.listCredentialField(c.definitionMapById[defId].Spec.GetResourceSpecification().GetFields()["properties"], "", credentialFields)
	return credentialFields
}

func (c *BaseOperator) listCredentialField(input *structpb.Value, prefix string, credentialFields []string) []string {
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
