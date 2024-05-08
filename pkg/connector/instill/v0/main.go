//go:generate compogen readme --connector ./config ./README.mdx
package instill

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"

	commonPB "github.com/instill-ai/protogen-go/common/task/v1alpha"
	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	once      sync.Once
	con       *connector
)

type connector struct {
	base.BaseConnector

	// Workaround solution
	cacheDefinition *pipelinePB.ConnectorDefinition
}

type execution struct {
	base.BaseConnectorExecution
}

func Init(bc base.BaseConnector) *connector {
	once.Do(func() {
		con = &connector{BaseConnector: bc}
		err := con.LoadConnectorDefinition(definitionJSON, tasksJSON, nil)
		if err != nil {
			panic(err)
		}
	})
	return con
}

func (c *connector) CreateExecution(sysVars map[string]any, connection *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	return &base.ExecutionWrapper{Execution: &execution{
		BaseConnectorExecution: base.BaseConnectorExecution{Connector: c, SystemVariables: sysVars, Connection: connection, Task: task},
	}}, nil
}

func getHeaderAuthorization(vars map[string]any) string {
	if v, ok := vars["__PIPELINE_HEADER_AUTHORIZATION"]; ok {
		return v.(string)
	}
	return ""
}
func getInstillUserUID(vars map[string]any) string {
	return vars["__PIPELINE_USER_UID"].(string)
}

func getModelServerURL(vars map[string]any) string {
	if v, ok := vars["__MODEL_BACKEND"]; ok {
		return v.(string)
	}
	return ""
}

func getMgmtServerURL(vars map[string]any) string {
	if v, ok := vars["__MGMT_BACKEND"]; ok {
		return v.(string)
	}
	return ""
}

// This is a workaround solution for caching the definition in memory if the model list is static.
func useStaticModelList(vars map[string]any) bool {
	if v, ok := vars["__STATIC_MODEL_LIST"]; ok {
		return v.(bool)
	}
	return false
}

func getRequestMetadata(vars map[string]any) metadata.MD {
	return metadata.Pairs(
		"Authorization", getHeaderAuthorization(vars),
		"Instill-User-Uid", getInstillUserUID(vars),
		"Instill-Auth-Type", "user",
	)
}

func (e *execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	var err error

	if len(inputs) <= 0 || inputs[0] == nil {
		return inputs, fmt.Errorf("invalid input")
	}

	gRPCCLient, gRPCCLientConn := initModelPublicServiceClient(getModelServerURL(e.SystemVariables))
	if gRPCCLientConn != nil {
		defer gRPCCLientConn.Close()
	}

	mgmtGRPCCLient, mgmtGRPCCLientConn := initMgmtPublicServiceClient(getMgmtServerURL(e.SystemVariables))
	if mgmtGRPCCLientConn != nil {
		defer mgmtGRPCCLientConn.Close()
	}

	modelNameSplits := strings.Split(inputs[0].GetFields()["model_name"].GetStringValue(), "/")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	ctx = metadata.NewOutgoingContext(ctx, getRequestMetadata(e.SystemVariables))
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

func (c *connector) Test(sysVars map[string]any, connection *structpb.Struct) error {
	gRPCCLient, gRPCCLientConn := initModelPublicServiceClient(getModelServerURL(sysVars))
	if gRPCCLientConn != nil {
		defer gRPCCLientConn.Close()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ctx = metadata.NewOutgoingContext(ctx, getRequestMetadata(sysVars))
	_, err := gRPCCLient.ListModels(ctx, &modelPB.ListModelsRequest{})
	if err != nil {
		return err
	}

	return nil
}

type ModelsResp struct {
	Models []struct {
		Name string `json:"name"`
		Task string `json:"task"`
	} `json:"models"`
}

// Generate the `model_name` enum based on the task.
// This implementation is a temporary solution due to the incomplete feature set of Instill Model.
// We'll re-implement this after Instill Model is stable.
func (c *connector) GetConnectorDefinition(sysVars map[string]any, component *pipelinePB.ConnectorComponent) (*pipelinePB.ConnectorDefinition, error) {
	if useStaticModelList(sysVars) && c.cacheDefinition != nil {
		return c.cacheDefinition, nil
	}

	oriDef, err := c.BaseConnector.GetConnectorDefinition(nil, nil)
	if err != nil {
		return nil, err
	}
	def := proto.Clone(oriDef).(*pipelinePB.ConnectorDefinition)

	if getModelServerURL(sysVars) == "" {
		return def, nil
	}

	gRPCCLient, gRPCCLientConn := initModelPublicServiceClient(getModelServerURL(sysVars))
	if gRPCCLientConn != nil {
		defer gRPCCLientConn.Close()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ctx = metadata.NewOutgoingContext(ctx, getRequestMetadata(sysVars))

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
	if useStaticModelList(sysVars) {
		c.cacheDefinition = def
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
