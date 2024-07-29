package whatsapp

import (
	"context"
	"fmt"
	"sync"

	_ "embed"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
)

const (
	taskSendTextBasedTemplateMessage      = "TASK_SEND_TEXT_BASED_TEMPLATE_MESSAGE"
	taskSendMediaBasedTemplateMessage     = "TASK_SEND_MEDIA_BASED_TEMPLATE_MESSAGE"
	taskSendLocationBasedTemplateMessage  = "TASK_SEND_LOCATION_BASED_TEMPLATE_MESSAGE"
	taskSendAuthenticationTemplateMessage = "TASK_SEND_AUTHENTICATION_TEMPLATE_MESSAGE"
	taskSendTemplateMessage               = "TASK_SEND_TEMPLATE_MESSAGE"
	taskSendTextMessage                   = "TASK_SEND_TEXT_MESSAGE"
	taskSendMediaMessage                  = "TASK_SEND_MEDIA_MESSAGE"
	taskSendLocationMessage               = "TASK_SEND_LOCATION_MESSAGE"
	taskSendContactMessage                = "TASK_SEND_CONTACT_MESSAGE"

	basePath = "https://graph.facebook.com"
	version  = "v20.0"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	//go:embed config/setup.json
	setupJSON []byte

	once sync.Once
	comp *component
)

type component struct {
	base.Component
}

type execution struct {
	base.ComponentExecution
	client  WhatsappInterface
	execute func(*structpb.Struct) (*structpb.Struct, error)
}

// Init returns an implementation of IComponent that implements the greeting
// task.
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
		ComponentExecution: base.ComponentExecution{Component: c, SystemVariables: sysVars, Task: task},
		client:             newClient(setup, c.GetLogger()),
	}

	switch task {
	case taskSendTextBasedTemplateMessage:
		e.execute = e.SendTextBasedTemplateMessage
	case taskSendMediaBasedTemplateMessage:
		e.execute = e.SendMediaBasedTemplateMessage
	case taskSendLocationBasedTemplateMessage:
		e.execute = e.SendLocationBasedTemplateMessage
	case taskSendAuthenticationTemplateMessage:
		e.execute = e.SendAuthenticationTemplateMessage
	case taskSendTextMessage:
		e.execute = e.SendTextMessage
	case taskSendMediaMessage:
		e.execute = e.TaskSendMediaMessage
	case taskSendLocationMessage:
		e.execute = e.TaskSendLocationMessage
	case taskSendContactMessage:
		e.execute = e.TaskSendContactMessage
	default:
		return nil, fmt.Errorf("unsupported task")
	}

	return &base.ExecutionWrapper{Execution: e}, nil
}

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
