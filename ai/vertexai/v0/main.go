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
	TextToImageTask    = "TASK_TEXT_TO_IMAGE"
	ImageToImageTask   = "TASK_IMAGE_TO_IMAGE"
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
	case TextToImageTask, ImageToImageTask:
		e.execute = e.generateImage
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

func concateResponse(resp *genai.GenerateContentResponse) string {
	fullResponse := ""
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			fullResponse = fullResponse + fmt.Sprint(part)
		}
	}
	return fullResponse
}

type ChatMessage struct {
	Role    string              `json:"role"`
	Content []MultiModalContent `json:"content"`
}
type URL struct {
	URL string `json:"url"`
}

type MultiModalContent struct {
	ImageURL URL    `json:"image-url"`
	Text     string `json:"text"`
	Type     string `json:"type"`
}

type textGenerationInput struct {
	ChatHistory  []ChatMessage `json:"chat-history"`
	MaxNewTokens int           `json:"max-new-tokens"`
	ModelName    string        `json:"model-name"`
	Prompt       string        `json:"prompt"`
	PromptImages []string      `json:"prompt-images"`
	Seed         int           `json:"seed"`
	SystemMsg    string        `json:"system-message"`
	Temperature  float64       `json:"temperature"`
	TopK         int           `json:"top-k"`
}

type imageGenerationInput struct {
	CFGScale    float64 `json:"cfg-scale"`
	ModelName   string  `json:"model-name"`
	Prompt      string  `json:"prompt"`
	Samples     int     `json:"samples"`
	Seed        int     `json:"seed"`
	Steps       int     `json:"steps"`
	PromptImage string  `json:"prompt-image"`
}

type imageGenerationOutput struct {
	Images []string `json:"images"`
}

type vertexAISetup struct {
	ProjectID string `json:"project-id"`
	Cred      string `json:"google-credential"`
	Location  string `json:"location"`
}

type textGenerationOutput struct {
	Text  string              `json:"text"`
	Usage textGenerationUsage `json:"usage"`
}
type textGenerationUsage struct {
	InputTokens  int `json:"input-tokens"`
	OutputTokens int `json:"output-tokens"`
}

func (e *execution) generateText(in *structpb.Struct) (*structpb.Struct, error) {
	setupStruct := vertexAISetup{}
	err := base.ConvertFromStructpb(e.GetSetup(), &setupStruct)
	if err != nil {
		return nil, err
	}
	inputStruct := textGenerationInput{}
	err = base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	// Temporary way to get the credential. It should be migrated to a more secure implementation.
	credJSON := []byte(setupStruct.Cred)
	client, err := genai.NewClient(ctx, setupStruct.ProjectID, setupStruct.Location, option.WithCredentialsJSON(credJSON))

	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	defer client.Close()

	model := client.GenerativeModel(inputStruct.ModelName)
	model.SetTemperature(float32(inputStruct.Temperature))
	model.SetMaxOutputTokens(int32(inputStruct.MaxNewTokens))
	model.SetTopK(float32(inputStruct.TopK))
	promptPart := genai.Text(inputStruct.Prompt)
	resp, err := model.GenerateContent(ctx, promptPart)
	if err != nil {
		return nil, fmt.Errorf("error generating content: %w", err)
	}

	outputStruct := textGenerationOutput{
		Text: concateResponse(resp),
		Usage: textGenerationUsage{
			InputTokens:  int(resp.UsageMetadata.PromptTokenCount),
			OutputTokens: int(resp.UsageMetadata.CandidatesTokenCount),
		},
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

func (e *execution) generateImage(in *structpb.Struct) (*structpb.Struct, error) {
	setupStruct := vertexAISetup{}
	err := base.ConvertFromStructpb(e.GetSetup(), &setupStruct)
	if err != nil {
		return nil, err
	}
	inputStruct := imageGenerationInput{}
	err = base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}
	outputStruct := imageGenerationOutput{}
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
