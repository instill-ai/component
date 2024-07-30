//go:generate compogen readme ./config ./README.mdx
package openai

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gabriel-vasile/mimetype"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/x/errmsg"
)

const (
	host = "https://api.openai.com"

	TextGenerationTask    = "TASK_TEXT_GENERATION"
	TextEmbeddingsTask    = "TASK_TEXT_EMBEDDINGS"
	SpeechRecognitionTask = "TASK_SPEECH_RECOGNITION"
	TextToSpeechTask      = "TASK_TEXT_TO_SPEECH"
	TextToImageTask       = "TASK_TEXT_TO_IMAGE"

	cfgAPIKey       = "api-key"
	cfgOrganization = "organization"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/setup.json
	setupJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte

	once sync.Once
	comp *component
)

// Connector executes queries against OpenAI.
type component struct {
	base.Component

	instillAPIKey string
}

// Init returns an initialized OpenAI connector.
func Init(bc base.Component) *component {
	once.Do(func() {
		comp = &component{Component: bc}
		err := comp.LoadDefinition(definitionJSON, setupJSON, tasksJSON, nil)
		if err != nil {
			panic(err)
		}
	})

	return comp
}

// WithInstillCredentials loads Instill credentials into the component, which
// can be used to configure it with globally defined parameters instead of with
// user-defined credential values.
func (c *component) WithInstillCredentials(s map[string]any) *component {
	c.instillAPIKey = base.ReadFromGlobalConfig(cfgAPIKey, s)
	return c
}

// CreateExecution initializes a connector executor that can be used in a
// pipeline trigger.
func (c *component) CreateExecution(sysVars map[string]any, setup *structpb.Struct, task string) (base.IExecution, error) {
	resolvedSetup, resolved, err := c.resolveSetup(setup)
	if err != nil {
		return nil, err
	}

	return &execution{
		ComponentExecution: base.ComponentExecution{
			Component:       c,
			SystemVariables: sysVars,
			Setup:           resolvedSetup,
			Task:            task,
		},
		usesInstillCredentials: resolved,
	}, nil
}

// resolveSetup checks whether the component is configured to use the Instill
// credentials injected during initialization and, if so, returns a new setup
// with the secret credential values.
func (c *component) resolveSetup(setup *structpb.Struct) (*structpb.Struct, bool, error) {
	apiKey := setup.GetFields()[cfgAPIKey].GetStringValue()
	if apiKey != base.SecretKeyword {
		return setup, false, nil
	}

	if c.instillAPIKey == "" {
		return nil, false, base.NewUnresolvedCredential(cfgAPIKey)
	}

	setup.GetFields()[cfgAPIKey] = structpb.NewStringValue(c.instillAPIKey)
	return setup, true, nil
}

type execution struct {
	base.ComponentExecution
	usesInstillCredentials bool
}

func (e *execution) UsesInstillCredentials() bool {
	return e.usesInstillCredentials
}

func (e *execution) Execute(_ context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	client := newClient(e.Setup, e.GetLogger())
	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		switch e.Task {
		case TextGenerationTask:
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
						messages = append(messages, multiModalMessage{Role: chat.Role, Content: chat.Content})
					} else {
						content := ""
						for _, c := range chat.Content {
							// OpenAI doesn't support multi-modal content for
							// non-user roles.
							if c.Type == "text" {
								content = *c.Text
							}
						}
						messages = append(messages, message{Role: chat.Role, Content: content})
					}

				}
			} else if inputStruct.SystemMessage != nil {
				// If chat history is not provided, add the system message to the messages
				messages = append(messages, message{Role: "system", Content: *inputStruct.SystemMessage})
			}
			userContents := []content{}
			userContents = append(userContents, content{Type: "text", Text: &inputStruct.Prompt})
			for _, image := range inputStruct.Images {
				b, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(image))
				if err != nil {
					return nil, err
				}
				url := fmt.Sprintf("data:%s;base64,%s", mimetype.Detect(b).String(), base.TrimBase64Mime(image))
				userContents = append(userContents, content{Type: "image_url", ImageURL: &imageURL{URL: url}})
			}
			messages = append(messages, multiModalMessage{Role: "user", Content: userContents})

			body := textCompletionReq{
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

			resp := textCompletionResp{}
			req := client.R().SetResult(&resp).SetBody(body)
			if _, err := req.Post(completionsPath); err != nil {
				return inputs, err
			}

			outputStruct := TextCompletionOutput{
				Texts: []string{},
				Usage: usage(resp.Usage),
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

		case TextEmbeddingsTask:
			inputStruct := TextEmbeddingsInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}

			resp := TextEmbeddingsResp{}

			var reqParams TextEmbeddingsReq
			if inputStruct.Dimensions == 0 {
				reqParams = TextEmbeddingsReq{
					Model: inputStruct.Model,
					Input: []string{inputStruct.Text},
				}
			} else {
				reqParams = TextEmbeddingsReq{
					Model:      inputStruct.Model,
					Input:      []string{inputStruct.Text},
					Dimensions: inputStruct.Dimensions,
				}
			}

			req := client.R().SetBody(reqParams).SetResult(&resp)

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

		case SpeechRecognitionTask:
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

				// Verbosity is passed to extract result duration.
				ResponseFormat: "verbose_json",
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

		case TextToSpeechTask:
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

		case TextToImageTask:

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
func (c *component) Test(_ map[string]any, setup *structpb.Struct) error {
	models := ListModelsResponse{}
	req := newClient(setup, c.GetLogger()).R().SetResult(&models)

	if _, err := req.Get(listModelsPath); err != nil {
		return err
	}

	if len(models.Data) == 0 {
		return fmt.Errorf("no models")
	}

	return nil
}
