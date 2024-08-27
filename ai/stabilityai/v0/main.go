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

	instillAPIKey string
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

// WithInstillCredentials loads Instill credentials into the component, which
// can be used to configure it with globally defined parameters instead of with
// user-defined credential values.
func (c *component) WithInstillCredentials(s map[string]any) *component {
	c.instillAPIKey = base.ReadFromGlobalConfig(cfgAPIKey, s)
	return c
}

// resolveSetup checks whether the component is configured to use the Instill
// credentials injected during initialization and, if so, returns a new setup
// with the secret credential values.
func (c *component) resolveSetup(setup *structpb.Struct) (*structpb.Struct, bool, error) {
	apiKey := setup.GetFields()[cfgAPIKey].GetStringValue()
	if apiKey != "" && apiKey != base.SecretKeyword {
		return setup, false, nil
	}

	if c.instillAPIKey == "" {
		return nil, false, base.NewUnresolvedCredential(cfgAPIKey)
	}

	setup.GetFields()[cfgAPIKey] = structpb.NewStringValue(c.instillAPIKey)
	return setup, true, nil
}

// CreateExecution initializes a connector executor that can be used in a
// pipeline trigger.
func (c *component) CreateExecution(x base.ComponentExecution) (base.IExecution, error) {
	resolvedSetup, resolved, err := c.resolveSetup(x.Setup)
	if err != nil {
		return nil, err
	}

	x.Setup = resolvedSetup

	return &execution{
		ComponentExecution:     x,
		usesInstillCredentials: resolved,
	}, nil
}

type execution struct {
	base.ComponentExecution
	usesInstillCredentials bool
}

func (e *execution) UsesInstillCredentials() bool {
	return e.usesInstillCredentials
}

func (e *execution) Execute(ctx context.Context, in base.InputReader, out base.OutputWriter) error {
	inputs, err := in.Read(ctx)
	if err != nil {
		return err
	}
	client := newClient(e.Setup, e.GetLogger())
	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		switch e.Task {
		case TextToImageTask:
			params, err := parseTextToImageReq(input)
			if err != nil {
				return err
			}

			resp := ImageTaskRes{}
			req := client.R().SetResult(&resp).SetBody(params)

			if _, err := req.Post(params.path); err != nil {
				return err
			}

			output, err := textToImageOutput(resp)
			if err != nil {
				return err
			}

			outputs = append(outputs, output)
		case ImageToImageTask:
			params, err := parseImageToImageReq(input)
			if err != nil {
				return err
			}

			data, ct, err := params.getBytes()
			if err != nil {
				return err
			}

			resp := ImageTaskRes{}
			req := client.R().SetBody(data).SetResult(&resp).SetHeader("Content-Type", ct)

			if _, err := req.Post(params.path); err != nil {
				return err
			}

			output, err := imageToImageOutput(resp)
			if err != nil {
				return err
			}

			outputs = append(outputs, output)

		default:
			return errmsg.AddMessage(
				fmt.Errorf("not supported task: %s", e.Task),
				fmt.Sprintf("%s task is not supported.", e.Task),
			)
		}
	}
	return out.Write(ctx, outputs)
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
