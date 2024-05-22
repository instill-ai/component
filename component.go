package component

import (
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/ai/archetypeai/v0"
	"github.com/instill-ai/component/ai/huggingface/v0"
	"github.com/instill-ai/component/ai/instill/v0"
	"github.com/instill-ai/component/ai/openai/v0"
	"github.com/instill-ai/component/ai/stabilityai/v0"
	"github.com/instill-ai/component/application/googlesearch/v0"
	"github.com/instill-ai/component/application/numbers/v0"
	"github.com/instill-ai/component/application/redis/v0"
	"github.com/instill-ai/component/application/restapi/v0"
	"github.com/instill-ai/component/application/slack/v0"
	"github.com/instill-ai/component/application/website/v0"
	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/data/bigquery/v0"
	"github.com/instill-ai/component/data/googlecloudstorage/v0"
	"github.com/instill-ai/component/data/pinecone/v0"
	"github.com/instill-ai/component/operator/base64/v0"
	"github.com/instill-ai/component/operator/document/v0"
	"github.com/instill-ai/component/operator/image/v0"
	"github.com/instill-ai/component/operator/json/v0"
	"github.com/instill-ai/component/operator/text/v0"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

var (
	once      sync.Once
	compStore *Store
)

// Store holds in-memory information about the initialized components.
type Store struct {
	operatorUIDs   []uuid.UUID
	operatorUIDMap map[uuid.UUID]*operator
	operatorIDMap  map[string]*operator

	connectorUIDs   []uuid.UUID
	connectorUIDMap map[uuid.UUID]*connector
	connectorIDMap  map[string]*connector
}

type operator struct {
	op base.IOperator
}

type connector struct {
	con base.IConnector
}

// ConnectionSecrets contains the global connection secrets of each
// implemented connector (referenced by ID). Connectors may use these secrets
// to skip the connector configuration step and have a ready-to-run
// connection.
type ConnectionSecrets map[string]map[string]any

// Init initializes the components implemented in this repository and loads
// their information to memory.
func Init(
	logger *zap.Logger,
	secrets ConnectionSecrets,
	usageHandlerCreators map[string]base.UsageHandlerCreator,
) *Store {
	baseOp := base.Operator{Logger: logger}
	baseConn := base.Connector{Logger: logger}

	once.Do(func() {
		compStore = &Store{
			operatorUIDMap:  map[uuid.UUID]*operator{},
			operatorIDMap:   map[string]*operator{},
			connectorUIDMap: map[uuid.UUID]*connector{},
			connectorIDMap:  map[string]*connector{},
		}
		compStore.ImportOperator(base64.Init(baseOp))
		compStore.ImportOperator(json.Init(baseOp))
		compStore.ImportOperator(image.Init(baseOp))
		compStore.ImportOperator(text.Init(baseOp))
		compStore.ImportOperator(document.Init(baseOp))

		{
			// StabilityAI
			conn := stabilityai.Init(baseConn)

			// Secret doesn't allow hyphens
			conn = conn.WithSecrets(secrets["stabilityai"]).
				WithUsageHandlerCreator(usageHandlerCreators[conn.GetID()])
			compStore.ImportConnector(conn)
		}

		compStore.ImportConnector(instill.Init(baseConn))
		compStore.ImportConnector(huggingface.Init(baseConn))

		{
			// OpenAI
			conn := openai.Init(baseConn)
			conn = conn.WithSecrets(secrets[conn.GetID()]).
				WithUsageHandlerCreator(usageHandlerCreators[conn.GetID()])
			compStore.ImportConnector(conn)
		}

		compStore.ImportConnector(archetypeai.Init(baseConn))
		compStore.ImportConnector(numbers.Init(baseConn))
		compStore.ImportConnector(bigquery.Init(baseConn))
		compStore.ImportConnector(googlecloudstorage.Init(baseConn))
		compStore.ImportConnector(googlesearch.Init(baseConn))
		compStore.ImportConnector(pinecone.Init(baseConn))
		compStore.ImportConnector(redis.Init(baseConn))
		compStore.ImportConnector(restapi.Init(baseConn))
		compStore.ImportConnector(website.Init(baseConn))
		compStore.ImportConnector(slack.Init(baseConn))

	})
	return compStore
}

// ImportOperator loads the operator definitions into memory.
func (s *Store) ImportOperator(op base.IOperator) {
	o := &operator{op: op}
	s.operatorUIDMap[op.GetUID()] = o
	s.operatorIDMap[op.GetID()] = o
	s.operatorUIDs = append(s.operatorUIDs, op.GetUID())
}

// ImportConnector loads the connector definitions into memory.
func (s *Store) ImportConnector(con base.IConnector) {
	c := &connector{con: con}
	s.connectorUIDMap[con.GetUID()] = c
	s.connectorIDMap[con.GetID()] = c
	s.connectorUIDs = append(s.connectorUIDs, con.GetUID())
}

// CreateExecution initializes the execution of a connector given its UID.
func (s *Store) CreateExecution(defUID uuid.UUID, sysVars map[string]any, connection *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	if op, ok := s.operatorUIDMap[defUID]; ok {
		return op.op.CreateExecution(sysVars, task)
	}
	if con, ok := s.connectorUIDMap[defUID]; ok {
		return con.con.CreateExecution(sysVars, connection, task)
	}
	return nil, fmt.Errorf("component definition not found")
}

// GetOperatorDefinitionByUID returns a operator definition by its UID.
func (s *Store) GetOperatorDefinitionByUID(defUID uuid.UUID, sysVars map[string]any, component *base.OperatorComponent) (*pipelinePB.OperatorDefinition, error) {
	if op, ok := s.operatorUIDMap[defUID]; ok {
		return op.op.GetOperatorDefinition(sysVars, component)
	}
	return nil, fmt.Errorf("operator definition not found")
}

// GetOperatorDefinitionByID returns a operator definition by its ID.
func (s *Store) GetOperatorDefinitionByID(defID string, sysVars map[string]any, component *base.OperatorComponent) (*pipelinePB.OperatorDefinition, error) {
	if op, ok := s.operatorIDMap[defID]; ok {
		return op.op.GetOperatorDefinition(sysVars, component)
	}
	return nil, fmt.Errorf("operator definition not found")
}

// ListOperatorDefinitions returns all the loaded operator definitions.
func (s *Store) ListOperatorDefinitions(sysVars map[string]any, returnTombstone bool) []*pipelinePB.OperatorDefinition {
	defs := []*pipelinePB.OperatorDefinition{}
	for _, uid := range s.operatorUIDs {
		op := s.operatorUIDMap[uid]
		def, err := op.op.GetOperatorDefinition(sysVars, nil)
		if err == nil {
			if !def.Tombstone || returnTombstone {
				defs = append(defs, def)
			}
		}
	}
	return defs
}

// GetConnectorDefinitionByUID returns a connector definition by its UID.
func (s *Store) GetConnectorDefinitionByUID(defUID uuid.UUID, sysVars map[string]any, component *base.ConnectorComponent) (*pipelinePB.ConnectorDefinition, error) {
	if con, ok := s.connectorUIDMap[defUID]; ok {
		return con.con.GetConnectorDefinition(sysVars, component)
	}
	return nil, fmt.Errorf("connector definition not found")
}

// GetConnectorDefinitionByID returns a connector definition by its ID.
func (s *Store) GetConnectorDefinitionByID(defID string, sysVars map[string]any, component *base.ConnectorComponent) (*pipelinePB.ConnectorDefinition, error) {
	if con, ok := s.connectorIDMap[defID]; ok {
		return con.con.GetConnectorDefinition(sysVars, component)
	}
	return nil, fmt.Errorf("connector definition not found")
}

// ListConnectorDefinitions returns all the loaded connector definitions.
func (s *Store) ListConnectorDefinitions(sysVars map[string]any, returnTombstone bool) []*pipelinePB.ConnectorDefinition {
	defs := []*pipelinePB.ConnectorDefinition{}
	for _, uid := range s.connectorUIDs {
		con := s.connectorUIDMap[uid]
		def, err := con.con.GetConnectorDefinition(sysVars, nil)
		if err == nil {
			if !def.Tombstone || returnTombstone {
				defs = append(defs, def)
			}
		}
	}
	return defs
}

func (s *Store) IsSecretField(defUID uuid.UUID, target string) (bool, error) {
	if con, ok := s.connectorUIDMap[defUID]; ok {
		return con.con.IsSecretField(target), nil
	}
	return false, fmt.Errorf("connector definition not found")
}
