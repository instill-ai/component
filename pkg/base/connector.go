package base

import (
	"fmt"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

// `IConnector` define the function interface for all connectors.
type IConnector interface {

	// Functions that shared for all connectors
	// Add connector definition
	AddConnectorDefinition(uid uuid.UUID, id string, def IDefinition) error

	// Get the map of connector definitions under this connector
	GetConnectorDefinitionMap() map[uuid.UUID]IDefinition
	// Get the connector definition by definition uid
	GetConnectorDefinitionByUid(defUid uuid.UUID) (IDefinition, error)
	// Get the connector definition by definition uid
	GetConnectorDefinitionById(defId string) (IDefinition, error)
	// Get the list of connector definitions under this connector
	ListConnectorDefinitions() []IDefinition
	// Get the list of connector definitions uuids
	ListConnectorDefinitionUids() []uuid.UUID

	// A helper function to check the connector has this definition by uid.
	HasUid(defUid uuid.UUID) bool

	// Functions that need to be implenmented in connector implenmentation
	// Create a connection by defition uid and connector configuration
	CreateConnection(defUid uuid.UUID, config *structpb.Struct, logger *zap.Logger) (IConnection, error)
}

type BaseConnector struct {
	// Store all the connector defintions in the connector
	definitionMapByUid map[uuid.UUID]IDefinition
	definitionMapById  map[string]IDefinition

	// Used for ordered
	definitionUids []uuid.UUID

	// Logger
	Logger *zap.Logger
}

type IConnection interface {
	// Functions that shared for all connectors
	// Validate the input format
	Validate(input interface{}) error

	// Functions that need to be implenmented in connector implenmentation
	// Execute
	Execute(input []*connectorPB.DataPayload) ([]*connectorPB.DataPayload, error)
	// Test connection
	Test() (connectorPB.Connector_State, error)
}

type IDefinition interface {
	GetId() string
	GetUid() string
	GetConnectorDefinition() *connectorPB.ConnectorDefinition
}

type BaseConnection struct {
	// Logger for connection
	Logger *zap.Logger
}

func (c *BaseConnector) AddConnectorDefinition(uid uuid.UUID, id string, def IDefinition) error {
	if c.definitionMapByUid == nil {
		c.definitionMapByUid = map[uuid.UUID]IDefinition{}
	}
	if c.definitionMapById == nil {
		c.definitionMapById = map[string]IDefinition{}
	}
	c.definitionUids = append(c.definitionUids, uid)
	c.definitionMapByUid[uid] = def
	c.definitionMapById[id] = def
	return nil
}

func (c *BaseConnector) GetConnectorDefinitionMap() map[uuid.UUID]IDefinition {
	return c.definitionMapByUid
}

func (c *BaseConnector) ListConnectorDefinitions() []IDefinition {
	definitions := []IDefinition{}
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

func (c *BaseConnector) GetConnectorDefinitionByUid(defUid uuid.UUID) (IDefinition, error) {
	val, ok := c.definitionMapByUid[defUid]
	if !ok {
		return nil, fmt.Errorf("get connector defintion error")
	}
	return val, nil
}

func (c *BaseConnector) GetConnectorDefinitionById(defId string) (IDefinition, error) {

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

func (conn *BaseConnection) Validate(input interface{}) error {
	// validate by vdp-protocol
	return nil
}
