package pkg

import (
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

var once sync.Once
var connector base.IConnector

type Connector struct {
	base.Connector
	connectorUIDMap map[uuid.UUID]base.IConnector
	connectorIDMap  map[string]base.IConnector
}

type ConnectorOptions struct {
	Airbyte airbyte.ConnectorOptions
}

func Init(logger *zap.Logger, options ConnectorOptions) base.IConnector {
	once.Do(func() {

		connector = &Connector{
			Connector:       base.Connector{Component: base.Component{Logger: logger}},
			connectorUIDMap: map[uuid.UUID]base.IConnector{},
			connectorIDMap:  map[string]base.IConnector{},
		}

		connector.(*Connector).ImportDefinitions(stabilityai.Init(logger))
		connector.(*Connector).ImportDefinitions(instill.Init(logger))
		connector.(*Connector).ImportDefinitions(huggingface.Init(logger))
		connector.(*Connector).ImportDefinitions(openai.Init(logger))
		connector.(*Connector).ImportDefinitions(archetypeai.Init(logger))
		connector.(*Connector).ImportDefinitions(numbers.Init(logger))
		connector.(*Connector).ImportDefinitions(airbyte.Init(logger, options.Airbyte))
		connector.(*Connector).ImportDefinitions(bigquery.Init(logger))
		connector.(*Connector).ImportDefinitions(googlecloudstorage.Init(logger))
		connector.(*Connector).ImportDefinitions(googlesearch.Init(logger))
		connector.(*Connector).ImportDefinitions(pinecone.Init(logger))
		connector.(*Connector).ImportDefinitions(redis.Init(logger))
		connector.(*Connector).ImportDefinitions(restapi.Init(logger))
		connector.(*Connector).ImportDefinitions(website.Init(logger))

	})
	return connector
}
func (c *Connector) ImportDefinitions(con base.IConnector) {
	for _, v := range con.ListConnectorDefinitions() {
		err := c.AddConnectorDefinition(v)
		if err != nil {
			panic(err)
		}
		c.connectorUIDMap[uuid.FromStringOrNil(v.Uid)] = con
		c.connectorIDMap[v.Id] = con
	}
}

func (c *Connector) CreateExecution(defUID uuid.UUID, task string, config *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	return c.connectorUIDMap[defUID].CreateExecution(defUID, task, config, logger)
}

func (c *Connector) Test(defUID uuid.UUID, config *structpb.Struct, logger *zap.Logger) (pipelinePB.Connector_State, error) {
	return c.connectorUIDMap[defUID].Test(defUID, config, logger)
}

func (c *Connector) GetConnectorDefinitionByID(defID string, resourceConfig *structpb.Struct, component *pipelinePB.ConnectorComponent) (*pipelinePB.ConnectorDefinition, error) {
	return c.connectorIDMap[defID].GetConnectorDefinitionByID(defID, resourceConfig, component)
}

func (c *Connector) GetConnectorDefinitionByUID(defUID uuid.UUID, resourceConfig *structpb.Struct, component *pipelinePB.ConnectorComponent) (*pipelinePB.ConnectorDefinition, error) {
	return c.connectorUIDMap[defUID].GetConnectorDefinitionByUID(defUID, resourceConfig, component)
}
