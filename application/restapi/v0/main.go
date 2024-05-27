//go:generate compogen readme --connector ./config ./README.mdx
package restapi

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/x/errmsg"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

const (
	taskGet     = "TASK_GET"
	taskPost    = "TASK_POST"
	taskPatch   = "TASK_PATCH"
	taskPut     = "TASK_PUT"
	taskDelete  = "TASK_DELETE"
	taskHead    = "TASK_HEAD"
	taskOptions = "TASK_OPTIONS"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte

	//go:embed config/tasks.json
	tasksJSON []byte

	once sync.Once
	con  *connector

	taskMethod = map[string]string{
		taskGet:     http.MethodGet,
		taskPost:    http.MethodPost,
		taskPatch:   http.MethodPatch,
		taskPut:     http.MethodPut,
		taskDelete:  http.MethodDelete,
		taskHead:    http.MethodHead,
		taskOptions: http.MethodOptions,
	}
)

type connector struct {
	base.Connector
}

type execution struct {
	base.ConnectorExecution
}

func Init(bc base.Connector) *connector {
	once.Do(func() {
		con = &connector{Connector: bc}
		err := con.LoadConnectorDefinition(definitionJSON, tasksJSON, nil)
		if err != nil {
			panic(err)
		}
	})
	return con
}

func (c *connector) CreateExecution(sysVars map[string]any, connection *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	return &base.ExecutionWrapper{Execution: &execution{
		ConnectorExecution: base.ConnectorExecution{Connector: c, SystemVariables: sysVars, Connection: connection, Task: task},
	}}, nil
}

func getAuthentication(config *structpb.Struct) (authentication, error) {
	auth := config.GetFields()["authentication"].GetStructValue()
	authType := auth.GetFields()["auth_type"].GetStringValue()

	switch authType {
	case string(noAuthType):
		authStruct := noAuth{}
		err := base.ConvertFromStructpb(auth, &authStruct)
		if err != nil {
			return nil, err
		}
		return authStruct, nil
	case string(basicAuthType):
		authStruct := basicAuth{}
		err := base.ConvertFromStructpb(auth, &authStruct)
		if err != nil {
			return nil, err
		}
		return authStruct, nil
	case string(apiKeyType):
		authStruct := apiKeyAuth{}
		err := base.ConvertFromStructpb(auth, &authStruct)
		if err != nil {
			return nil, err
		}
		return authStruct, nil
	case string(bearerTokenType):
		authStruct := bearerTokenAuth{}
		err := base.ConvertFromStructpb(auth, &authStruct)
		if err != nil {
			return nil, err
		}
		return authStruct, nil
	default:
		return nil, errors.New("invalid authentication type")
	}
}

func (e *execution) Execute(_ context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {

	method, ok := taskMethod[e.Task]
	if !ok {
		return nil, errmsg.AddMessage(
			fmt.Errorf("not supported task: %s", e.Task),
			fmt.Sprintf("%s task is not supported.", e.Task),
		)
	}

	outputs := []*structpb.Struct{}
	for _, input := range inputs {
		taskIn := TaskInput{}
		taskOut := TaskOutput{}

		if err := base.ConvertFromStructpb(input, &taskIn); err != nil {
			return nil, err
		}

		// We may have different url in batch.
		client, err := newClient(e.Connection, e.GetLogger())
		if err != nil {
			return nil, err
		}

		// An API error is a valid output in this connector.
		req := client.R().SetResult(&taskOut.Body).SetError(&taskOut.Body)
		if taskIn.Body != nil {
			req.SetBody(taskIn.Body)
		}

		resp, err := req.Execute(method, taskIn.EndpointURL)
		if err != nil {
			return nil, err
		}

		taskOut.StatusCode = resp.StatusCode()
		taskOut.Header = resp.Header()

		output, err := base.ConvertToStructpb(taskOut)
		if err != nil {
			return nil, err
		}

		outputs = append(outputs, output)
	}
	return outputs, nil
}

func (c *connector) Test(sysVars map[string]any, connection *structpb.Struct) error {
	// we don't need to validate the connection since no url setting here
	return nil
}

// Generate the model_name enum based on the task
func (c *connector) GetConnectorDefinition(sysVars map[string]any, component *pipelinePB.ConnectorComponent) (*pipelinePB.ConnectorDefinition, error) {
	oriDef, err := c.Connector.GetConnectorDefinition(nil, nil)
	if err != nil {
		return nil, err
	}
	if sysVars == nil && component == nil {
		return oriDef, nil
	}

	def := proto.Clone(oriDef).(*pipelinePB.ConnectorDefinition)
	if component == nil {
		return def, nil
	}
	if component.Task == "" {
		return def, nil
	}
	if _, ok := component.Input.Fields["output_body_schema"]; !ok {
		return def, nil
	}

	schStr := component.Input.Fields["output_body_schema"].GetStringValue()
	sch := &structpb.Struct{}
	_ = json.Unmarshal([]byte(schStr), sch)
	spec := def.Spec.DataSpecifications[component.Task]
	spec.Output = sch
	return def, nil
}
