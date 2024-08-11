//go:generate compogen readme ./config ./README.mdx --extraContents TASK_TEXT_GENERATION_CHAT=.compogen/setup-hosting.mdx
package ollama

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
)

const (
	TaskTextGenerationChat = "TASK_TEXT_GENERATION_CHAT"
	TaskTextEmbeddings     = "TASK_TEXT_EMBEDDINGS"
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

type OllamaSetup struct {
	AutoPull bool   `json:"auto-pull"`
	Endpoint string `json:"endpoint"`
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

type OllamaClientInterface interface {
	Chat(ChatRequest) (ChatResponse, error)
	Embed(EmbedRequest) (EmbedResponse, error)
	IsAutoPull() bool
}

type execution struct {
	base.ComponentExecution
	client  OllamaClientInterface
	execute func(*structpb.Struct) (*structpb.Struct, error)
}

func (e *execution) Execute(ctx context.Context, in base.InputReader, out base.OutputWriter) error {
	inputs, err := in.Read(ctx)
	if err != nil {
		return err
	}
	outputs := make([]*structpb.Struct, len(inputs))

	// The execution takes a array of inputs and returns an array of outputs. The execution is done sequentially.
	for i, input := range inputs {
		output, err := e.execute(input)
		if err != nil {
			return err
		}

		outputs[i] = output
	}

	return out.Write(ctx, outputs)
}

func (c *component) CreateExecution(x base.ComponentExecution) (base.IExecution, error) {
	setupStruct := &OllamaSetup{}
	if err := base.ConvertFromStructpb(x.Setup, setupStruct); err != nil {
		return nil, fmt.Errorf("error parsing setup, %v", err)
	}

	e := &execution{
		ComponentExecution: x,
		client:             NewClient(setupStruct.Endpoint, setupStruct.AutoPull, c.Logger),
	}
	switch x.Task {
	case TaskTextGenerationChat:
		e.execute = e.TaskTextGenerationChat
	case TaskTextEmbeddings:
		e.execute = e.TaskTextEmbeddings
	default:
		return nil, fmt.Errorf("unsupported task")
	}
	return e, nil
}
