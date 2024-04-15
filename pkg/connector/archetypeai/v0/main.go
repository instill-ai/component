//go:generate compogen readme --connector ./config ./README.mdx
package archetypeai

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/gofrs/uuid"
	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/component/pkg/connector/util"
	"github.com/instill-ai/component/pkg/connector/util/httpclient"
	"github.com/instill-ai/x/errmsg"
)

const (
	taskDescribe   = "TASK_DESCRIBE"
	taskSummarize  = "TASK_SUMMARIZE"
	taskUploadFile = "TASK_UPLOAD_FILE"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte

	once     sync.Once
	baseConn base.IConnector
)

type connector struct {
	base.Connector
}

type execution struct {
	base.Execution
	execute func(*structpb.Struct) (*structpb.Struct, error)
	client  *httpclient.Client
}

// Init returns an implementation of IConnector that interacts with Archetype
// AI.
func Init(logger *zap.Logger) base.IConnector {
	once.Do(func() {
		baseConn = &connector{
			Connector: base.Connector{
				Component: base.Component{Logger: logger},
			},
		}
		if err := baseConn.LoadConnectorDefinition(definitionJSON, tasksJSON, nil); err != nil {
			logger.Fatal(err.Error())
		}
	})

	return baseConn
}

// CreateExecution returns an IExecution that executes tasks in Archetype AI.
func (c *connector) CreateExecution(defUID uuid.UUID, task string, config *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	e := &execution{
		client: newClient(config, logger),
	}

	switch task {
	case taskDescribe:
		e.execute = e.describe
	case taskSummarize:
		e.execute = e.summarize
	case taskUploadFile:
		e.execute = e.uploadFile
	default:
		return nil, errmsg.AddMessage(
			fmt.Errorf("not supported task: %s", task),
			fmt.Sprintf("%s task is not supported.", task),
		)
	}

	e.Execution = base.CreateExecutionHelper(e, c, defUID, task, config, logger)

	return e, nil
}

// Execute performs calls the Archetype AI API to execute a task.
func (e *execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
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

func (e *execution) describe(in *structpb.Struct) (*structpb.Struct, error) {
	params := fileQueryParams{}
	if err := base.ConvertFromStructpb(in, &params); err != nil {
		return nil, err
	}

	// We have a 1-1 mapping between the VDP user input and the Archetype AI
	// request. If this stops being the case in the future, we'll need a
	// describeReq structure.
	resp := describeResp{}
	req := e.client.R().SetBody(params).SetResult(&resp)

	if _, err := req.Post(describePath); err != nil {
		return nil, err
	}

	// Archetype AI might return a 200 status even if the operation failed
	// (e.g. if the file doesn't exist).
	if resp.Status != statusCompleted {
		return nil, errmsg.AddMessage(
			fmt.Errorf("response with non-completed status"),
			fmt.Sprintf(`Archetype AI didn't complete query %s: status is "%s".`, resp.QueryID, resp.Status),
		)
	}

	out, err := base.ConvertToStructpb(describeOutput{
		Descriptions: resp.Response,
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (e *execution) summarize(in *structpb.Struct) (*structpb.Struct, error) {
	params := fileQueryParams{}
	if err := base.ConvertFromStructpb(in, &params); err != nil {
		return nil, err
	}

	// We have a 1-1 mapping between the VDP user input and the Archetype AI
	// request. If this stops being the case in the future, we'll need a
	// summarizeReq structure.
	resp := summarizeResp{}
	req := e.client.R().SetBody(params).SetResult(&resp)

	if _, err := req.Post(summarizePath); err != nil {
		return nil, err
	}

	// Archetype AI might return a 200 status even if the operation failed
	// (e.g. if the file doesn't exist).
	if resp.Status != statusCompleted {
		return nil, errmsg.AddMessage(
			fmt.Errorf("response with non-completed status"),
			fmt.Sprintf(`Archetype AI didn't complete query %s: status is "%s".`, resp.QueryID, resp.Status),
		)
	}

	out, err := base.ConvertToStructpb(summarizeOutput{
		Response: resp.Response.ProcessedText,
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (e *execution) uploadFile(in *structpb.Struct) (*structpb.Struct, error) {
	params := uploadFileParams{}
	if err := base.ConvertFromStructpb(in, &params); err != nil {
		return nil, err
	}

	resp := uploadFileResp{}
	req := e.client.R().SetResult(&resp)

	b, err := util.DecodeBase64(params.File)
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	req.SetFileReader("file", id.String(), bytes.NewReader(b))
	if _, err := req.Post(uploadFilePath); err != nil {
		return nil, err
	}

	if !resp.IsValid {
		errMsg := "invalid file."
		if len(resp.Errors) > 0 {
			errMsg = strings.Join(resp.Errors, " ")
		}

		return nil, errmsg.AddMessage(
			fmt.Errorf("file upload failed"),
			fmt.Sprintf(`Couldn't complete upload: %s`, errMsg),
		)
	}

	out, err := base.ConvertToStructpb(resp.uploadFileOutput)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// Test checks the connectivity of the connector.
func (c *connector) Test(_ uuid.UUID, _ *structpb.Struct, _ *zap.Logger) error {
	// TODO Archetype AI API is not public yet. We could test the connection
	// by calling one of the endpoints used in the available tasks. However,
	// these are not designed for specifically for this purpose. When we know
	// of an endpoint that's more suited for this, it should be used instead.
	return nil
}

func getAPIKey(config *structpb.Struct) string {
	return config.GetFields()["api_key"].GetStringValue()
}

// getBasePath returns Archetype AI's API URL. This configuration param allows
// us to override the API the connector will point to. It isn't meant to be
// exposed to users. Rather, it can serve to test the logic against a fake
// server.
// TODO instead of having the API value hardcoded in the codebase, it should
// be read from a config file or environment variable.
func getBasePath(config *structpb.Struct) string {
	v, ok := config.GetFields()["base_path"]
	if !ok {
		return host
	}
	return v.GetStringValue()
}
