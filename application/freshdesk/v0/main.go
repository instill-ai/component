//go:generate compogen readme ./config ./README.mdx

package freshdesk

import (
	"context"
	"fmt"
	"strings"
	"sync"

	_ "embed"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
)

const (
	version = "v2"

	taskGetTicket     = "TASK_GET_TICKET"
	taskCreateTicket  = "TASK_CREATE_TICKET"
	taskGetContact    = "TASK_GET_CONTACT"
	taskCreateContact = "TASK_CREATE_CONTACT"
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
	client  FreshdeskInterface
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

func (c *component) CreateExecution(x base.ComponentExecution) (base.IExecution, error) {
	e := &execution{
		ComponentExecution: x,
		client:             newClient(x.Setup, c.GetLogger()),
	}

	switch x.Task {
	case taskGetTicket:
		e.execute = e.TaskGetTicket
	case taskCreateTicket:
		e.execute = e.TaskCreateTicket
	case taskGetContact:
		e.execute = e.TaskGetContact
	default:
		return nil, fmt.Errorf("unsupported task")
	}

	return e, nil
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

func convertTimestampResp(timestamp string) string {
	// freshdesk response timestamp is always in the format of YYYY-MM-DDTHH:MM:SSZ and in UTC.
	// this function will convert it to YYYY-MM-DD HH:MM:SS UTC

	if timestamp == "" {
		return timestamp
	}
	formattedTime := strings.Replace(timestamp, "T", " ", 1)
	formattedTime = strings.Replace(formattedTime, "Z", " ", 1)
	formattedTime += "UTC"

	return formattedTime
}

// func checkForNilInt64(input *[]int64) *[]int64 {
// 	if *input == nil {
// 		return &[]int64{}
// 	}
// 	return input
// }

func checkForNilString(input *[]string) *[]string {
	if *input == nil {
		return &[]string{}
	}
	return input
}
