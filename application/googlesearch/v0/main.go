//go:generate compogen readme ./config ./README.mdx
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

//go:embed config/setup.json
var setupJSON []byte

//go:embed config/tasks.json
var tasksJSON []byte

var once sync.Once
var comp *component

type component struct {
	base.Component
}

type execution struct {
	base.ComponentExecution
}

func Init(bc base.Component) *component {
	once.Do(func() {
		comp = &component{Component: bc}
		err := comp.LoadDefinition(definitionJSON, setupJSON, tasksJSON, nil)
		if err != nil {
			panic(err)
		}
	})
	return comp
}

func (c *component) CreateExecution(sysVars map[string]any, setup *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	return &base.ExecutionWrapper{Execution: &execution{
		ComponentExecution: base.ComponentExecution{Component: c, SystemVariables: sysVars, Setup: setup, Task: task},
	}}, nil
}

// NewService creates a Google custom search service
func NewService(apiKey string) (*customsearch.Service, error) {
	return customsearch.NewService(context.Background(), option.WithAPIKey(apiKey))
}

func getAPIKey(setup *structpb.Struct) string {
	return setup.GetFields()["api-key"].GetStringValue()
}

func getSearchEngineID(setup *structpb.Struct) string {
	return setup.GetFields()["cse-id"].GetStringValue()
}

func (e *execution) Execute(ctx context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {

	service, err := NewService(getAPIKey(e.Setup))
	if err != nil || service == nil {
		return nil, fmt.Errorf("error creating Google custom search service: %v", err)
	}
	cseListCall := service.Cse.List().Cx(getSearchEngineID(e.Setup))

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

func (c *component) Test(sysVars map[string]any, setup *structpb.Struct) error {

	service, err := NewService(getAPIKey(setup))
	if err != nil || service == nil {
		return fmt.Errorf("error creating Google custom search service: %v", err)
	}
	if service == nil {
		return fmt.Errorf("error creating Google custom search service: %v", err)
	}
	return nil
}
