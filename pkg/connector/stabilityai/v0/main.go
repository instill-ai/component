//go:generate compogen readme --connector ./config ./README.mdx
package stabilityai

import (
	_ "embed"
	"fmt"
	"sync"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/x/errmsg"
)

const (
	host             = "https://api.stability.ai"
	textToImageTask  = "TASK_TEXT_TO_IMAGE"
	imageToImageTask = "TASK_IMAGE_TO_IMAGE"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	//go:embed config/stabilityai.json
	stabilityaiJSON []byte
	once            sync.Once
	con             *connector
)

type connector struct {
	base.Connector
}

type execution struct {
	base.ConnectorExecution
}

func Init(bc base.Connector) *connector {
	once.Do(func() {
		con = &connector{Connector: bc}
		err := con.LoadConnectorDefinition(definitionJSON, tasksJSON, map[string][]byte{"stabilityai.json": stabilityaiJSON})
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

func getAPIKey(config *structpb.Struct) string {
	return config.GetFields()["api_key"].GetStringValue()
}

// getBasePath returns Stability AI's API URL. This configuration param allows
// us to override the API the connector will point to. It isn't meant to be
// exposed to users. Rather, it can serve to test the logic against a fake
// server.
// TODO instead of having the API value hardcoded in the codebase, it should be
// read from a config file or environment variable.
func getBasePath(config *structpb.Struct) string {
	v, ok := config.GetFields()["base_path"]
	if !ok {
		return host
	}
	return v.GetStringValue()
}

func (e *execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	client := newClient(e.Connection, e.GetLogger())
	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		switch e.Task {
		case textToImageTask:
			params, err := parseTextToImageReq(input)
			if err != nil {
				return inputs, err
			}

			resp := ImageTaskRes{}
			req := client.R().SetResult(&resp).SetBody(params)

			if _, err := req.Post(params.path); err != nil {
				return inputs, err
			}

			output, err := textToImageOutput(resp)
			if err != nil {
				return nil, err
			}

			outputs = append(outputs, output)
		case imageToImageTask:
			params, err := parseImageToImageReq(input)
			if err != nil {
				return inputs, err
			}

			data, ct, err := params.getBytes()
			if err != nil {
				return inputs, err
			}

			resp := ImageTaskRes{}
			req := client.R().SetBody(data).SetResult(&resp).SetHeader("Content-Type", ct)

			if _, err := req.Post(params.path); err != nil {
				return inputs, err
			}

			output, err := imageToImageOutput(resp)
			if err != nil {
				return nil, err
			}

			outputs = append(outputs, output)

		default:
			return nil, errmsg.AddMessage(
				fmt.Errorf("not supported task: %s", e.Task),
				fmt.Sprintf("%s task is not supported.", e.Task),
			)
		}
	}
	return outputs, nil
}

// Test checks the connector state.
func (c *connector) Test(sysVars map[string]any, connection *structpb.Struct) error {
	var engines []Engine
	req := newClient(connection, c.Logger).R().SetResult(&engines)

	if _, err := req.Get(listEnginesPath); err != nil {
		return err
	}

	if len(engines) == 0 {
		return fmt.Errorf("no engines")
	}

	return nil
}
