//go:generate compogen readme --operator ./config ./README.mdx
package base64

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
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
	operator  base.IOperator
)

type Operator struct {
	base.Operator
}

type Execution struct {
	base.Execution
}

type Base64 struct {
	Data string `json:"data"`
}

func Init(logger *zap.Logger) base.IOperator {
	once.Do(func() {
		operator = &Operator{
			Operator: base.Operator{
				Component: base.Component{Logger: logger},
			},
		}
		err := operator.LoadOperatorDefinition(definitionJSON, tasksJSON, nil)
		if err != nil {
			logger.Fatal(err.Error())
		}
	})
	return operator
}

func (o *Operator) CreateExecution(defUID uuid.UUID, task string, config *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	e := &Execution{}
	e.Execution = base.CreateExecutionHelper(e, o, defUID, task, config, logger)
	return e, nil
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		base64Struct := Base64{}
		err := base.ConvertFromStructpb(input, &base64Struct)
		if err != nil {
			return nil, err
		}
		switch e.Task {
		case encode:
			base64Struct.Data = Encode(base64Struct.Data)
		case decode:
			base64Struct.Data, err = Decode(base64Struct.Data)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("not supported task: %s", e.Task)
		}
		outputJSON, err := json.Marshal(base64Struct)
		if err != nil {
			return nil, err
		}
		output := structpb.Struct{}
		err = protojson.Unmarshal(outputJSON, &output)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, &output)
	}
	return outputs, nil
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
