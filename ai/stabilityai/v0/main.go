//go:generate compogen readme ./config ./README.mdx
package stabilityai

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/x/errmsg"
)

const (
	host = "https://api.stability.ai"

	TextToImageTask  = "TASK_TEXT_TO_IMAGE"
	ImageToImageTask = "TASK_IMAGE_TO_IMAGE"

	cfgAPIKey = "api-key"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/setup.json
	setupJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	//go:embed config/stabilityai.json
	stabilityaiJSON []byte
	once            sync.Once
	comp            *component
)

// Connector executes queries against StabilityAI.
type component struct {
	base.Component

	secretAPIKey string
}

// Init returns an initialized StabilityAI connector.
func Init(bc base.Component) *component {
	once.Do(func() {
		comp = &component{Component: bc}
		err := comp.LoadDefinition(definitionJSON, setupJSON, tasksJSON, map[string][]byte{"stabilityai.json": stabilityaiJSON})
		if err != nil {
			panic(err)
		}
	})

	return comp
}

// WithSecrets loads secrets into the connector, which can be used to configure
// it with globaly defined parameters.
func (c *component) WithSecrets(s map[string]any) *component {
	c.secretAPIKey = base.ReadFromSecrets(cfgAPIKey, s)

	return c
}

// resolveSecrets looks for references to a global secret in the setup
// and replaces them by the global secret injected during initialization.
func (c *component) resolveSecrets(conn *structpb.Struct) (*structpb.Struct, bool, error) {
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
func (c *component) CreateExecution(sysVars map[string]any, setup *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	resolvedSetup, resolved, err := c.resolveSecrets(setup)
	if err != nil {
		return nil, err
	}

	return &base.ExecutionWrapper{Execution: &execution{
		ComponentExecution: base.ComponentExecution{
			Component:       c,
			SystemVariables: sysVars,
			Setup:           resolvedSetup,
			Task:            task,
		},
		usesSecret: resolved,
	}}, nil
}

type execution struct {
	base.ComponentExecution
	usesSecret bool
}

func (e *execution) UsesSecret() bool {
	return e.usesSecret
}

func (e *execution) Execute(_ context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	client := newClient(e.Setup, e.GetLogger())
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
func (c *component) Test(sysVars map[string]any, setup *structpb.Struct) error {
	var engines []Engine
	req := newClient(setup, c.GetLogger()).R().SetResult(&engines)

	if _, err := req.Get(listEnginesPath); err != nil {
		return err
	}

	if len(engines) == 0 {
		return fmt.Errorf("no engines")
	}

	return nil
}
