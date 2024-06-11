package anthropic

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
)

const (
	textGenerationTask = "TASK_TEXT_GENERATION"
	cfgAPIKey          = "api_key"
	host               = "https://api.anthropic.com"
	messagesPath       = "/v1/messages"
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

type execution struct {
	base.ComponentExecution
	execute func(*structpb.Struct) (*structpb.Struct, error)
}

func (c *component) CreateExecution(sysVars map[string]any, setup *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	e := &execution{
		ComponentExecution: base.ComponentExecution{Component: c, SystemVariables: sysVars, Task: task, Setup: setup},
	}

	// A simple if statement would be enough in a component with a single task.
	// If the number of task grows, here is where the execution task would be
	// selected.
	switch task {
	case textGenerationTask:
		e.execute = e.generateText
	default:
		return nil, fmt.Errorf("unsupported task")
	}
	return &base.ExecutionWrapper{Execution: e}, nil
}

func (e *execution) Execute(_ context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := make([]*structpb.Struct, len(inputs))

	// An execution  might take several inputs. One result will be returned for
	// each one of them, containing the execution output for that set of
	// parameters.
	for i, input := range inputs {
		output, err := e.execute(input)
		if err != nil {
			return nil, err
		}

		outputs[i] = output
	}

	return outputs, nil
}

func (e *execution) generateText(in *structpb.Struct) (*structpb.Struct, error) {
	client := newClient(e.Setup, e.GetLogger())

	prompt := in.Fields["prompt"].GetStringValue()
	model := in.Fields["model"].GetStringValue()
	max_token := 1024

	messages := []message{{
		Role:    "user",
		Content: []content{{Type: "text", Text: prompt}},
	}}

	body := messagesReq{
		Messages:  messages,
		Model:     model,
		MaxTokens: max_token,
	}

	resp := messagesResp{}
	req := client.R().SetResult(&resp).SetBody(body)
	if _, err := req.Post(messagesPath); err != nil {
		//fmt.Println(req.RawRequest)
		return in, err
	}

	outputStruct := messagesOutput{
		Texts: []string{},
		Usage: resp.Usage,
	}
	for _, c := range resp.Content {
		outputStruct.Texts = append(outputStruct.Texts, c.Text)
	}

	outputJSON, err := json.Marshal(outputStruct)
	if err != nil {
		return nil, err
	}
	output := structpb.Struct{}
	err = protojson.Unmarshal(outputJSON, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}
