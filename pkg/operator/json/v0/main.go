//go:generate compogen readme --operator ./config ./README.mdx
package json

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/itchyny/gojq"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/x/errmsg"
)

const (
	taskMarshal   = "TASK_MARSHAL"
	taskUnmarshal = "TASK_UNMARSHAL"
	taskJQ        = "TASK_JQ"
)

var (
	//go:embed config/definitions.json
	definitionsJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte

	once sync.Once
	op   base.IOperator
)

type operator struct {
	base.Operator
}

type execution struct {
	base.Execution
	execute func(*structpb.Struct) (*structpb.Struct, error)
}

// Init returns an implementation of IOperator that processes JSON objects.
func Init(logger *zap.Logger) base.IOperator {
	once.Do(func() {
		op = &operator{
			Operator: base.Operator{
				Component: base.Component{Logger: logger},
			},
		}
		err := op.LoadOperatorDefinitions(definitionsJSON, tasksJSON, nil)
		if err != nil {
			logger.Fatal(err.Error())
		}
	})
	return op
}

func (o *operator) CreateExecution(defUID uuid.UUID, task string, config *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	e := &execution{}

	switch task {
	case taskMarshal:
		e.execute = e.marshal
	case taskUnmarshal:
		e.execute = e.unmarshal
	case taskJQ:
		e.execute = e.jq
	default:
		return nil, errmsg.AddMessage(
			fmt.Errorf("not supported task: %s", task),
			fmt.Sprintf("%s task is not supported.", task),
		)
	}

	e.Execution = base.CreateExecutionHelper(e, o, defUID, task, config, logger)

	return e, nil
}

func (e *execution) marshal(in *structpb.Struct) (*structpb.Struct, error) {
	out := new(structpb.Struct)

	b, err := protojson.Marshal(in.Fields["json"])
	if err != nil {
		return nil, errmsg.AddMessage(err, "Couldn't convert the provided object to JSON.")
	}

	out.Fields = map[string]*structpb.Value{
		"string": structpb.NewStringValue(string(b)),
	}

	return out, nil
}

func (e *execution) unmarshal(in *structpb.Struct) (*structpb.Struct, error) {
	out := new(structpb.Struct)

	b := []byte(in.Fields["string"].GetStringValue())
	obj := new(structpb.Struct)
	if err := protojson.Unmarshal(b, obj); err != nil {
		return nil, errmsg.AddMessage(err, "Couldn't parse the JSON string. Please check the syntax is correct.")
	}

	out.Fields = map[string]*structpb.Value{
		"json": structpb.NewStructValue(obj),
	}

	return out, nil
}

func (e *execution) jq(in *structpb.Struct) (*structpb.Struct, error) {
	out := new(structpb.Struct)

	b := []byte(in.Fields["jsonInput"].GetStringValue())
	var input any
	if err := json.Unmarshal(b, &input); err != nil {
		return nil, errmsg.AddMessage(err, "Couldn't parse the JSON input. Please check the syntax is correct.")
	}

	queryStr := in.Fields["jqFilter"].GetStringValue()
	q, err := gojq.Parse(queryStr)
	if err != nil {
		// Error messages from gojq are human-friendly enough.
		msg := fmt.Sprintf("Couldn't parse the jq filter: %s. Please check the syntax is correct.", err.Error())
		return nil, errmsg.AddMessage(err, msg)
	}

	results := []any{}
	iter := q.Run(input)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}

		if err, ok := v.(error); ok {
			msg := fmt.Sprintf("Couldn't apply the jq filter: %s.", err.Error())
			return nil, errmsg.AddMessage(err, msg)
		}

		results = append(results, v)
	}

	list, err := structpb.NewList(results)
	if err != nil {
		return nil, err
	}

	out.Fields = map[string]*structpb.Value{
		"results": structpb.NewListValue(list),
	}

	return out, nil
}

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
