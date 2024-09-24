//go:generate compogen readme ./config ./README.mdx
package universalai

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
)

const (
	TextChatTask    = "TASK_CHAT"
	cfgAPIKey       = "api-key"
	cfgOrganization = "organization"
	retryCount      = 3
)

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

// Connector executes queries against OpenAI.
type component struct {
	base.Component

	instillAPIKey map[string]string
}

type execution struct {
	base.ComponentExecution

	component              *component
	usesInstillCredentials bool
	execute                func(*structpb.Struct, *base.Job, context.Context) (*structpb.Struct, error)
}

// Init returns an initialized AI connector.
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

func (c *component) CreateExecution(x base.ComponentExecution) (base.IExecution, error) {
	e := &execution{
		ComponentExecution: x,
		component:          c,
	}

	switch x.Task {
	case TextChatTask:
		e.execute = e.ExecuteTextChat
	default:
		return nil, fmt.Errorf("unknown task: %s", x.Task)
	}

	return e, nil
}

func (e *execution) Execute(ctx context.Context, jobs []*base.Job) error {
	return base.ConcurrentExecutor(ctx, jobs, e.execute)
}

func (e *execution) UsesInstillCredentials() bool {
	return e.usesInstillCredentials
}

// WithInstillCredentials loads Instill credentials into the component, which
// can be used to configure it with globally defined parameters instead of with
// user-defined credential values.
func (c *component) WithInstillCredentials(vendor string, s map[string]any) *component {
	if c.instillAPIKey == nil {
		c.instillAPIKey = make(map[string]string)
	}
	c.instillAPIKey[vendor] = base.ReadFromGlobalConfig(cfgAPIKey, s)
	return c
}

// resolveSetup checks whether the component is configured to use the Instill
// credentials injected during initialization and, if so, returns a new setup
// with the secret credential values.
func (c *component) resolveSetup(vendor string, setup *structpb.Struct) (*structpb.Struct, bool, error) {
	if setup == nil || setup.Fields == nil {
		setup = &structpb.Struct{Fields: map[string]*structpb.Value{}}
	}
	if v, ok := setup.GetFields()[cfgAPIKey]; ok {
		apiKey := v.GetStringValue()
		if apiKey != "" && apiKey != base.SecretKeyword {
			return setup, false, nil
		}
	}

	if c.instillAPIKey[vendor] == "" {
		return nil, false, base.NewUnresolvedCredential(cfgAPIKey)
	}

	setup.GetFields()[cfgAPIKey] = structpb.NewStringValue(c.instillAPIKey[vendor])
	return setup, true, nil
}
