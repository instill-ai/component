package pkg

import (
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
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

type ConnectorOptions struct{}

func Init(logger *zap.Logger, usageHandler base.UsageHandler, options ConnectorOptions) base.IConnector {
	once.Do(func() {

		connector = &Connector{
			Connector:       base.Connector{Component: base.Component{Logger: logger, UsageHandler: usageHandler}},
			connectorUIDMap: map[uuid.UUID]base.IConnector{},
			connectorIDMap:  map[string]base.IConnector{},
		}

		connector.(*Connector).ImportDefinitions(stabilityai.Init(logger, usageHandler))
		connector.(*Connector).ImportDefinitions(instill.Init(logger, usageHandler))
		connector.(*Connector).ImportDefinitions(huggingface.Init(logger, usageHandler))
		connector.(*Connector).ImportDefinitions(openai.Init(logger, usageHandler))
		connector.(*Connector).ImportDefinitions(archetypeai.Init(logger, usageHandler))
		connector.(*Connector).ImportDefinitions(numbers.Init(logger, usageHandler))
		connector.(*Connector).ImportDefinitions(bigquery.Init(logger, usageHandler))
		connector.(*Connector).ImportDefinitions(googlecloudstorage.Init(logger, usageHandler))
		connector.(*Connector).ImportDefinitions(googlesearch.Init(logger, usageHandler))
		connector.(*Connector).ImportDefinitions(pinecone.Init(logger, usageHandler))
		connector.(*Connector).ImportDefinitions(redis.Init(logger, usageHandler))
		connector.(*Connector).ImportDefinitions(restapi.Init(logger, usageHandler))
		connector.(*Connector).ImportDefinitions(website.Init(logger, usageHandler))

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

func (c *Connector) CreateExecution(defUID uuid.UUID, task string, connection *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	return c.connectorUIDMap[defUID].CreateExecution(defUID, task, connection, logger)
}

func (c *Connector) Test(defUID uuid.UUID, config *structpb.Struct, logger *zap.Logger) error {
	return c.connectorUIDMap[defUID].Test(defUID, config, logger)
}

func (c *Connector) GetConnectorDefinitionByID(defID string, component *pipelinePB.ConnectorComponent) (*pipelinePB.ConnectorDefinition, error) {
	return c.connectorIDMap[defID].GetConnectorDefinitionByID(defID, component)
}

func (c *Connector) GetConnectorDefinitionByUID(defUID uuid.UUID, component *pipelinePB.ConnectorComponent) (*pipelinePB.ConnectorDefinition, error) {
	return c.connectorUIDMap[defUID].GetConnectorDefinitionByUID(defUID, component)
}
