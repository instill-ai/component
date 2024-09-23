//go:generate compogen readme ./config ./README.mdx
package collection

import (
	"context"
	"fmt"
	"sync"

	_ "embed"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/x/errmsg"
)

const (
	taskDeclare     = "TASK_DECLARE"
	taskAppendArray = "TASK_APPEND_ARRAY"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
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
}

// Init returns an implementation of IOperator that processes JSON objects.
func Init(bc base.Component) *component {
	once.Do(func() {
		comp = &component{Component: bc}
		err := comp.LoadDefinition(definitionJSON, nil, tasksJSON, nil)
		if err != nil {
			panic(err)
		}
	})
	return comp
}

// CreateExecution initializes a component executor that can be used in a
// pipeline trigger.
func (c *component) CreateExecution(x base.ComponentExecution) (base.IExecution, error) {
	e := &execution{ComponentExecution: x}

	switch x.Task {
	case taskDeclare:
		e.execute = e.declare
	case taskAppendArray:
		e.execute = e.appendArray
	default:
		return nil, errmsg.AddMessage(
			fmt.Errorf("not supported task: %s", x.Task),
			fmt.Sprintf("%s task is not supported.", x.Task),
		)
	}
	return e, nil
}

func (e *execution) declare(in *structpb.Struct) (*structpb.Struct, error) {
	out := in
	return out, nil
}

func (e *execution) appendArray(in *structpb.Struct) (*structpb.Struct, error) {
	arr := in.Fields["array"]
	data := in.Fields["element"]
	arr.GetListValue().Values = append(arr.GetListValue().Values, data)

	out := &structpb.Struct{Fields: make(map[string]*structpb.Value)}
	out.Fields["array"] = arr
	return out, nil
}

func (e *execution) Execute(ctx context.Context, jobs []*base.Job) error {
	return base.SequentialExecutor(ctx, jobs, e.execute)
}
