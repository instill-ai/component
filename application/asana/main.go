//go:generate compogen readme ./config ./README.mdx
package asana

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/x/errmsg"
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

const (
	apiBaseURL         = "https://app.asana.com/api/1.0"
	taskAsanaGoal      = "TASK_ASANA_GOAL"
	taskAsanaTask      = "TASK_ASANA_TASK"
	taskAsanaPortfolio = "TASK_ASANA_PORTFOLIO"
	taskAsanaProject   = "TASK_ASANA_PROJECT"
)

type component struct {
	base.Component
}

type execution struct {
	base.ComponentExecution
	execute func(context.Context, *structpb.Struct) (*structpb.Struct, error)
	client  Client
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

func (c *component) CreateExecution(x base.ComponentExecution) (base.IExecution, error) {
	ctx := context.Background()
	asanaClient, err := newClient(ctx, x.Setup, c.Logger)
	if err != nil {
		return nil, err
	}
	e := &execution{
		ComponentExecution: x,
		client:             *asanaClient,
	}
	switch x.Task {
	case taskAsanaGoal:
		e.execute = e.client.GoalRelatedTask
	case taskAsanaTask:
		e.execute = e.client.TaskRelatedTask
	case taskAsanaPortfolio:
		e.execute = e.client.PortfolioRelatedTask
	case taskAsanaProject:
		e.execute = e.client.ProjectRelatedTask
	default:
		return nil, errmsg.AddMessage(
			fmt.Errorf("not supported task: %s", x.Task),
			fmt.Sprintf("%s task is not supported.", x.Task),
		)
	}
	return e, nil
}

func (e *execution) Execute(ctx context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := make([]*structpb.Struct, len(inputs))
	for i, input := range inputs {
		input, err := e.ComponentExecution.FillInDefaultValues(input)
		if err != nil {
			return nil, err
		}
		output, err := e.execute(ctx, input)
		if err != nil {
			return nil, err
		}
		outputs[i] = output
	}
	return outputs, nil
}
