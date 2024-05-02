//go:generate compogen readme --connector ./config ./README.mdx
package slack

import (
	_ "embed"
	"fmt"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/component/pkg/connector/util/httpclient"
	"github.com/instill-ai/x/errmsg"
)

const (
	taskWriteMessage = "TASK_WRITE_MESSAGE"
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
	client  *httpclient.Client
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
		client:                 newClient(connection, c.Logger),
	}

	switch task {
	case taskWriteMessage:
		e.execute = e.sendMessage
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

func (e *execution) sendMessage(in *structpb.Struct) (*structpb.Struct, error) {
	params := SendingSlackMessage{}

	if err := base.ConvertFromStructpb(in, &params); err != nil {
		return nil, err
	}

	slackChannels, err := fetchChannelInfo(e.client)
	if err != nil {
		return nil, err
	}

	var slackChannelID string
	for _, slackChannel := range *slackChannels {
		if slackChannel.Name == params.ChannelName {
			slackChannelID = slackChannel.ID
			break
		}
	}

	if slackChannelID == "" {
		err := fmt.Errorf("there is no match name in slack channel [%v]", params.ChannelName)
		return nil, err
	}

	sendingData := SendingData{
		Channel: slackChannelID,
		Text:    params.Message,
	}

	err = postMessageToSlackChannel(e.client, sendingData)
	if err != nil {
		return nil, err
	}

	out, err := base.ConvertToStructpb(WriteTaskResp{
		Result: "succeed",
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (c connector) Test(sysVars map[string]any, connection *structpb.Struct) error {

	return nil
}