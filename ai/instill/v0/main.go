//go:generate compogen readme ./config ./README.mdx
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

	"github.com/instill-ai/component/base"

	commonPB "github.com/instill-ai/protogen-go/common/task/v1alpha"
	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
	pb "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

// TODO: The Instill Model component will be refactored soon to align the data
// structure with Instill Model.

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

// CreateExecution initializes a connector executor that can be used in a
// pipeline trigger.
func (c *component) CreateExecution(x base.ComponentExecution) (base.IExecution, error) {
	return &execution{ComponentExecution: x}, nil
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

func getInstillRequesterUID(vars map[string]any) string {
	return vars["__PIPELINE_REQUESTER_UID"].(string)
}

func getModelServerURL(vars map[string]any) string {
	if v, ok := vars["__MODEL_BACKEND"]; ok {
		return v.(string)
	}
	return ""
}

func getRequestMetadata(vars map[string]any) metadata.MD {
	md := metadata.Pairs(
		"Authorization", getHeaderAuthorization(vars),
		"Instill-User-Uid", getInstillUserUID(vars),
		"Instill-Auth-Type", "user",
	)

	if requester := getInstillRequesterUID(vars); requester != "" {
		md.Set("Instill-Requester-Uid", requester)
	}

	return md
}

func (e *execution) Execute(ctx context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	var err error

	if len(inputs) <= 0 || inputs[0] == nil {
		return inputs, fmt.Errorf("invalid input")
	}

	// TODO, we should move this to CreateExecution
	gRPCClient, gRPCCientConn := initModelPublicServiceClient(getModelServerURL(e.SystemVariables))
	if gRPCCientConn != nil {
		defer gRPCCientConn.Close()
	}

	modelNameSplits := strings.Split(inputs[0].GetFields()["model-name"].GetStringValue(), "/")

	nsID := modelNameSplits[0]
	modelID := modelNameSplits[1]
	version := modelNameSplits[2]
	var result []*structpb.Struct
	switch e.Task {
	case commonPB.Task_TASK_UNSPECIFIED.String():
		result, err = e.executeUnspecified(gRPCClient, nsID, modelID, version, inputs)
	case commonPB.Task_TASK_CLASSIFICATION.String():
		result, err = e.executeImageClassification(gRPCClient, nsID, modelID, version, inputs)
	case commonPB.Task_TASK_DETECTION.String():
		result, err = e.executeObjectDetection(gRPCClient, nsID, modelID, version, inputs)
	case commonPB.Task_TASK_KEYPOINT.String():
		result, err = e.executeKeyPointDetection(gRPCClient, nsID, modelID, version, inputs)
	case commonPB.Task_TASK_OCR.String():
		result, err = e.executeOCR(gRPCClient, nsID, modelID, version, inputs)
	case commonPB.Task_TASK_INSTANCE_SEGMENTATION.String():
		result, err = e.executeInstanceSegmentation(gRPCClient, nsID, modelID, version, inputs)
	case commonPB.Task_TASK_SEMANTIC_SEGMENTATION.String():
		result, err = e.executeSemanticSegmentation(gRPCClient, nsID, modelID, version, inputs)
	case commonPB.Task_TASK_TEXT_TO_IMAGE.String():
		result, err = e.executeTextToImage(gRPCClient, nsID, modelID, version, inputs)
	case commonPB.Task_TASK_TEXT_GENERATION.String():
		result, err = e.executeTextGeneration(gRPCClient, nsID, modelID, version, inputs)
	case commonPB.Task_TASK_TEXT_GENERATION_CHAT.String():
		result, err = e.executeTextGenerationChat(gRPCClient, nsID, modelID, version, inputs)
	case commonPB.Task_TASK_VISUAL_QUESTION_ANSWERING.String():
		result, err = e.executeVisualQuestionAnswering(gRPCClient, nsID, modelID, version, inputs)
	case commonPB.Task_TASK_IMAGE_TO_IMAGE.String():
		result, err = e.executeImageToImage(gRPCClient, nsID, modelID, version, inputs)
	default:
		return inputs, fmt.Errorf("unsupported task: %s", e.Task)
	}

	return result, err
}

func (c *component) Test(sysVars map[string]any, setup *structpb.Struct) error {
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
func (c *component) GetDefinition(sysVars map[string]any, compConfig *base.ComponentConfig) (*pb.ComponentDefinition, error) {

	oriDef, err := c.Component.GetDefinition(nil, nil)
	if err != nil {
		return nil, err
	}
	if sysVars == nil && compConfig == nil {
		return oriDef, nil
	}
	def := proto.Clone(oriDef).(*pb.ComponentDefinition)

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
	pageSize := int32(100)
	modelNameMap := map[string]*structpb.ListValue{}
	for {
		resp, err := gRPCCLient.ListModels(ctx, &modelPB.ListModelsRequest{PageToken: &pageToken, PageSize: &pageSize, View: modelPB.View_VIEW_BASIC.Enum()})
		if err != nil {
			return def, nil
		}
		for _, m := range resp.Models {
			if _, ok := modelNameMap[m.Task.String()]; !ok {
				modelNameMap[m.Task.String()] = &structpb.ListValue{}
			}
			namePaths := strings.Split(m.Name, "/")
			for _, v := range m.Versions {
				modelNameMap[m.Task.String()].Values = append(modelNameMap[m.Task.String()].Values, structpb.NewStringValue(fmt.Sprintf("%s/%s/%s", namePaths[1], namePaths[3], v)))
			}
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}

	for _, sch := range def.Spec.ComponentSpecification.Fields["oneOf"].GetListValue().Values {
		task := sch.GetStructValue().Fields["properties"].GetStructValue().Fields["task"].GetStructValue().Fields["const"].GetStringValue()
		if _, ok := modelNameMap[task]; ok {
			addModelEnum(sch.GetStructValue().Fields, modelNameMap[task])
		}

	}

	return def, nil
}

func addModelEnum(compSpec map[string]*structpb.Value, modelName *structpb.ListValue) {
	if compSpec == nil {
		return
	}
	for key, sch := range compSpec {
		if key == "model-name" {
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
