//go:generate compogen readme ./config ./README.mdx
package base64

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
)

const (
	encode = "TASK_ENCODE"
	decode = "TASK_DECODE"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	once      sync.Once
	comp      *component
)

type component struct {
	base.Component
}

type execution struct {
	base.ComponentExecution
}

type Base64 struct {
	Data string `json:"data"`
}

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

// CreateExecution initializes a connector executor that can be used in a
// pipeline trigger.
func (c *component) CreateExecution(x base.ComponentExecution) (base.IExecution, error) {
	return &execution{ComponentExecution: x}, nil
}

func (e *execution) Execute(ctx context.Context, in base.InputReader, out base.OutputWriter) error {
	inputs, err := in.Read(ctx)
	if err != nil {
		return err
	}
	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		base64Struct := Base64{}
		err := base.ConvertFromStructpb(input, &base64Struct)
		if err != nil {
			return err
		}
		switch e.Task {
		case encode:
			base64Struct.Data = Encode(base64Struct.Data)
		case decode:
			base64Struct.Data, err = Decode(base64Struct.Data)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("not supported task: %s", e.Task)
		}
		outputJSON, err := json.Marshal(base64Struct)
		if err != nil {
			return err
		}
		output := structpb.Struct{}
		err = protojson.Unmarshal(outputJSON, &output)
		if err != nil {
			return err
		}
		outputs = append(outputs, &output)
	}
	return out.Write(ctx, outputs)
}

func Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func Decode(str string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(str))
	if err != nil {
		return str, err
	}
	return string(b), nil
}
