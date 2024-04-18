//go:generate compogen readme --connector ./config ./README.mdx
package pinecone

import (
	_ "embed"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/component/pkg/connector/util/httpclient"
)

const (
	taskQuery  = "TASK_QUERY"
	taskUpsert = "TASK_UPSERT"

	upsertPath = "/vectors/upsert"
	queryPath  = "/query"
)

//go:embed config/definition.json
var definitionJSON []byte

//go:embed config/tasks.json
var tasksJSON []byte

var once sync.Once
var connector base.IConnector

type Connector struct {
	base.Connector
}

type Execution struct {
	base.Execution
}

func Init(logger *zap.Logger, usageHandler base.UsageHandler) base.IConnector {
	once.Do(func() {
		connector = &Connector{
			Connector: base.Connector{
				Component: base.Component{Logger: logger, UsageHandler: usageHandler},
			},
		}
		err := connector.LoadConnectorDefinition(definitionJSON, tasksJSON, nil)
		if err != nil {
			logger.Fatal(err.Error())
		}
	})
	return connector
}

func (c *Connector) CreateExecution(defUID uuid.UUID, task string, connection *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	e := &Execution{}
	e.Execution = base.CreateExecutionHelper(e, c, defUID, task, connection, logger)
	return e, nil
}

func newClient(config *structpb.Struct, logger *zap.Logger) *httpclient.Client {
	c := httpclient.New("Pinecone", getURL(config),
		httpclient.WithLogger(logger),
		httpclient.WithEndUserError(new(errBody)),
	)

	c.SetHeader("Api-Key", getAPIKey(config))

	return c
}

func getAPIKey(config *structpb.Struct) string {
	return config.GetFields()["api_key"].GetStringValue()
}

func getURL(config *structpb.Struct) string {
	return config.GetFields()["url"].GetStringValue()
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	req := newClient(e.Config, e.Logger).R()
	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		var output *structpb.Struct
		switch e.Task {
		case taskQuery:
			inputStruct := queryInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}

			// Each query request can contain only one of the parameters
			// vector, or id.
			// Ref: https://docs.pinecone.io/reference/query
			if inputStruct.ID != "" {
				inputStruct.Vector = nil
			}

			resp := queryResp{}
			req.SetResult(&resp).SetBody(inputStruct.asRequest())

			if _, err := req.Post(queryPath); err != nil {
				return nil, httpclient.WrapURLError(err)
			}

			resp = resp.filterOutBelowThreshold(inputStruct.MinScore)

			output, err = base.ConvertToStructpb(resp)
			if err != nil {
				return nil, err
			}
		case taskUpsert:
			v := upsertInput{}
			err := base.ConvertFromStructpb(input, &v)
			if err != nil {
				return nil, err
			}

			resp := upsertResp{}
			req.SetResult(&resp).SetBody(upsertReq{
				Vectors:   []vector{v.vector},
				Namespace: v.Namespace,
			})

			if _, err := req.Post(upsertPath); err != nil {
				return nil, httpclient.WrapURLError(err)
			}

			output, err = base.ConvertToStructpb(upsertOutput(resp))
			if err != nil {
				return nil, err
			}
		}
		outputs = append(outputs, output)
	}
	return outputs, nil
}

func (c *Connector) Test(defUID uuid.UUID, config *structpb.Struct, logger *zap.Logger) error {
	//TODO: change this
	return nil
}
