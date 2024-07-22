//go:generate compogen readme ./config ./README.mdx

package hubspot

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	hubspot "github.com/belong-inc/go-hubspot"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	taskGetContact          = "TASK_GET_CONTACT"
	taskCreateContact       = "TASK_CREATE_CONTACT"
	taskGetDeal             = "TASK_GET_DEAL"
	taskCreateDeal          = "TASK_CREATE_DEAL"
	taskGetCompany          = "TASK_GET_COMPANY"
	taskCreateCompany       = "TASK_CREATE_COMPANY"
	taskGetTicket           = "TASK_GET_TICKET"
	taskCreateTicket        = "TASK_CREATE_TICKET"
	taskGetThread           = "TASK_GET_THREAD"
	taskRetrieveAssociation = "TASK_RETRIEVE_ASSOCIATION"
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
	client  *CustomClient
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

func getToken(setup *structpb.Struct) string {
	return setup.GetFields()["token"].GetStringValue()
}

// custom client to support thread task
func hubspotNewCustomClient(setup *structpb.Struct) *CustomClient {
	client, err := NewCustomClient(hubspot.SetPrivateAppToken(getToken(setup)))

	if err != nil {
		panic(err)
	}

	return client
}

func (c *component) CreateExecution(sysVars map[string]any, setup *structpb.Struct, task string) (*base.ExecutionWrapper, error) {

	e := &execution{
		ComponentExecution: base.ComponentExecution{Component: c, SystemVariables: sysVars, Task: task},
		client:             hubspotNewCustomClient(setup),
	}

	switch task {
	case taskGetContact:
		e.execute = e.GetContact
	case taskCreateContact:
		e.execute = e.CreateContact
	case taskGetDeal:
		e.execute = e.GetDeal
	case taskCreateDeal:
		e.execute = e.CreateDeal
	case taskGetCompany:
		e.execute = e.GetCompany
	case taskCreateCompany:
		e.execute = e.CreateCompany
	case taskGetTicket:
		e.execute = e.GetTicket
	case taskCreateTicket:
		e.execute = e.CreateTicket
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
