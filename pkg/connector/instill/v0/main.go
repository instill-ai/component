package instill

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"

	commonPB "github.com/instill-ai/protogen-go/common/task/v1alpha"
	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

const (
	internalMode = "Internal Mode"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	once      sync.Once
	connector base.IConnector
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

func getMode(config *structpb.Struct) string {
	return config.GetFields()["mode"].GetStringValue()
}

func getAPIKey(config *structpb.Struct) string {
	if getMode(config) == internalMode {
		return config.GetFields()["header_authorization"].GetStringValue()
	}
	return fmt.Sprintf("Bearer %s", config.GetFields()["api_token"].GetStringValue())
}
func getInstillUserUID(config *structpb.Struct) string {
	return config.GetFields()["instill_user_uid"].GetStringValue()
}

func getModelServerURL(config *structpb.Struct) string {
	if getMode(config) == internalMode {
		return config.GetFields()["instill_model_backend"].GetStringValue()
	}
	serverURL := config.GetFields()["server_url"].GetStringValue()
	if strings.HasPrefix(serverURL, "https://") {
		if len(strings.Split(serverURL, ":")) == 2 {
			serverURL = serverURL + ":443"
		}
	} else if strings.HasPrefix(serverURL, "http://") {
		if len(strings.Split(serverURL, ":")) == 2 {
			serverURL = serverURL + ":80"
		}
	}
	return serverURL
}

func getMgmtServerURL(config *structpb.Struct) string {
	if getMode(config) == internalMode {
		return config.GetFields()["instill_mgmt_backend"].GetStringValue()
	}
	serverURL := config.GetFields()["server_url"].GetStringValue()
	if strings.HasPrefix(serverURL, "https://") {
		if len(strings.Split(serverURL, ":")) == 2 {
			serverURL = serverURL + ":443"
		}
	} else if strings.HasPrefix(serverURL, "http://") {
		if len(strings.Split(serverURL, ":")) == 2 {
			serverURL = serverURL + ":80"
		}
	}
	return serverURL
}
func getRequestMetadata(cfg *structpb.Struct) metadata.MD {
	return metadata.Pairs(
		"Authorization", getAPIKey(cfg),
		"Instill-User-Uid", getInstillUserUID(cfg),
		"Instill-Auth-Type", "user",
	)
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	var err error

	if len(inputs) <= 0 || inputs[0] == nil {
		return inputs, fmt.Errorf("invalid input")
	}

	gRPCCLient, gRPCCLientConn := initModelPublicServiceClient(getModelServerURL(e.Config))
	if gRPCCLientConn != nil {
		defer gRPCCLientConn.Close()
	}

	mgmtGRPCCLient, mgmtGRPCCLientConn := initMgmtPublicServiceClient(getMgmtServerURL(e.Config))
	if mgmtGRPCCLientConn != nil {
		defer mgmtGRPCCLientConn.Close()
	}

	modelNameSplits := strings.Split(inputs[0].GetFields()["model_name"].GetStringValue(), "/")
	ctx := metadata.NewOutgoingContext(context.Background(), getRequestMetadata(e.Config))
	nsResp, err := mgmtGRPCCLient.CheckNamespace(ctx, &mgmtPB.CheckNamespaceRequest{
		Id: modelNameSplits[0],
	})
	if err != nil {
		return nil, err
	}
	nsType := ""
	if nsResp.Type == mgmtPB.CheckNamespaceResponse_NAMESPACE_ORGANIZATION {
		nsType = "organizations"
	} else {
		nsType = "users"
	}

	modelName := fmt.Sprintf("%s/%s/models/%s", nsType, modelNameSplits[0], modelNameSplits[1])

	var result []*structpb.Struct
	switch e.Task {
	case commonPB.Task_TASK_UNSPECIFIED.String():
		result, err = e.executeUnspecified(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_CLASSIFICATION.String():
		result, err = e.executeImageClassification(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_DETECTION.String():
		result, err = e.executeObjectDetection(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_KEYPOINT.String():
		result, err = e.executeKeyPointDetection(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_OCR.String():
		result, err = e.executeOCR(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_INSTANCE_SEGMENTATION.String():
		result, err = e.executeInstanceSegmentation(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_SEMANTIC_SEGMENTATION.String():
		result, err = e.executeSemanticSegmentation(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_TEXT_TO_IMAGE.String():
		result, err = e.executeTextToImage(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_TEXT_GENERATION.String():
		result, err = e.executeTextGeneration(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_TEXT_GENERATION_CHAT.String():
		result, err = e.executeTextGenerationChat(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_VISUAL_QUESTION_ANSWERING.String():
		result, err = e.executeVisualQuestionAnswering(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_IMAGE_TO_IMAGE.String():
		result, err = e.executeImageToImage(gRPCCLient, modelName, inputs)
	default:
		return inputs, fmt.Errorf("unsupported task: %s", e.Task)
	}

	return result, err
}

func (c *Connector) Test(_ uuid.UUID, config *structpb.Struct, logger *zap.Logger) error {
	gRPCCLient, gRPCCLientConn := initModelPublicServiceClient(getModelServerURL(config))
	if gRPCCLientConn != nil {
		defer gRPCCLientConn.Close()
	}
	ctx := metadata.NewOutgoingContext(context.Background(), getRequestMetadata(config))
	_, err := gRPCCLient.ListModels(ctx, &modelPB.ListModelsRequest{})
	if err != nil {
		return err
	}

	return nil
}

func (c *Connector) GetConnectorDefinitionByID(defID string, component *pipelinePB.ConnectorComponent) (*pipelinePB.ConnectorDefinition, error) {
	def, err := c.Connector.GetConnectorDefinitionByID(defID, component)
	if err != nil {
		return nil, err
	}

	return c.GetConnectorDefinitionByUID(uuid.FromStringOrNil(def.Uid), component)
}

type ModelsResp struct {
	Models []struct {
		Name string `json:"name"`
		Task string `json:"task"`
	} `json:"models"`
}

// Generate the model_name enum based on the task
func (c *Connector) GetConnectorDefinitionByUID(defUID uuid.UUID, component *pipelinePB.ConnectorComponent) (*pipelinePB.ConnectorDefinition, error) {
	oriDef, err := c.Connector.GetConnectorDefinitionByUID(defUID, component)
	if err != nil {
		return nil, err
	}
	def := proto.Clone(oriDef).(*pipelinePB.ConnectorDefinition)

	if component.Connection != nil {
		if getModelServerURL(component.Connection) == "" {
			return def, nil
		}

		gRPCCLient, gRPCCLientConn := initModelPublicServiceClient(getModelServerURL(component.Connection))
		if gRPCCLientConn != nil {
			defer gRPCCLientConn.Close()
		}
		ctx := metadata.NewOutgoingContext(context.Background(), getRequestMetadata(component.Connection))
		// We should query by pages and accumulate them in the future

		pageToken := ""
		models := []*modelPB.Model{}
		for {
			resp, err := gRPCCLient.ListModels(ctx, &modelPB.ListModelsRequest{PageToken: &pageToken})
			if err != nil {
				return def, nil
			}
			models = append(models, resp.Models...)
			pageToken = resp.NextPageToken
			if pageToken == "" {
				break
			}
		}

		modelNameMap := map[string]*structpb.ListValue{}

		modelName := &structpb.ListValue{}
		for _, model := range models {
			if _, ok := modelNameMap[model.Task.String()]; !ok {
				modelNameMap[model.Task.String()] = &structpb.ListValue{}
			}
			namePaths := strings.Split(model.Name, "/")
			modelName.Values = append(modelName.Values, structpb.NewStringValue(fmt.Sprintf("%s/%s", namePaths[1], namePaths[3])))
			modelNameMap[model.Task.String()].Values = append(modelNameMap[model.Task.String()].Values, structpb.NewStringValue(fmt.Sprintf("%s/%s", namePaths[1], namePaths[3])))
		}
		for _, sch := range def.Spec.ComponentSpecification.Fields["oneOf"].GetListValue().Values {
			task := sch.GetStructValue().Fields["properties"].GetStructValue().Fields["task"].GetStructValue().Fields["const"].GetStringValue()
			if _, ok := modelNameMap[task]; ok {
				addModelEnum(sch.GetStructValue().Fields, modelNameMap[task])
			}

		}
	}
	return def, nil
}

func addModelEnum(compSpec map[string]*structpb.Value, modelName *structpb.ListValue) {
	if compSpec == nil {
		return
	}
	for key, sch := range compSpec {
		if key == "model_name" {
			sch.GetStructValue().Fields["enum"] = structpb.NewListValue(modelName)
		}

		if sch.GetStructValue() != nil {
			addModelEnum(sch.GetStructValue().Fields, modelName)
		}
		if sch.GetListValue() != nil {
			for _, v := range sch.GetListValue().Values {
				if v.GetStructValue() != nil {
					addModelEnum(v.GetStructValue().Fields, modelName)
				}
			}
		}
	}
}
