package googlesearch

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

const (
	taskSearch = "TASK_SEARCH"
)

//go:embed config/definitions.json
var definitionsJSON []byte

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

func Init(logger *zap.Logger) base.IConnector {
	once.Do(func() {
		connector = &Connector{
			Connector: base.Connector{
				Component: base.Component{Logger: logger},
			},
		}
		err := connector.LoadConnectorDefinitions(definitionsJSON, tasksJSON, nil)
		if err != nil {
			logger.Fatal(err.Error())
		}
	})
	return connector
}

func (c *Connector) CreateExecution(defUID uuid.UUID, task string, config *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	e := &Execution{}
	e.Execution = base.CreateExecutionHelper(e, c, defUID, task, config, logger)
	return e, nil
}

// NewService creates a Google custom search service
func NewService(apiKey string) (*customsearch.Service, error) {
	return customsearch.NewService(context.Background(), option.WithAPIKey(apiKey))
}

func getAPIKey(config *structpb.Struct) string {
	return config.GetFields()["api_key"].GetStringValue()
}

func getSearchEngineID(config *structpb.Struct) string {
	return config.GetFields()["cse_id"].GetStringValue()
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {

	service, err := NewService(getAPIKey(e.Config))
	if err != nil || service == nil {
		return nil, fmt.Errorf("error creating Google custom search service: %v", err)
	}
	cseListCall := service.Cse.List().Cx(getSearchEngineID(e.Config))

	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		switch e.Task {
		case taskSearch:

			inputStruct := SearchInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}

			// Make the search request
			outputStruct, err := search(cseListCall, inputStruct)

			if err != nil {
				return nil, err
			}

			outputJSON, err := json.Marshal(outputStruct)
			if err != nil {
				return nil, err
			}
			output := structpb.Struct{}
			err = json.Unmarshal(outputJSON, &output)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, &output)

		default:
			return nil, fmt.Errorf("not supported task: %s", e.Task)
		}
	}

	return outputs, nil
}

func (c *Connector) Test(defUID uuid.UUID, config *structpb.Struct, logger *zap.Logger) (pipelinePB.Connector_State, error) {

	service, err := NewService(getAPIKey(config))
	if err != nil || service == nil {
		return pipelinePB.Connector_STATE_ERROR, fmt.Errorf("error creating Google custom search service: %v", err)
	}
	if service == nil {
		return pipelinePB.Connector_STATE_ERROR, fmt.Errorf("error creating Google custom search service: %v", err)
	}
	return pipelinePB.Connector_STATE_CONNECTED, nil
}
