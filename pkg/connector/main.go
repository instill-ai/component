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
	conStore *ConnectorStore
)

type ConnectorStore struct {
	connectorUIDs   []uuid.UUID
	connectorUIDMap map[uuid.UUID]*connector
	connectorIDMap  map[string]*connector
}

type connector struct {
	con base.IConnector
}

func Init(logger *zap.Logger, usageHandler base.UsageHandler) *ConnectorStore {
	once.Do(func() {

		conStore = &ConnectorStore{
			connectorUIDMap: map[uuid.UUID]*connector{},
			connectorIDMap:  map[string]*connector{},
		}

		conStore.Import(stabilityai.Init(logger, usageHandler))
		conStore.Import(instill.Init(logger, usageHandler))
		conStore.Import(huggingface.Init(logger, usageHandler))
		conStore.Import(openai.Init(logger, usageHandler))
		conStore.Import(archetypeai.Init(logger, usageHandler))
		conStore.Import(numbers.Init(logger, usageHandler))
		conStore.Import(airbyte.Init(logger, usageHandler))
		conStore.Import(bigquery.Init(logger, usageHandler))
		conStore.Import(googlecloudstorage.Init(logger, usageHandler))
		conStore.Import(googlesearch.Init(logger, usageHandler))
		conStore.Import(pinecone.Init(logger, usageHandler))
		conStore.Import(redis.Init(logger, usageHandler))
		conStore.Import(restapi.Init(logger, usageHandler))
		conStore.Import(website.Init(logger, usageHandler))

	})
	return conStore
}

// Imports imports the connector definitions
func (cs *ConnectorStore) Import(con base.IConnector) {
	c := &connector{con: con}
	cs.connectorUIDMap[con.GetUID()] = c
	cs.connectorIDMap[con.GetID()] = c
	cs.connectorUIDs = append(cs.connectorUIDs, con.GetUID())
}

func (cs *ConnectorStore) CreateExecution(defUID uuid.UUID, sysVars map[string]any, connection *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	if con, ok := cs.connectorUIDMap[defUID]; ok {
		return con.con.CreateExecution(sysVars, connection, task)
	}
	return nil, fmt.Errorf("connector definition not found")
}

func (cs *ConnectorStore) GetConnectorDefinitionByUID(defUID uuid.UUID, sysVars map[string]any, component *pipelinePB.ConnectorComponent) (*pipelinePB.ConnectorDefinition, error) {
	if con, ok := cs.connectorUIDMap[defUID]; ok {
		return con.con.GetConnectorDefinition(sysVars, component)
	}
	return nil, fmt.Errorf("connector definition not found")
}

// Get the connector definition by definition id
func (cs *ConnectorStore) GetConnectorDefinitionByID(defID string, sysVars map[string]any, component *pipelinePB.ConnectorComponent) (*pipelinePB.ConnectorDefinition, error) {
	if con, ok := cs.connectorIDMap[defID]; ok {
		return con.con.GetConnectorDefinition(sysVars, component)
	}
	return nil, fmt.Errorf("connector definition not found")
}

// Get the list of connector definitions under this connector
func (cs *ConnectorStore) ListConnectorDefinitions(returnTombstone bool) []*pipelinePB.ConnectorDefinition {
	defs := []*pipelinePB.ConnectorDefinition{}
	for _, uid := range cs.connectorUIDs {
		con := cs.connectorUIDMap[uid]
		def, err := con.con.GetConnectorDefinition(nil, nil)
		if err == nil {
			if !def.Tombstone || returnTombstone {
				defs = append(defs, def)
			}
		}
	}
	return defs
}

func (cs *ConnectorStore) IsCredentialField(defUID uuid.UUID, target string) (bool, error) {
	if con, ok := cs.connectorUIDMap[defUID]; ok {
		return con.con.IsCredentialField(target), nil
	}
	return false, fmt.Errorf("connector definition not found")
}
