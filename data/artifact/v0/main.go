//go:generate compogen readme ./config ./README.mdx --extraContents intro=.compogen/extra-intro.mdx
package artifact

import (
	"context"
	"fmt"
	"sync"

	_ "embed"

	"github.com/instill-ai/component/base"
	artifactPB "github.com/instill-ai/protogen-go/artifact/artifact/v1alpha"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	taskUploadFile          string = "TASK_UPLOAD_FILE"
	taskUploadFiles         string = "TASK_UPLOAD_FILES"
	taskGetFilesMetadata    string = "TASK_GET_FILES_METADATA"
	taskGetChunksMetadata   string = "TASK_GET_CHUNKS_METADATA"
	taskGetFileInMarkdown   string = "TASK_GET_FILE_IN_MARKDOWN"
	taskMatchFileStatus     string = "TASK_MATCH_FILE_STATUS"
	taskRetrieve            string = "TASK_RETRIEVE"
	taskAsk                 string = "TASK_ASK"
	taskRetrieveChatHistory string = "TASK_RETRIEVE_CHAT_HISTORY"
	taskWriteChatMessage    string = "TASK_WRITE_CHAT_MESSAGE"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	once      sync.Once
	comp      *component
)

type component struct {
	base.Component
}

type execution struct {
	base.ComponentExecution

	execute    func(*structpb.Struct) (*structpb.Struct, error)
	client     artifactPB.ArtifactPublicServiceClient
	connection Connection
}

func Init(bc base.Component) *component {
	once.Do(func() {
		comp = &component{Component: bc}
		err := comp.LoadDefinition(definitionJSON, nil, tasksJSON, nil)
		if err != nil {
			panic(err)
		}
	})
	return comp
}

func (c *component) CreateExecution(x base.ComponentExecution) (base.IExecution, error) {
	e := &execution{ComponentExecution: x}

	client, connection, err := initArtifactClient(getArtifactServerURL(e.SystemVariables))

	if err != nil {
		return nil, fmt.Errorf("failed to create client connection: %w", err)
	}

	e.client, e.connection = client, connection

	switch x.Task {
	case taskUploadFile:
		e.execute = e.uploadFile
	case taskUploadFiles:
		e.execute = e.uploadFiles
	case taskGetFilesMetadata:
		e.execute = e.getFilesMetadata
	case taskGetChunksMetadata:
		e.execute = e.getChunksMetadata
	case taskGetFileInMarkdown:
		e.execute = e.getFileInMarkdown
	case taskMatchFileStatus:
		e.execute = e.matchFileStatus
	case taskRetrieve:
		e.execute = e.searchChunks
	case taskAsk:
		e.execute = e.query
	case taskRetrieveChatHistory:
		e.execute = e.retrieveChatHistory
	case taskWriteChatMessage:
		e.execute = e.writeChatMessage
	default:
		return nil, fmt.Errorf(x.Task + " task is not supported.")
	}

	return e, nil
}

func (e *execution) Execute(_ context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := make([]*structpb.Struct, len(inputs))

	for i, input := range inputs {
		output, err := e.execute(input)
		if err != nil {
			return nil, err
		}

		outputs[i] = output
	}

	return outputs, nil
}
