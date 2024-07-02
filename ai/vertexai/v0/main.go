//go:generate compogen readme ./config ./README.mdx
package vertexai

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"cloud.google.com/go/vertexai/genai"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
)

const (
	TextGenerationTask = "TASK_TEXT_GENERATION_CHAT"
	cfgAPIKey          = "api-key"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/setup.json
	setupJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	// This is a temporary location for the credential file. This will be moved to a more secure location.
	//go:embed tmp/cred.json
	credJSON []byte

	once sync.Once
	comp *component
)

type component struct {
	base.Component
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
}

func (c *component) CreateExecution(sysVars map[string]any, setup *structpb.Struct, task string) (*base.ExecutionWrapper, error) {

	e := &execution{
		ComponentExecution: base.ComponentExecution{Component: c, SystemVariables: sysVars, Task: task, Setup: setup},
	}
	switch task {
	case TextGenerationTask:
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

type messagesOutput struct {
	Text string `json:"text"`
}

func concateResponse(resp *genai.GenerateContentResponse) string {
	fullResponse := ""
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			fullResponse = fullResponse + fmt.Sprint(part)
		}
	}
	return fullResponse
}

func (e *execution) generateText(in *structpb.Struct) (*structpb.Struct, error) {
	prompt := in.Fields["prompt"].GetStringValue()
	modelName := "gemini-1.5-flash-001"
	projectID := in.Fields["project-id"].GetStringValue()
	location := "us-central1" // London, check: https://cloud.google.com/vertex-ai/docs/general/locations#europe

	ctx := context.Background()
	// Temporary way to get the credential. It should be migrated to a more secure implementation.
	client, err := genai.NewClient(ctx, projectID, location, option.WithCredentialsJSON(credJSON))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	model := client.GenerativeModel(modelName)
	model.SetTemperature(0.9)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	outputStruct := messagesOutput{
		Text: concateResponse(resp),
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
