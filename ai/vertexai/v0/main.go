//go:generate compogen readme ./config ./README.mdx
package vertexai

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
)

const (
	TextGenerationTask    = "TASK_TEXT_GENERATION_CHAT"
	TextToImageTask       = "TASK_TEXT_TO_IMAGE"
	ImageToImageTask      = "TASK_IMAGE_TO_IMAGE"
	SpeechRecognitionTask = "TASK_SPEECH_RECOGNITION"
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
		return nil, fmt.Errorf("image generation is curently under approval process, contact support for more information")
	case SpeechRecognitionTask:
		e.execute = e.speechRecognition
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

type vertexAISetup struct {
	ProjectID string `json:"project-id"`
	Cred      string `json:"google-credential"`
	Location  string `json:"location"`
}
