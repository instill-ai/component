package huggingface

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/x/errmsg"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

const (
	textGenerationTask         = "TASK_TEXT_GENERATION"
	textToImageTask            = "TASK_TEXT_TO_IMAGE"
	fillMaskTask               = "TASK_FILL_MASK"
	summarizationTask          = "TASK_SUMMARIZATION"
	textClassificationTask     = "TASK_TEXT_CLASSIFICATION"
	tokenClassificationTask    = "TASK_TOKEN_CLASSIFICATION"
	translationTask            = "TASK_TRANSLATION"
	zeroShotClassificationTask = "TASK_ZERO_SHOT_CLASSIFICATION"
	featureExtractionTask      = "TASK_FEATURE_EXTRACTION"
	questionAnsweringTask      = "TASK_QUESTION_ANSWERING"
	tableQuestionAnsweringTask = "TASK_TABLE_QUESTION_ANSWERING"
	sentenceSimilarityTask     = "TASK_SENTENCE_SIMILARITY"
	conversationalTask         = "TASK_CONVERSATIONAL"
	imageClassificationTask    = "TASK_IMAGE_CLASSIFICATION"
	imageSegmentationTask      = "TASK_IMAGE_SEGMENTATION"
	objectDetectionTask        = "TASK_OBJECT_DETECTION"
	imageToTextTask            = "TASK_IMAGE_TO_TEXT"
	speechRecognitionTask      = "TASK_SPEECH_RECOGNITION"
	audioClassificationTask    = "TASK_AUDIO_CLASSIFICATION"
)

