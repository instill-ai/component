//go:generate compogen readme ./config ./README.mdx
package slack

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/x/errmsg"
	"github.com/slack-go/slack"
)

const (
	taskWriteMessage = "TASK_WRITE_MESSAGE"
	taskReadMessage  = "TASK_READ_MESSAGE"
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

type component struct {
	base.Component
}

type execution struct {
	base.ComponentExecution

	execute func(*structpb.Struct) (*structpb.Struct, error)
	client  SlackClient
}

type SlackClient interface {
	GetConversations(params *slack.GetConversationsParameters) ([]slack.Channel, string, error)
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
	GetConversationHistory(params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error)
	GetConversationReplies(params *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error)
}

// Init returns an implementation of IConnector that interacts with Slack.
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
	e := &execution{
		ComponentExecution: base.ComponentExecution{Component: c, SystemVariables: sysVars, Setup: setup, Task: task},
		client:             newClient(setup),
	}

	switch task {
	case taskWriteMessage:
		e.execute = e.sendMessage
	case taskReadMessage:
		e.execute = e.readMessage
	default:
		return nil, errmsg.AddMessage(
			fmt.Errorf("not supported task: %s", task),
			fmt.Sprintf("%s task is not supported.", task),
		)
	}

	return &base.ExecutionWrapper{Execution: e}, nil
}

// Execute performs calls the Slack API to execute a task.
func (e *execution) Execute(_ context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := make([]*structpb.Struct, len(inputs))

	for i, input := range inputs {
		output, err := e.execute(input)
		if err != nil {
			return nil, err
		}

		outputs[i] = output
	}

	return outputs, nil
}

func (c *component) Test(sysVars map[string]any, setup *structpb.Struct) error {

	return nil
}
