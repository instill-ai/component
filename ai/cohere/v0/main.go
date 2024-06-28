//go:generate compogen readme ./config ./README.mdx
package cohere

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"slices"
	"sync"

	cohereSDK "github.com/cohere-ai/cohere-go/v2"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	textGenerationTask = "TASK_TEXT_GENERATION_CHAT"
	cfgAPIKey          = "api-key"
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

type CohereClient interface {
	generateTextChat(request cohereSDK.ChatRequest) (cohereSDK.NonStreamedChatResponse, error)
}

// These structs are used to send the request /  parse the response from the API, this following their naming convension.
// reference: https://docs.anthropic.com/en/api/messages

type ciatation struct {
	Start int    `json:"start"`
	End   int    `json:"end"`
	Text  string `json:"text"`
}

type messagesOutput struct {
	Text       string      `json:"text"`
	Ciatations []ciatation `json:"citations"`
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
	client  CohereClient
}

func (c *component) CreateExecution(sysVars map[string]any, setup *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	e := &execution{
		ComponentExecution: base.ComponentExecution{Component: c, SystemVariables: sysVars, Task: task, Setup: setup},
		client:             newClient(getAPIKey(setup), c.GetLogger()),
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

func (e *execution) generateText(in *structpb.Struct) (*structpb.Struct, error) {

	prompt := in.Fields["prompt"].GetStringValue()

	messages := []*cohereSDK.Message{}

	systemPrompt := in.Fields["system-prompt"].GetStringValue()

	if systemPrompt != "" {
		message := cohereSDK.Message{}
		message.Role = "SYSTEM"
		message.Chatbot = &cohereSDK.ChatMessage{Message: systemPrompt}
		messages = slices.Insert(messages, 0, &message)
	}

	documents := []map[string]string{}
	if in.Fields["documents"] != nil {
		for _, doc := range in.Fields["documents"].GetListValue().Values {
			document := map[string]string{}
			document["text"] = doc.GetStringValue()
			documents = append(documents, document)
		}
	}

	modelName := in.Fields["model-name"].GetStringValue()
	maxTokens := int(in.Fields["max-new-tokens"].GetNumberValue())
	temperature := float64(in.Fields["temperature"].GetNumberValue())
	topK := int(in.Fields["top-k"].GetNumberValue())
	seed := int(in.Fields["seed"].GetNumberValue())

	// This is a mock data for the documents for testing purposes
	mockDocuments := []map[string]string{
		{"text": "Earth isn’t actually round."},
		{"text": "Coral reefs are Earth’s largest living structure."},
		{"text": "The Great Wall of China is not visible from space."},
		{"text": "Humans have more than five senses."},
		{"text": "Antarctica is home to the largest ice sheet on Earth."},
		{"text": "The Moon is drifting away from Earth."},
		{"text": "The Great Pyramid of Giza is not the tallest pyramid in the world."},
		{"text": "The Earth’s core is as hot as the surface of the Sun."},
		{"text": "The Earth’s magnetic poles are not fixed."},
		{"text": "The Earth’s atmosphere is mostly nitrogen."},
	}
	documents = append(documents, mockDocuments...)

	req := cohereSDK.ChatRequest{
		Message:     prompt,
		Model:       &modelName,
		ChatHistory: messages,
		MaxTokens:   &maxTokens,
		Temperature: &temperature,
		K:           &topK,
		Seed:        &seed,
		Documents:   documents,
	}

	resp, err := e.client.generateTextChat(req)

	if err != nil {
		return nil, err
	}

	citations := []ciatation{}

	for _, c := range resp.Citations {
		citation := ciatation{
			Start: c.Start,
			End:   c.End,
			Text:  c.Text,
		}
		citations = append(citations, citation)
	}

	print("### Received text: ")
	fmt.Printf("%s\n", resp.Text)

	outputStruct := messagesOutput{
		Text:       resp.Text,
		Ciatations: citations,
	}

	outputJSON, err := json.Marshal(outputStruct)
	if err != nil {
		return nil, err
	}
	println("### Output JSON: ")
	println(string(outputJSON))
	output := structpb.Struct{}
	err = protojson.Unmarshal(outputJSON, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}
