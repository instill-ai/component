//go:generate compogen readme ./config ./README.mdx
package redis

import (
	"context"
	"fmt"
	"sync"

	_ "embed"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
)

const (
	taskWriteChatMessage           = "TASK_WRITE_CHAT_MESSAGE"
	taskWriteMultiModalChatMessage = "TASK_WRITE_MULTI_MODAL_CHAT_MESSAGE"
	taskRetrieveChatHistory        = "TASK_RETRIEVE_CHAT_HISTORY"
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
	return &execution{
		ComponentExecution: x,
	}, nil
}

func (e *execution) Execute(ctx context.Context, in base.InputReader, out base.OutputWriter) error {
	inputs, err := in.Read(ctx)
	if err != nil {
		return err
	}
	outputs := []*structpb.Struct{}

	client, err := NewClient(e.Setup)
	if err != nil {
		return err
	}
	defer client.Close()

	for _, input := range inputs {
		var output *structpb.Struct
		switch e.Task {
		case taskWriteChatMessage:
			inputStruct := ChatMessageWriteInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return err
			}
			outputStruct := WriteMessage(client, inputStruct)
			output, err = base.ConvertToStructpb(outputStruct)
			if err != nil {
				return err
			}
		case taskWriteMultiModalChatMessage:
			inputStruct := ChatMultiModalMessageWriteInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return err
			}
			outputStruct := WriteMultiModelMessage(client, inputStruct)
			output, err = base.ConvertToStructpb(outputStruct)
			if err != nil {
				return err
			}
		case taskRetrieveChatHistory:
			inputStruct := ChatHistoryRetrieveInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return err
			}
			outputStruct := RetrieveSessionMessages(client, inputStruct)
			output, err = base.ConvertToStructpb(outputStruct)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported task: %s", e.Task)
		}
		outputs = append(outputs, output)
	}
	return out.Write(ctx, outputs)
}

func (c *component) Test(sysVars map[string]any, setup *structpb.Struct) error {
	client, err := NewClient(setup)
	if err != nil {
		return err
	}
	defer client.Close()

	// Ping the Redis server to check the setup
	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
	return nil
}
