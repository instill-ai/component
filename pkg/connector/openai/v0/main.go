//go:generate compogen readme --connector ./config ./README.mdx
package openai

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/gabriel-vasile/mimetype"
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

	cfgAPIKey       = "api_key"
	cfgOrganization = "organization"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	//go:embed config/openai.json
	openAIJSON []byte

	once sync.Once
	con  *Connector
)

// Connector executes queries against OpenAI.
type Connector struct {
	base.BaseConnector

	// Global secrets.
	globalAPIKey string
}

type execution struct {
	base.BaseConnectorExecution
}

// Init returns an initialized OpenAI connector.
func Init(l *zap.Logger, u base.UsageHandler) *Connector {
	once.Do(func() {
		con = &Connector{
			BaseConnector: base.BaseConnector{
				Logger:       l,
				UsageHandler: u,
			},
		}

		err := con.LoadConnectorDefinition(definitionJSON, tasksJSON, map[string][]byte{"openai.json": openAIJSON})
		if err != nil {
			panic(err)
		}
	})

	return con
}

// The connection parameter is defined with snake_case, but the
// environment variable configuration loader replaces underscores by dots,
// so we can't use the parameter key directly.
func readFromSecrets(key string, s map[string]any) string {
	sanitized := strings.ReplaceAll(key, "_", "")
	if v, ok := s[sanitized].(string); ok {
		return v
	}

	return ""
}

// WithGlobalConnection reads the global connection configuration, which can
// be used to execute the connector with globally defined secrets.
func (c *Connector) WithGlobalConnection(s map[string]any) *Connector {
	c.globalAPIKey = readFromSecrets(cfgAPIKey, s)

	return c
}

// CreateExecution initializes a connector executor that can be used in a
// pipeline trigger.
func (c *Connector) CreateExecution(sysVars map[string]any, connection *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	resolvedConnection, err := c.resolveSecrets(connection)
	if err != nil {
		return nil, err
	}

	return &base.ExecutionWrapper{Execution: &execution{
		BaseConnectorExecution: base.BaseConnectorExecution{
			Connector:       c,
			SystemVariables: sysVars,
			Connection:      resolvedConnection,
			Task:            task,
		},
	}}, nil
}

// resolveSecrets looks for references to a global secret in the connection
// and replaces them by the global secret injected during initialization.
func (c *Connector) resolveSecrets(conn *structpb.Struct) (*structpb.Struct, error) {
	apiKey := conn.GetFields()[cfgAPIKey].GetStringValue()
	if apiKey == base.ConnectionGlobalSecret {
		if c.globalAPIKey == "" {
			return nil, base.NewUnresolvedGlobalSecret(cfgAPIKey)
		}

		conn.GetFields()[cfgAPIKey] = structpb.NewStringValue(c.globalAPIKey)
	}

	return conn, nil
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
	return config.GetFields()[cfgAPIKey].GetStringValue()
}

func getOrg(config *structpb.Struct) string {
	val, ok := config.GetFields()[cfgOrganization]
	if !ok {
		return ""
	}
	return val.GetStringValue()
}

func (e *execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	client := newClient(e.Connection, e.GetLogger())
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
func (c *Connector) Test(_ map[string]any, connection *structpb.Struct) error {
	models := ListModelsResponse{}
	req := newClient(connection, c.Logger).R().SetResult(&models)

	if _, err := req.Get(listModelsPath); err != nil {
		return err
	}

	if len(models.Data) == 0 {
		return fmt.Errorf("no models")
	}

	return nil
}
