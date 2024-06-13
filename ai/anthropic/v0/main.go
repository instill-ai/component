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
	textGenerationTask = "TASK_TEXT_GENERATION_CHAT"
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

	messages := []message{}

	if in.Fields["chat_history"] != nil {
		for _, el := range in.Fields["chat_history"].GetListValue().GetValues() {

			contents := []content{}
			for _, cn := range el.GetStructValue().Fields["content"].GetListValue().GetValues() {

				content_type := cn.GetStructValue().Fields["type"].GetStringValue()
				// anthrothpic models does not support image urls
				if content_type == "text" {
					content := content{
						Type: "text",
						Text: cn.GetStructValue().Fields["text"].GetStringValue(),
					}
					contents = append(contents, content)
				}
			}
			message := message{
				Role:    el.GetStructValue().Fields["role"].GetStringValue(),
				Content: contents,
			}
			messages = append(messages, message)
		}
	}

	final_message := message{
		Role:    "user",
		Content: []content{{Type: "text", Text: prompt}},
	}

	if in.Fields["prompt_images"] != nil {
		for _, image := range in.Fields["prompt_images"].GetListValue().GetValues() {
			// need to add support for different file types in the future.
			image := content{
				Type:   "image",
				Source: &source{Type: "base64", MediaType: "image/jpeg", Data: base.TrimBase64Mime(image.GetStringValue())},
			}
			final_message.Content = append(final_message.Content, image)
		}
	}

	messages = append(messages, final_message)

	body := messagesReq{
		Messages:    messages,
		Model:       in.Fields["model_name"].GetStringValue(),
		MaxTokens:   int(in.Fields["max_new_tokens"].GetNumberValue()),
		System:      in.Fields["system_message"].GetStringValue(),
		TopK:        int(in.Fields["top_k"].GetNumberValue()),
		Temperature: float32(in.Fields["temperature"].GetNumberValue()),
	}

	resp := messagesResp{}
	req := client.R().SetResult(&resp).SetBody(body)
	if _, err := req.Post(messagesPath); err != nil {
		fmt.Println("#### request body ###")
		j, _ := json.MarshalIndent(body, "", "\t")
		fmt.Println(string(j))
		return in, err
	}

	outputStruct := messagesOutput{
		Text: "",
	}
	for _, c := range resp.Content {
		outputStruct.Text += c.Text
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
