package redis

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

const (
	taskWriteChatMessage           = "TASK_WRITE_CHAT_MESSAGE"
	taskWriteMultiModalChatMessage = "TASK_WRITE_MULTI_MODAL_CHAT_MESSAGE"
	taskRetrieveChatHistory        = "TASK_RETRIEVE_CHAT_HISTORY"
)

var (
	//go:embed config/definitions.json
	definitionsJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte

	once      sync.Once
	connector base.IConnector
)

type Connector struct {
	base.Connector
}

type Execution struct {
	base.Execution
}

func Init(logger *zap.Logger) base.IConnector {
	once.Do(func() {
		connector = &Connector{
			Connector: base.Connector{
				Component: base.Component{Logger: logger},
			},
		}
		err := connector.LoadConnectorDefinitions(definitionsJSON, tasksJSON, nil)
		if err != nil {
			logger.Fatal(err.Error())
		}
	})
	return connector
}

func (c *Connector) CreateExecution(defUID uuid.UUID, task string, config *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	e := &Execution{}
	e.Execution = base.CreateExecutionHelper(e, c, defUID, task, config, logger)
	return e, nil
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := []*structpb.Struct{}

	client, err := NewClient(e.Config)
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

func (c *Connector) Test(defUID uuid.UUID, config *structpb.Struct, logger *zap.Logger) (pipelinePB.Connector_State, error) {
	client, err := NewClient(config)
	if err != nil {
		return pipelinePB.Connector_STATE_ERROR, err
	}
	defer client.Close()

	// Ping the Redis server to check the connection
	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		return pipelinePB.Connector_STATE_DISCONNECTED, err
	}
	return pipelinePB.Connector_STATE_CONNECTED, nil
}
