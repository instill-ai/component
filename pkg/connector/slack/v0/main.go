//go:generate compogen readme --connector ./config ./README.mdx
package slack

import (
	_ "embed"
	"fmt"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
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
	//go:embed config/tasks.json
	tasksJSON []byte

	once sync.Once
	con  *connector
)

type connector struct {
	base.BaseConnector
}

type execution struct {
	base.BaseConnectorExecution

	execute func(*structpb.Struct) (*structpb.Struct, error)
	client  *slack.Client
}

// Init returns an implementation of IConnector that interacts with Slack.
func Init(l *zap.Logger, u base.UsageHandler) *connector {
	once.Do(func() {
		con = &connector{
			BaseConnector: base.BaseConnector{
				Logger:       l,
				UsageHandler: u,
			},
		}
		err := con.LoadConnectorDefinition(definitionJSON, tasksJSON, nil)
		if err != nil {
			panic(err)
		}
	})
	return con
}

func (c *connector) CreateExecution(sysVars map[string]any, connection *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	e := &execution{
		BaseConnectorExecution: base.BaseConnectorExecution{Connector: c, SystemVariables: sysVars, Connection: connection, Task: task},
		client:                 newClient(connection),
	}

	switch task {
	case taskWriteMessage:
		e.execute = e.sendMessage
	case taskReadMessage:
		// TODO: Read Task
		// e.execute = e.readMessage
	default:
		return nil, errmsg.AddMessage(
			fmt.Errorf("not supported task: %s", task),
			fmt.Sprintf("%s task is not supported.", task),
		)
	}

	return &base.ExecutionWrapper{Execution: e}, nil
}

// Execute performs calls the Slack API to execute a task.
func (e *execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
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

func (c connector) Test(sysVars map[string]any, connection *structpb.Struct) error {

	return nil
}
