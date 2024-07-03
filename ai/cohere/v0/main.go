//go:generate compogen readme ./config ./README.mdx
package cohere

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	cohereSDK "github.com/cohere-ai/cohere-go/v2"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	TextGenerationTask = "TASK_TEXT_GENERATION_CHAT"
	TextEmbeddingTask  = "TASK_TEXT_EMBEDDINGS"
	TextRerankTask     = "TASK_TEXT_RERANKING"
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
	generateEmbedding(request cohereSDK.EmbedRequest) (cohereSDK.EmbedResponse, error)
	generateRerank(request cohereSDK.RerankRequest) (cohereSDK.RerankResponse, error)
}

type cohereUsage struct {
	InputTokens  int `json:"input-tokens"`
	OutputTokens int `json:"output-tokens"`
}

// These structs are used to send the request /  parse the response from the API, this following their naming convension.
// reference: https://docs.anthropic.com/en/api/messages

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
	case TextGenerationTask:
		e.execute = e.taskTextGeneration
	case TextEmbeddingTask:
		e.execute = e.taskEmbedding
	case TextRerankTask:
		e.execute = e.taskRerank
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
