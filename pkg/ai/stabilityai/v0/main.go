//go:generate compogen readme --connector ./config ./README.mdx
package stabilityai

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/x/errmsg"
)

const (
	host = "https://api.stability.ai"

	TextToImageTask  = "TASK_TEXT_TO_IMAGE"
	ImageToImageTask = "TASK_IMAGE_TO_IMAGE"

	cfgAPIKey = "api_key"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	//go:embed config/stabilityai.json
	stabilityaiJSON []byte
	once            sync.Once
	con             *Connector
)

// Connector executes queries against StabilityAI.
type Connector struct {
	base.Connector

	usageHandlerCreator base.UsageHandlerCreator
	secretAPIKey        string
}

// Init returns an initialized StabilityAI connector.
func Init(bc base.Connector) *Connector {
	once.Do(func() {
		con = &Connector{Connector: bc}
		err := con.LoadConnectorDefinition(definitionJSON, tasksJSON, map[string][]byte{"stabilityai.json": stabilityaiJSON})
		if err != nil {
			panic(err)
		}
	})

	return con
}

// WithSecrets loads secrets into the connector, which can be used to configure
// it with globaly defined parameters.
func (c *Connector) WithSecrets(s map[string]any) *Connector {
	c.secretAPIKey = base.ReadFromSecrets(cfgAPIKey, s)

	return c
}

// WithUsageHandlerCreator overrides the UsageHandlerCreator method.
func (c *Connector) WithUsageHandlerCreator(newUH base.UsageHandlerCreator) *Connector {
	c.usageHandlerCreator = newUH
	return c
}

// UsageHandlerCreator returns a function to initialize a UsageHandler.
func (c *Connector) UsageHandlerCreator() base.UsageHandlerCreator {
	if c.usageHandlerCreator == nil {
		return c.Connector.UsageHandlerCreator()
	}
	return c.usageHandlerCreator
}

// resolveSecrets looks for references to a global secret in the connection
// and replaces them by the global secret injected during initialization.
func (c *Connector) resolveSecrets(conn *structpb.Struct) (*structpb.Struct, bool, error) {
	apiKey := conn.GetFields()[cfgAPIKey].GetStringValue()
	if apiKey != base.SecretKeyword {
		return conn, false, nil
	}

	if c.secretAPIKey == "" {
		return nil, false, base.NewUnresolvedSecret(cfgAPIKey)
	}

	conn.GetFields()[cfgAPIKey] = structpb.NewStringValue(c.secretAPIKey)
	return conn, true, nil
}

// CreateExecution initializes a connector executor that can be used in a
// pipeline trigger.
func (c *Connector) CreateExecution(sysVars map[string]any, connection *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	resolvedConnection, resolved, err := c.resolveSecrets(connection)
	if err != nil {
		return nil, err
	}

	return &base.ExecutionWrapper{Execution: &execution{
		ConnectorExecution: base.ConnectorExecution{
			Connector:       c,
			SystemVariables: sysVars,
			Connection:      resolvedConnection,
			Task:            task,
		},
		usesSecret: resolved,
	}}, nil
}

type execution struct {
	base.ConnectorExecution
	usesSecret bool
}

func (e *execution) UsesSecret() bool {
	return e.usesSecret
}

func (e *execution) Execute(_ context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	client := newClient(e.Connection, e.GetLogger())
	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		switch e.Task {
		case TextToImageTask:
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
		case ImageToImageTask:
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
func (c *Connector) Test(sysVars map[string]any, connection *structpb.Struct) error {
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
