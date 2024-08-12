//go:generate compogen readme ./config ./README.mdx
package ai21labs

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	cfgAPIKey = "api-key"
	baseURL   = "https://api.ai21.com/"
)

const ()

type ExecuteFunction func(*structpb.Struct) (*structpb.Struct, error)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/setup.json
	setupJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte

	once sync.Once
	comp *component
)

type component struct {
	base.Component
	instillAPIKey string
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

type execution struct {
	base.ComponentExecution
	execute                func(*structpb.Struct) (*structpb.Struct, error)
	client                 AI21labsClientInterface
	usesInstillCredentials bool
}

// WithInstillCredentials loads Instill credentials into the component, which
// can be used to configure it with globally defined parameters instead of with
// user-defined credential values.
func (c *component) WithInstillCredentials(s map[string]any) *component {
	c.instillAPIKey = base.ReadFromGlobalConfig(cfgAPIKey, s)
	return c
}

type AI21labsSetup struct {
	APIKey string `json:"api-key"`
}

func (c *component) CreateExecution(x base.ComponentExecution) (base.IExecution, error) {
	resolvedSetup, resolved, err := c.resolveSetup(x.Setup)
	if err != nil {
		return nil, err
	}

	x.Setup = resolvedSetup

	e := &execution{
		ComponentExecution:     x,
		client:                 newClient(getAPIKey(resolvedSetup), getBasePath(x.Setup), c.Logger),
		usesInstillCredentials: resolved,
	}

	taskMap := map[string]func(*structpb.Struct) (*structpb.Struct, error){
		TaskTextGenerationChat:         e.TaskTextGenerationChat,
		TaskTextEmbeddings:             e.TaskTextEmbeddings,
		TaskContextualAnswering:        e.TaskContextualAnswering,
		TaskTextSummarization:          e.TaskTextSummarization,
		TaskTextSummarizationBySegment: e.TaskTextSummarizationBySegment,
		TaskTextParaphrasing:           e.TaskTextParaphrasing,
		TaskGrammarCheck:               e.TaskGrammarCheck,
		TaskTextImprovement:            e.TaskTextImprovement,
		TaskTextSegmentation:           e.TaskTextSegmentation,
	}

	if taskFunc, ok := taskMap[x.Task]; ok {
		e.execute = taskFunc
	} else {
		return nil, fmt.Errorf("unsupported task")
	}
	return e, nil
}

// resolveSetup checks whether the component is configured to use the Instill
// credentials injected during initialization and, if so, returns a new setup
// with the secret credential values.
func (c *component) resolveSetup(setup *structpb.Struct) (*structpb.Struct, bool, error) {
	apiKey := setup.GetFields()[cfgAPIKey].GetStringValue()
	if apiKey != base.SecretKeyword {
		return setup, false, nil
	}

	if c.instillAPIKey == "" {
		return nil, false, base.NewUnresolvedCredential(cfgAPIKey)
	}

	setup.GetFields()[cfgAPIKey] = structpb.NewStringValue(c.instillAPIKey)
	return setup, true, nil
}

func (e *execution) UsesInstillCredentials() bool {
	return e.usesInstillCredentials
}

func (e *execution) Execute(_ context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := make([]*structpb.Struct, len(inputs))

	// The execution takes a array of inputs and returns an array of outputs. The execution is done sequentially.
	for i, input := range inputs {
		output, err := e.execute(input)
		if err != nil {
			return nil, err
		}

		outputs[i] = output
	}

	return outputs, nil
}

// This function is not used in the codebase. It is only used in the tests.
func getBasePath(setup *structpb.Struct) string {
	v, ok := setup.GetFields()["base-path"]
	if !ok {
		return baseURL
	}
	return v.GetStringValue()
}
