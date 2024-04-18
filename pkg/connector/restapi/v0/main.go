package restapi

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/x/errmsg"
	"go.uber.org/zap"
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

	once      sync.Once
	connector base.IConnector

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

type Connector struct {
	base.Connector
}

type Execution struct {
	base.Execution
}

func Init(logger *zap.Logger, usageHandler base.UsageHandler) base.IConnector {
	once.Do(func() {
		connector = &Connector{
			Connector: base.Connector{
				Component: base.Component{Logger: logger, UsageHandler: usageHandler},
			},
		}
		err := connector.LoadConnectorDefinition(definitionJSON, tasksJSON, nil)
		if err != nil {
			logger.Fatal(err.Error())
		}
	})
	return connector
}

func (c *Connector) CreateExecution(defUID uuid.UUID, task string, connection *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	e := &Execution{}
	e.Execution = base.CreateExecutionHelper(e, c, defUID, task, connection, logger)
	return e, nil
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

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {

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
		client, err := newClient(e.Config, e.Logger)
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

func (c *Connector) Test(defUID uuid.UUID, config *structpb.Struct, logger *zap.Logger) error {
	// we don't need to validate the connection since no url setting here
	return nil
}

func (c *Connector) GetConnectorDefinitionByID(defID string, component *pipelinePB.ConnectorComponent) (*pipelinePB.ConnectorDefinition, error) {
	def, err := c.Connector.GetConnectorDefinitionByID(defID, component)
	if err != nil {
		return nil, err
	}

	return c.GetConnectorDefinitionByUID(uuid.FromStringOrNil(def.Uid), component)
}

// Generate the model_name enum based on the task
func (c *Connector) GetConnectorDefinitionByUID(defUID uuid.UUID, component *pipelinePB.ConnectorComponent) (*pipelinePB.ConnectorDefinition, error) {
	oriDef, err := c.Connector.GetConnectorDefinitionByUID(defUID, component)
	if err != nil {
		return nil, err
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
