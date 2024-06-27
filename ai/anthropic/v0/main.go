//go:generate compogen readme ./config ./README.mdx
package anthropic

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"slices"
	"sync"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
)

const (
	textGenerationTask = "TASK_TEXT_GENERATION_CHAT"
	cfgAPIKey          = "api-key"
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

	supportedImageExtensions = []string{"jpeg", "png", "gif", "webp"}
)

type component struct {
	base.Component
}

type AnthropicClient interface {
	generateTextChat(request messagesReq) (messagesResp, error)
}

// These structs are used to send the request /  parse the response from the API, this following their naming convension.
// reference: https://docs.anthropic.com/en/api/messages
type messagesResp struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Role       string    `json:"role"`
	Content    []content `json:"content"`
	Model      string    `json:"model"`
	StopReason string    `json:"stop_reason,omitempty"`
	Usage      usage     `json:"usage"`
}

type messagesReq struct {
	Model         string      `json:"model"`
	Messages      []message   `json:"messages"`
	MaxTokens     int         `json:"max_tokens"`
	Metadata      interface{} `json:"metadata"`
	StopSequences []string    `json:"stop_sequences,omitempty"`
	Stream        bool        `json:"stream,omitempty"`
	System        string      `json:"system,omitempty"`
	Temperature   float32     `json:"temperature,omitempty"`
	TopK          int         `json:"top_k,omitempty"`
	TopP          float32     `json:"top_p,omitempty"`
}

type messagesOutput struct {
	Text string `json:"text"`
}

type message struct {
	Role    string    `json:"role"`
	Content []content `json:"content"`
}

type usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// doesn't support anthropic tools at the moment
type content struct {
	Type   string  `json:"type"`
	Text   string  `json:"text,omitempty"`
	Source *source `json:"source,omitempty"`
}

type source struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
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

type execution struct {
	base.ComponentExecution
	execute func(*structpb.Struct) (*structpb.Struct, error)
	client  AnthropicClient
}

func (c *component) CreateExecution(sysVars map[string]any, setup *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	e := &execution{
		ComponentExecution: base.ComponentExecution{Component: c, SystemVariables: sysVars, Task: task, Setup: setup},
		client:             newClient(getAPIKey(setup), getBasePath(setup), c.GetLogger()),
	}
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

	// The execution takes a array of inputs and returns an array of outputs. The execution is done sequentially.
	for i, input := range inputs {
		output, err := e.execute(input)
		if err != nil {
			return nil, err
		}

		outputs[i] = output
	}

	return outputs, nil
}

func retriveChatMessage(chatHistoryPbList *structpb.ListValue) []message {
	messages := []message{}
	for _, messagePbValue := range chatHistoryPbList.GetValues() {
		contents := []content{}
		for _, contentPbValue := range messagePbValue.GetStructValue().Fields["content"].GetListValue().GetValues() {
			contentType := contentPbValue.GetStructValue().Fields["type"].GetStringValue()
			// anthrothpic models does not support image urls
			if contentType == "text" {
				content := content{
					Type: "text",
					Text: contentPbValue.GetStructValue().Fields["text"].GetStringValue(),
				}
				contents = append(contents, content)
			}
		}
		completeMessage := message{Role: messagePbValue.GetStructValue().Fields["role"].GetStringValue(), Content: contents}
		messages = append(messages, completeMessage)
	}
	return messages
}

func (e *execution) generateText(in *structpb.Struct) (*structpb.Struct, error) {

	prompt := in.Fields["prompt"].GetStringValue()

	messages := []message{}

	chatHistory := in.Fields["chat-history"].GetListValue()

	if chatHistory != nil {
		messages = retriveChatMessage(chatHistory)
	}

	finalMessage := message{
		Role:    "user",
		Content: []content{{Type: "text", Text: prompt}},
	}

	if in.Fields["prompt-images"] != nil {
		for _, image := range in.Fields["prompt-images"].GetListValue().GetValues() {
			extension := base.GetBase64FileExtension(image.GetStringValue())
			// check if the image extension is supported
			if !slices.Contains(supportedImageExtensions, extension) {
				return nil, fmt.Errorf("unsupported image extension, expected one of: %v , got %s", supportedImageExtensions, extension)
			}
			image := content{
				Type:   "image",
				Source: &source{Type: "base64", MediaType: fmt.Sprintf("image/%s", extension), Data: base.TrimBase64Mime(image.GetStringValue())},
			}
			finalMessage.Content = append(finalMessage.Content, image)
		}
	}

	messages = append(messages, finalMessage)

	req := messagesReq{
		Messages:    messages,
		Model:       in.Fields["model-name"].GetStringValue(),
		MaxTokens:   int(in.Fields["max-new-tokens"].GetNumberValue()),
		System:      in.Fields["system-message"].GetStringValue(),
		TopK:        int(in.Fields["top-k"].GetNumberValue()),
		Temperature: float32(in.Fields["temperature"].GetNumberValue()),
	}

	resp, err := e.client.generateTextChat(req)

	if err != nil {
		return nil, err
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