var (
	//go:embed config/definitions.json
	definitionsJSON []byte
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

func Init(logger *zap.Logger) base.IConnector {
	once.Do(func() {
		connector = &Connector{
			Connector: base.Connector{
				Component: base.Component{Logger: logger},
			},
		}
		err := connector.LoadConnectorDefinitions(definitionsJSON, tasksJSON, nil)
		if err != nil {
			logger.Fatal(err.Error())
		}
	})
	return connector
}

func (c *Connector) CreateExecution(defUID uuid.UUID, task string, config *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	e := &Execution{}
	e.Execution = base.CreateExecutionHelper(e, c, defUID, task, config, logger)
	return e, nil
}

func getAPIKey(config *structpb.Struct) string {
	return config.GetFields()["api_key"].GetStringValue()
}

func getBaseURL(config *structpb.Struct) string {
	return config.GetFields()["base_url"].GetStringValue()
}

func isCustomEndpoint(config *structpb.Struct) bool {
	return config.GetFields()["is_custom_endpoint"].GetBoolValue()
}

func wrapSliceInStruct(data []byte, key string) (*structpb.Struct, error) {
	var list []any
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}

	results, err := structpb.NewList(list)
	if err != nil {
		return nil, err
	}

	return &structpb.Struct{
		Fields: map[string]*structpb.Value{
			key: structpb.NewListValue(results),
		},
	}, nil
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	client := newClient(e.Config, e.Logger)
	outputs := []*structpb.Struct{}

	path := "/"
	if !isCustomEndpoint(e.Config) {
		path = modelsPath + inputs[0].GetFields()["model"].GetStringValue()
	}

	for _, input := range inputs {
		switch e.Task {
		case textGenerationTask:
			inputStruct := TextGenerationRequest{}
			if err := base.ConvertFromStructpb(input, &inputStruct); err != nil {
				return nil, err
			}

			resp := []TextGenerationResponse{}
			req := client.R().SetBody(inputStruct).SetResult(&resp)
			if _, err := post(req, path); err != nil {
				return nil, err
			}

			if len(resp) < 1 {
				err := fmt.Errorf("invalid response")
				return nil, errmsg.AddMessage(err, "Hugging Face didn't return any result")
			}

			output, err := structpb.NewStruct(map[string]any{"generated_text": resp[0].GeneratedText})
			if err != nil {
				return nil, err
			}

			outputs = append(outputs, output)
		case textToImageTask:
			inputStruct := TextToImageRequest{}
			if err := base.ConvertFromStructpb(input, &inputStruct); err != nil {
				return nil, err
			}

			req := client.R().SetBody(inputStruct)
			resp, err := post(req, path)
			if err != nil {
				return nil, err
			}

			rawImg := base64.StdEncoding.EncodeToString(resp.Body())
			output, err := structpb.NewStruct(map[string]any{
				"image": fmt.Sprintf("data:image/jpeg;base64,%s", rawImg),
			})
			if err != nil {
				return nil, err
			}

			outputs = append(outputs, output)
		case fillMaskTask:
			inputStruct := FillMaskRequest{}
			if err := base.ConvertFromStructpb(input, &inputStruct); err != nil {
				return nil, err
			}

			req := client.R().SetBody(inputStruct)
			resp, err := post(req, path)
			if err != nil {
				return nil, err
			}

			output, err := wrapSliceInStruct(resp.Body(), "results")
			if err != nil {
				return nil, err
			}

			outputs = append(outputs, output)
		case summarizationTask:
			inputStruct := SummarizationRequest{}
			if err := base.ConvertFromStructpb(input, &inputStruct); err != nil {
				return nil, err
			}

			resp := []SummarizationResponse{}
			req := client.R().SetBody(inputStruct).SetResult(&resp)
			if _, err := post(req, path); err != nil {
				return nil, err
			}

			if len(resp) < 1 {
				err := fmt.Errorf("invalid response")
				return nil, errmsg.AddMessage(err, "Hugging Face didn't return any result")
			}

			output, err := structpb.NewStruct(map[string]any{"summary_text": resp[0].SummaryText})
			if err != nil {
				return nil, err
			}

			outputs = append(outputs, output)
		case textClassificationTask:
			inputStruct := TextClassificationRequest{}
			if err := base.ConvertFromStructpb(input, &inputStruct); err != nil {
				return nil, err
			}

			var resp [][]any
			req := client.R().SetBody(inputStruct).SetResult(&resp)
			if _, err := post(req, path); err != nil {
				return nil, err
			}

			if len(resp) < 1 {
				err := fmt.Errorf("invalid response")
				return nil, errmsg.AddMessage(err, "Hugging Face didn't return any result")
			}

			results, err := structpb.NewList(resp[0])
			if err != nil {
				return nil, err
			}

			output := &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"results": structpb.NewListValue(results),
				},
			}

			outputs = append(outputs, output)
		case tokenClassificationTask:
			inputStruct := TokenClassificationRequest{}
			if err := base.ConvertFromStructpb(input, &inputStruct); err != nil {
				return nil, err
			}
			req := client.R().SetBody(inputStruct)
			resp, err := post(req, path)
			if err != nil {
				return nil, err
			}

			output, err := wrapSliceInStruct(resp.Body(), "results")
			if err != nil {
				return nil, err
			}

			outputs = append(outputs, output)
		case translationTask:
			inputStruct := TranslationRequest{}
			if err := base.ConvertFromStructpb(input, &inputStruct); err != nil {
				return nil, err
			}

			resp := []TranslationResponse{}
			req := client.R().SetBody(inputStruct).SetResult(&resp)
			if _, err := post(req, path); err != nil {
				return nil, err
			}

			if len(resp) < 1 {
				err := fmt.Errorf("invalid response")
				return nil, errmsg.AddMessage(err, "Hugging Face didn't return any result")
			}

			output, err := structpb.NewStruct(map[string]any{"translation_text": resp[0].TranslationText})
			if err != nil {
				return nil, err
			}

			outputs = append(outputs, output)
		case zeroShotClassificationTask:
			inputStruct := ZeroShotRequest{}
			if err := base.ConvertFromStructpb(input, &inputStruct); err != nil {
				return nil, err
			}

			req := client.R().SetBody(inputStruct)
			resp, err := post(req, path)
			if err != nil {
				return nil, err
			}

			var output structpb.Struct
			if err = protojson.Unmarshal(resp.Body(), &output); err != nil {
				return nil, err
			}

			outputs = append(outputs, &output)
		// case featureExtractionTask:
		// TODO: fix this task
		// 	inputStruct := FeatureExtractionRequest{}
		// 	err := base.ConvertFromStructpb(input, &inputStruct)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	jsonBody, _ := json.Marshal(inputStruct)
		// 	resp, err := doer.MakeHFAPIRequest(jsonBody, model)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	threeDArr := [][][]float64{}
		// 	err = json.Unmarshal(resp, &threeDArr)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	if len(threeDArr) <= 0 {
		// 		return nil, errors.New("invalid response")
		// 	}
		// 	nestedArr := threeDArr[0]
		// 	features := structpb.ListValue{}
		// 	features.Values = make([]*structpb.Value, len(nestedArr))
		// 	for i, innerArr := range nestedArr {
		// 		innerValues := make([]*structpb.Value, len(innerArr))
		// 		for j := range innerArr {
		// 			innerValues[j] = &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: innerArr[j]}}
		// 		}
		// 		features.Values[i] = &structpb.Value{Kind: &structpb.Value_ListValue{ListValue: &structpb.ListValue{Values: innerValues}}}
		// 	}
		// 	output := structpb.Struct{
		// 		Fields: map[string]*structpb.Value{"feature": {Kind: &structpb.Value_ListValue{ListValue: &features}}},
		// 	}
		// 	outputs = append(outputs, &output)
		case questionAnsweringTask:
			inputStruct := QuestionAnsweringRequest{}
			if err := base.ConvertFromStructpb(input, &inputStruct); err != nil {
				return nil, err
			}
			req := client.R().SetBody(inputStruct)
			resp, err := post(req, path)
			if err != nil {
				return nil, err
			}

			var output structpb.Struct
			if err = protojson.Unmarshal(resp.Body(), &output); err != nil {
				return nil, err
			}

			outputs = append(outputs, &output)
		case tableQuestionAnsweringTask:
			inputStruct := TableQuestionAnsweringRequest{}
			if err := base.ConvertFromStructpb(input, &inputStruct); err != nil {
				return nil, err
			}

			req := client.R().SetBody(inputStruct)
			resp, err := post(req, path)
			if err != nil {
				return nil, err
			}

			var output structpb.Struct
			if err = protojson.Unmarshal(resp.Body(), &output); err != nil {
				return nil, err
			}

			outputs = append(outputs, &output)
		case sentenceSimilarityTask:
			inputStruct := SentenceSimilarityRequest{}
			if err := base.ConvertFromStructpb(input, &inputStruct); err != nil {
				return nil, err
			}

			req := client.R().SetBody(inputStruct)
			resp, err := post(req, path)
			if err != nil {
				return nil, err
			}

			output, err := wrapSliceInStruct(resp.Body(), "scores")
			if err != nil {
				return nil, err
			}

			outputs = append(outputs, output)
		case conversationalTask:
			inputStruct := ConversationalRequest{}
			if err := base.ConvertFromStructpb(input, &inputStruct); err != nil {
				return nil, err
			}

			req := client.R().SetBody(inputStruct)
			resp, err := post(req, path)
			if err != nil {
				return nil, err
			}

			var output structpb.Struct
			if err = protojson.Unmarshal(resp.Body(), &output); err != nil {
				return nil, err
			}

			outputs = append(outputs, &output)
		case imageClassificationTask:
			inputStruct := ImageRequest{}
			if err := base.ConvertFromStructpb(input, &inputStruct); err != nil {
				return nil, err
			}

			b, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(inputStruct.Image))
			if err != nil {
				return nil, err
			}

			req := client.R().SetBody(b)
			resp, err := post(req, path)
			if err != nil {
				return nil, err
			}

			output, err := wrapSliceInStruct(resp.Body(), "classes")
			if err != nil {
				return nil, err
			}

			outputs = append(outputs, output)
		case imageSegmentationTask:
			inputStruct := ImageRequest{}
			if err := base.ConvertFromStructpb(input, &inputStruct); err != nil {
				return nil, err
			}

			b, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(inputStruct.Image))
			if err != nil {
				return nil, err
			}

			resp := []ImageSegmentationResponse{}
			req := client.R().SetBody(b).SetResult(&resp)
			if _, err := post(req, path); err != nil {
				return nil, err
			}

			segments := &structpb.ListValue{
				Values: make([]*structpb.Value, len(resp)),
			}

			for i := range resp {
				segment, err := structpb.NewStruct(map[string]any{
					"score": resp[i].Score,
					"label": resp[i].Label,
					"mask":  fmt.Sprintf("data:image/png;base64,%s", resp[i].Mask),
				})

				if err != nil {
					return nil, err
				}

				segments.Values[i] = structpb.NewStructValue(segment)
			}

			output := &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"segments": structpb.NewListValue(segments),
				},
			}

			outputs = append(outputs, output)
		case objectDetectionTask:
			inputStruct := ImageRequest{}
			if err := base.ConvertFromStructpb(input, &inputStruct); err != nil {
				return nil, err
			}

			b, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(inputStruct.Image))
			if err != nil {
				return nil, err
			}

			req := client.R().SetBody(b)
			resp, err := post(req, path)
			if err != nil {
				return nil, err
			}

			output, err := wrapSliceInStruct(resp.Body(), "objects")
			if err != nil {
				return nil, err
			}

			outputs = append(outputs, output)
		case imageToTextTask:
			inputStruct := ImageRequest{}
			if err := base.ConvertFromStructpb(input, &inputStruct); err != nil {
				return nil, err
			}

			b, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(inputStruct.Image))
			if err != nil {
				return nil, err
			}

			resp := []ImageToTextResponse{}
			req := client.R().SetBody(b).SetResult(&resp)
			if _, err := post(req, path); err != nil {
				return nil, err
			}

			if len(resp) < 1 {
				err := fmt.Errorf("invalid response")
				return nil, errmsg.AddMessage(err, "Hugging Face didn't return any result")
			}

			output, err := structpb.NewStruct(map[string]any{"text": resp[0].GeneratedText})
			if err != nil {
				return nil, err
			}

			outputs = append(outputs, output)
		case speechRecognitionTask:
			inputStruct := AudioRequest{}
			if err := base.ConvertFromStructpb(input, &inputStruct); err != nil {
				return nil, err
			}

			b, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(inputStruct.Audio))
			if err != nil {
				return nil, err
			}

			req := client.R().SetBody(b)
			resp, err := post(req, path)
			if err != nil {
				return nil, err
			}

			output := new(structpb.Struct)
			if err := protojson.Unmarshal(resp.Body(), output); err != nil {
				return nil, err
			}

			outputs = append(outputs, output)
		case audioClassificationTask:
			inputStruct := AudioRequest{}
			if err := base.ConvertFromStructpb(input, &inputStruct); err != nil {
				return nil, err
			}

			b, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(inputStruct.Audio))
			if err != nil {
				return nil, err
			}

			req := client.R().SetBody(b)
			resp, err := post(req, path)
			if err != nil {
				return nil, err
			}

			output, err := wrapSliceInStruct(resp.Body(), "classes")
			if err != nil {
				return nil, err
			}

			outputs = append(outputs, output)
		default:
			return nil, errmsg.AddMessage(
				fmt.Errorf("not supported task: %s", e.Task),
				fmt.Sprintf("%s task is not supported.", e.Task),
			)
		}
	}

	return outputs, nil
}

func (c *Connector) Test(_ uuid.UUID, config *structpb.Struct, logger *zap.Logger) (pipelinePB.Connector_State, error) {
	req := newClient(config, logger).R()
	resp, err := req.Get("")
	if err != nil {
		return pipelinePB.Connector_STATE_ERROR, err
	}

	if resp.IsError() {
		return pipelinePB.Connector_STATE_DISCONNECTED, nil
	}

	return pipelinePB.Connector_STATE_CONNECTED, nil
}
