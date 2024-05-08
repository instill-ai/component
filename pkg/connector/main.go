package connector

import (
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/component/pkg/connector/airbyte/v0"
	"github.com/instill-ai/component/pkg/connector/archetypeai/v0"
	"github.com/instill-ai/component/pkg/connector/bigquery/v0"
	"github.com/instill-ai/component/pkg/connector/googlecloudstorage/v0"
	"github.com/instill-ai/component/pkg/connector/googlesearch/v0"
	"github.com/instill-ai/component/pkg/connector/huggingface/v0"
	"github.com/instill-ai/component/pkg/connector/instill/v0"
	"github.com/instill-ai/component/pkg/connector/numbers/v0"
	"github.com/instill-ai/component/pkg/connector/openai/v0"
	"github.com/instill-ai/component/pkg/connector/pinecone/v0"
	"github.com/instill-ai/component/pkg/connector/redis/v0"
	"github.com/instill-ai/component/pkg/connector/restapi/v0"
	"github.com/instill-ai/component/pkg/connector/stabilityai/v0"
	"github.com/instill-ai/component/pkg/connector/website/v0"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

var (
	once     sync.Once
	conStore *Store
)

// Store holds in-memory information about the initialized connectors.
type Store struct {
	connectorUIDs   []uuid.UUID
	connectorUIDMap map[uuid.UUID]*connector
	connectorIDMap  map[string]*connector
}

type connector struct {
	con base.IConnector
}

// ConnectionSecrets contains the global connection secrets of each
// implemented connector (referenced by ID). Connectors may use these secrets
// to skip the connector configuration step and have a ready-to-run
// connection.
type ConnectionSecrets map[string]map[string]any

// Init initializes the different connector components and loads their
// information to memory.
func Init(logger *zap.Logger, secrets ConnectionSecrets) *Store {
	baseConn := base.BaseConnector{Logger: logger}

	once.Do(func() {
		conStore = &Store{
			connectorUIDMap: map[uuid.UUID]*connector{},
			connectorIDMap:  map[string]*connector{},
		}

		conStore.Import(stabilityai.Init(baseConn))
		conStore.Import(instill.Init(baseConn))
		conStore.Import(huggingface.Init(baseConn))

		{
			// OpenAI
			conn := openai.Init(baseConn)
			conn = conn.WithGlobalCredentials(secrets[conn.GetID()])
			conStore.Import(conn)
		}

		conStore.Import(archetypeai.Init(baseConn))
		conStore.Import(numbers.Init(baseConn))
		conStore.Import(airbyte.Init(baseConn))
		conStore.Import(bigquery.Init(baseConn))
		conStore.Import(googlecloudstorage.Init(baseConn))
		conStore.Import(googlesearch.Init(baseConn))
		conStore.Import(pinecone.Init(baseConn))
		conStore.Import(redis.Init(baseConn))
		conStore.Import(restapi.Init(baseConn))
		conStore.Import(website.Init(baseConn))

	})

	return conStore
}

// Import loads the connector definitions into memory.
func (cs *Store) Import(con base.IConnector) {
	c := &connector{con: con}
	cs.connectorUIDMap[con.GetUID()] = c
	cs.connectorIDMap[con.GetID()] = c
	cs.connectorUIDs = append(cs.connectorUIDs, con.GetUID())
}

// CreateExecution initializes the execution of a connector given its UID.
func (cs *Store) CreateExecution(defUID uuid.UUID, sysVars map[string]any, connection *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	if con, ok := cs.connectorUIDMap[defUID]; ok {
		return con.con.CreateExecution(sysVars, connection, task)
	}
	return nil, fmt.Errorf("connector definition not found")
}

// GetConnectorDefinitionByUID returns a connector definition by its UID.
func (cs *Store) GetConnectorDefinitionByUID(defUID uuid.UUID, sysVars map[string]any, component *pipelinePB.ConnectorComponent) (*pipelinePB.ConnectorDefinition, error) {
	if con, ok := cs.connectorUIDMap[defUID]; ok {
		return con.con.GetConnectorDefinition(sysVars, component)
	}
	return nil, fmt.Errorf("connector definition not found")
}

// GetConnectorDefinitionByID returns a connector definition by its ID.
func (cs *Store) GetConnectorDefinitionByID(defID string, sysVars map[string]any, component *pipelinePB.ConnectorComponent) (*pipelinePB.ConnectorDefinition, error) {
	if con, ok := cs.connectorIDMap[defID]; ok {
		return con.con.GetConnectorDefinition(sysVars, component)
	}
	return nil, fmt.Errorf("connector definition not found")
}

// ListConnectorDefinitions returns all the loaded connector definitions.
func (cs *Store) ListConnectorDefinitions(sysVars map[string]any, returnTombstone bool) []*pipelinePB.ConnectorDefinition {
	defs := []*pipelinePB.ConnectorDefinition{}
	for _, uid := range cs.connectorUIDs {
		con := cs.connectorUIDMap[uid]
		def, err := con.con.GetConnectorDefinition(sysVars, nil)
		if err == nil {
			if !def.Tombstone || returnTombstone {
				defs = append(defs, def)
			}
		}
	}
	return defs
}

// IsCredentialField returns whether a given field in a connector contains
// credentials.
func (cs *Store) IsCredentialField(defUID uuid.UUID, target string) (bool, error) {
	if con, ok := cs.connectorUIDMap[defUID]; ok {
		return con.con.IsCredentialField(target), nil
	}
	return false, fmt.Errorf("connector definition not found")
}
