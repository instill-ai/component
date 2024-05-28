//go:generate compogen readme --connector ./config ./README.mdx
package googlesearch

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
)

const (
	taskSearch = "TASK_SEARCH"
)

//go:embed config/definition.json
var definitionJSON []byte

//go:embed config/tasks.json
var tasksJSON []byte

var once sync.Once
var con *connector

type connector struct {
	base.Connector
}

type execution struct {
	base.ConnectorExecution
}

func Init(bc base.Connector) *connector {
	once.Do(func() {
		con = &connector{Connector: bc}
		err := con.LoadConnectorDefinition(definitionJSON, tasksJSON, nil)
		if err != nil {
			panic(err)
		}
	})
	return con
}

func (c *connector) CreateExecution(sysVars map[string]any, connection *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	return &base.ExecutionWrapper{Execution: &execution{
		ConnectorExecution: base.ConnectorExecution{Connector: c, SystemVariables: sysVars, Connection: connection, Task: task},
	}}, nil
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

func (e *execution) Execute(ctx context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {

	service, err := NewService(getAPIKey(e.Connection))
	if err != nil || service == nil {
		return nil, fmt.Errorf("error creating Google custom search service: %v", err)
	}
	cseListCall := service.Cse.List().Cx(getSearchEngineID(e.Connection))

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

func (c *connector) Test(sysVars map[string]any, connection *structpb.Struct) error {

	service, err := NewService(getAPIKey(connection))
	if err != nil || service == nil {
		return fmt.Errorf("error creating Google custom search service: %v", err)
	}
	if service == nil {
		return fmt.Errorf("error creating Google custom search service: %v", err)
	}
	return nil
}
