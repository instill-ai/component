package openai

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/x/errmsg"
)

const (
	host                  = "https://api.openai.com"
	textGenerationTask    = "TASK_TEXT_GENERATION"
	textEmbeddingsTask    = "TASK_TEXT_EMBEDDINGS"
	speechRecognitionTask = "TASK_SPEECH_RECOGNITION"
	textToSpeechTask      = "TASK_TEXT_TO_SPEECH"
	textToImageTask       = "TASK_TEXT_TO_IMAGE"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	//go:embed config/openai.json
	openAIJSON []byte

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
		err := connector.LoadConnectorDefinition(definitionJSON, tasksJSON, map[string][]byte{"openai.json": openAIJSON})
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

// getBasePath returns OpenAI's API URL. This configuration param allows us to
// override the API the connector will point to. It isn't meant to be exposed
// to users. Rather, it can serve to test the logic against a fake server.
// TODO instead of having the API value hardcoded in the codebase, it should be
// read from a config file or environment variable.
func getBasePath(config *structpb.Struct) string {
	v, ok := config.GetFields()["base_path"]
	if !ok {
		return host
	}
	return v.GetStringValue()
}

func getAPIKey(config *structpb.Struct) string {
	return config.GetFields()["api_key"].GetStringValue()
}

func getOrg(config *structpb.Struct) string {
	val, ok := config.GetFields()["organization"]
	if !ok {
		return ""
	}
	return val.GetStringValue()
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	client := newClient(e.Config, e.Logger)
	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		switch e.Task {
		case textGenerationTask:
			inputStruct := TextCompletionInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}

			messages := []interface{}{}

			// If chat history is provided, add it to the messages, and ignore the system message
			if inputStruct.ChatHistory != nil {
				for _, chat := range inputStruct.ChatHistory {
					if chat.Role == "user" {
						messages = append(messages, MultiModalMessage{Role: chat.Role, Content: chat.Content})
					} else {
						content := ""
						for _, c := range chat.Content {
							// OpenAI doesn't support MultiModal Content for non-user role
							if c.Type == "text" {
								content = *c.Text
							}
						}
						messages = append(messages, Message{Role: chat.Role, Content: content})
					}

				}
			} else {
				// If chat history is not provided, add the system message to the messages
				if inputStruct.SystemMessage != nil {
					messages = append(messages, Message{Role: "system", Content: *inputStruct.SystemMessage})
				}
			}
			userContents := []Content{}
			userContents = append(userContents, Content{Type: "text", Text: &inputStruct.Prompt})
			for _, image := range inputStruct.Images {
				b, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(image))
				if err != nil {
					return nil, err
				}
				url := fmt.Sprintf("data:%s;base64,%s", mimetype.Detect(b).String(), base.TrimBase64Mime(image))
				userContents = append(userContents, Content{Type: "image_url", ImageURL: &ImageURL{URL: url}})
			}
			messages = append(messages, MultiModalMessage{Role: "user", Content: userContents})

			body := TextCompletionReq{
				Messages:         messages,
				Model:            inputStruct.Model,
				MaxTokens:        inputStruct.MaxTokens,
				Temperature:      inputStruct.Temperature,
				N:                inputStruct.N,
				TopP:             inputStruct.TopP,
				PresencePenalty:  inputStruct.PresencePenalty,
				FrequencyPenalty: inputStruct.FrequencyPenalty,
			}

			// workaround, the OpenAI service can not accept this param
			if inputStruct.Model != "gpt-4-vision-preview" {
				body.ResponseFormat = inputStruct.ResponseFormat
			}

			resp := TextCompletionResp{}
			req := client.R().SetResult(&resp).SetBody(body)
			if _, err := req.Post(completionsPath); err != nil {
				return inputs, err
			}

			outputStruct := TextCompletionOutput{
				Texts: []string{},
			}
			for _, c := range resp.Choices {
				outputStruct.Texts = append(outputStruct.Texts, c.Message.Content)
			}

			outputJSON, err := json.Marshal(outputStruct)
			if err != nil {
				return nil, err
			}
			output := structpb.Struct{}
			err = protojson.Unmarshal(outputJSON, &output)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, &output)

		case textEmbeddingsTask:
			inputStruct := TextEmbeddingsInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}

			resp := TextEmbeddingsResp{}
			req := client.R().SetBody(TextEmbeddingsReq{
				Model: inputStruct.Model,
				Input: []string{inputStruct.Text},
			}).SetResult(&resp)

			if _, err := req.Post(embeddingsPath); err != nil {
				return inputs, err
			}

			outputStruct := TextEmbeddingsOutput{
				Embedding: resp.Data[0].Embedding,
			}

			output, err := base.ConvertToStructpb(outputStruct)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, output)

		case speechRecognitionTask:
			inputStruct := AudioTranscriptionInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}

			audioBytes, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(inputStruct.Audio))
			if err != nil {
				return nil, err
			}

			data, ct, err := getBytes(AudioTranscriptionReq{
				File:        audioBytes,
				Model:       inputStruct.Model,
				Prompt:      inputStruct.Prompt,
				Language:    inputStruct.Prompt,
				Temperature: inputStruct.Temperature,
			})
			if err != nil {
				return inputs, err
			}

			resp := AudioTranscriptionResp{}
			req := client.R().SetBody(data).SetResult(&resp).SetHeader("Content-Type", ct)
			if _, err := req.Post(transcriptionsPath); err != nil {
				return inputs, err
			}

			output, err := base.ConvertToStructpb(resp)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, output)

		case textToSpeechTask:
			inputStruct := TextToSpeechInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}

			req := client.R().SetBody(TextToSpeechReq{
				Input:          inputStruct.Text,
				Model:          inputStruct.Model,
				Voice:          inputStruct.Voice,
				ResponseFormat: inputStruct.ResponseFormat,
				Speed:          inputStruct.Speed,
			})

			resp, err := req.Post(createSpeechPath)
			if err != nil {
				return inputs, err
			}

			audio := base64.StdEncoding.EncodeToString(resp.Body())
			outputStruct := TextToSpeechOutput{
				Audio: fmt.Sprintf("data:audio/wav;base64,%s", audio),
			}

			output, err := base.ConvertToStructpb(outputStruct)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, output)

		case textToImageTask:

			inputStruct := ImagesGenerationInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}

			resp := ImageGenerationsResp{}
			req := client.R().SetBody(ImageGenerationsReq{
				Model:          inputStruct.Model,
				Prompt:         inputStruct.Prompt,
				Quality:        inputStruct.Quality,
				Size:           inputStruct.Size,
				Style:          inputStruct.Style,
				N:              inputStruct.N,
				ResponseFormat: "b64_json",
			}).SetResult(&resp)

			if _, err := req.Post(imgGenerationPath); err != nil {
				return inputs, err
			}

			results := []ImageGenerationsOutputResult{}
			for _, data := range resp.Data {
				results = append(results, ImageGenerationsOutputResult{
					Image:         fmt.Sprintf("data:image/webp;base64,%s", data.Image),
					RevisedPrompt: data.RevisedPrompt,
				})
			}
			outputStruct := ImageGenerationsOutput{
				Results: results,
			}

			output, err := base.ConvertToStructpb(outputStruct)
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

// Test checks the connector state.
func (c *Connector) Test(_ uuid.UUID, config *structpb.Struct, logger *zap.Logger) error {
	models := ListModelsResponse{}
	req := newClient(config, logger).R().SetResult(&models)

	if _, err := req.Get(listModelsPath); err != nil {
		return err
	}

	if len(models.Data) == 0 {
		return fmt.Errorf("no models")
	}

	return nil
}
