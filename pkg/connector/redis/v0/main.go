//go:generate compogen readme --connector ./config ./README.mdx
package redis

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
)

const (
	taskWriteChatMessage           = "TASK_WRITE_CHAT_MESSAGE"
	taskWriteMultiModalChatMessage = "TASK_WRITE_MULTI_MODAL_CHAT_MESSAGE"
	taskRetrieveChatHistory        = "TASK_RETRIEVE_CHAT_HISTORY"
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
	base.Connector
}

type execution struct {
	base.ConnectorExecution
}

func Init(bc base.Connector) *connector {
	once.Do(func() {
		con = &connector{Connector: bc}
		err := con.LoadConnectorDefinition(definitionJSON, tasksJSON, nil)
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

func (e *execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := []*structpb.Struct{}

	client, err := NewClient(e.Connection)
	if err != nil {
		return outputs, err
	}
	defer client.Close()

	for _, input := range inputs {
		var output *structpb.Struct
		switch e.Task {
		case taskWriteChatMessage:
			inputStruct := ChatMessageWriteInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			outputStruct := WriteMessage(client, inputStruct)
			output, err = base.ConvertToStructpb(outputStruct)
			if err != nil {
				return nil, err
			}
		case taskWriteMultiModalChatMessage:
			inputStruct := ChatMultiModalMessageWriteInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			outputStruct := WriteMultiModelMessage(client, inputStruct)
			output, err = base.ConvertToStructpb(outputStruct)
			if err != nil {
				return nil, err
			}
		case taskRetrieveChatHistory:
			inputStruct := ChatHistoryRetrieveInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			outputStruct := RetrieveSessionMessages(client, inputStruct)
			output, err = base.ConvertToStructpb(outputStruct)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unsupported task: %s", e.Task)
		}
		outputs = append(outputs, output)
	}
	return outputs, nil
}

func (c *connector) Test(sysVars map[string]any, connection *structpb.Struct) error {
	client, err := NewClient(connection)
	if err != nil {
		return err
	}
	defer client.Close()

	// Ping the Redis server to check the connection
	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
	return nil
}
