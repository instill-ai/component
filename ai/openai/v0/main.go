//go:generate compogen readme ./config ./README.mdx
package openai

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/gabriel-vasile/mimetype"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/x/errmsg"

	openaisdk "github.com/sashabaranov/go-openai"
)

const (
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
func (c *component) CreateExecution(x base.ComponentExecution) (base.IExecution, error) {
	resolvedSetup, resolved, err := c.resolveSetup(x.Setup)
	if err != nil {
		return nil, err
	}

	x.Setup = resolvedSetup

	return &execution{
		ComponentExecution:     x,
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

func (e *execution) Execute(ctx context.Context, in base.InputReader, out base.OutputWriter) error {
	inputs, err := in.Read(ctx)
	if err != nil {
		return err
	}

	client := newClient(e.Setup, e.GetLogger())
	outputs := make([]*structpb.Struct, len(inputs))

	for batchIdx, input := range inputs {
		switch e.Task {
		case TextGenerationTask:
			inputStruct := TextCompletionInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return err
			}

			messages := []openaisdk.ChatCompletionMessage{}

			// If chat history is provided, add it to the messages, and ignore the system message
			if inputStruct.ChatHistory != nil {
				for _, chat := range inputStruct.ChatHistory {
					if chat.Role == "user" {
						multiContent := []openaisdk.ChatMessagePart{}
						for _, c := range chat.Content {
							multiContent = append(multiContent, openaisdk.ChatMessagePart{
								Type: openaisdk.ChatMessagePartType(c.Type),
								Text: c.Text,
								ImageURL: &openaisdk.ChatMessageImageURL{
									URL: c.ImageURL.URL,
								},
							})
						}

						messages = append(messages, openaisdk.ChatCompletionMessage{Role: chat.Role, MultiContent: multiContent})
					} else {
						content := ""
						for _, c := range chat.Content {
							// OpenAI doesn't support multi-modal content for
							// non-user roles.
							if c.Type == "text" {
								content = c.Text
							}
						}
						messages = append(messages, openaisdk.ChatCompletionMessage{Role: chat.Role, Content: content})
					}

				}
			} else if inputStruct.SystemMessage != "" {
				// If chat history is not provided, add the system message to the messages
				messages = append(messages, openaisdk.ChatCompletionMessage{Role: "system", Content: inputStruct.SystemMessage})
			}
			userContents := []openaisdk.ChatMessagePart{}
			userContents = append(userContents, openaisdk.ChatMessagePart{Type: "text", Text: inputStruct.Prompt})
			for _, image := range inputStruct.Images {
				b, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(image))
				if err != nil {
					return err
				}
				url := fmt.Sprintf("data:%s;base64,%s", mimetype.Detect(b).String(), base.TrimBase64Mime(image))
				userContents = append(userContents, openaisdk.ChatMessagePart{Type: "image_url", ImageURL: &openaisdk.ChatMessageImageURL{
					URL: url,
				}})
			}
			messages = append(messages, openaisdk.ChatCompletionMessage{Role: "user", MultiContent: userContents})

			body := openaisdk.ChatCompletionRequest{
				Messages:         messages,
				Model:            inputStruct.Model,
				MaxTokens:        inputStruct.MaxTokens,
				Temperature:      inputStruct.Temperature,
				N:                inputStruct.N,
				TopP:             inputStruct.TopP,
				PresencePenalty:  inputStruct.PresencePenalty,
				FrequencyPenalty: inputStruct.FrequencyPenalty,
				Stream:           true,
				StreamOptions:    &openaisdk.StreamOptions{IncludeUsage: true},
			}

			// workaround, the OpenAI service can not accept this param
			if inputStruct.Model != "gpt-4-vision-preview" {
				if inputStruct.ResponseFormat != nil {
					body.ResponseFormat = &openaisdk.ChatCompletionResponseFormat{
						Type: openaisdk.ChatCompletionResponseFormatType(inputStruct.ResponseFormat.Type),
					}
					if inputStruct.ResponseFormat.Type == "json_schema" {
						if inputStruct.Model == "gpt-4o-mini" || inputStruct.Model == "gpt-4o-2024-08-06" {
							sch := map[string]any{}
							if inputStruct.ResponseFormat.JSONSchema != "" {
								err = json.Unmarshal([]byte(inputStruct.ResponseFormat.JSONSchema), &sch)
								if err != nil {
									return err
								}
								openaiSchema := &openaisdk.ChatCompletionResponseFormatJSONSchema{}
								b, _ := json.Marshal(sch)
								_ = json.Unmarshal(b, openaiSchema)
								body.ResponseFormat = &openaisdk.ChatCompletionResponseFormat{
									Type:       openaisdk.ChatCompletionResponseFormatType(inputStruct.ResponseFormat.Type),
									JSONSchema: openaiSchema,
								}
							}

						} else {
							return fmt.Errorf("this model doesn't support response format: json_schema")
						}

					}
				}

			}

			stream, err := client.CreateChatCompletionStream(ctx, body)
			if err != nil {
				return err
			}
			outputStruct := TextCompletionOutput{}

			count := 0
			for {
				response, err := stream.Recv()

				if count == 5 || errors.Is(err, io.EOF) {
					outputJSON, inErr := json.Marshal(outputStruct)
					if inErr != nil {
						return inErr
					}
					output := &structpb.Struct{}
					inErr = protojson.Unmarshal(outputJSON, output)
					if inErr != nil {
						return inErr
					}
					outputs[batchIdx] = output
					inErr = out.Write(ctx, outputs)
					if inErr != nil {
						return inErr
					}
					count = 0
					if errors.Is(err, io.EOF) {
						break
					}
				}

				if err != nil {
					return err
				}
				if outputStruct.Texts == nil {
					outputStruct.Texts = make([]string, len(response.Choices))
				}
				for idx, c := range response.Choices {
					outputStruct.Texts[idx] += c.Delta.Content

				}
				if response.Usage != nil {
					outputStruct.Usage = usage{
						PromptTokens:     response.Usage.PromptTokens,
						CompletionTokens: response.Usage.CompletionTokens,
						TotalTokens:      response.Usage.TotalTokens,
					}
				}

				count += 1

			}
			fmt.Println()
			fmt.Println("xxxxx")

		case TextEmbeddingsTask:
			inputStruct := TextEmbeddingsInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return err
			}

			var reqParams openaisdk.EmbeddingRequest
			if inputStruct.Dimensions == 0 {
				reqParams = openaisdk.EmbeddingRequest{
					Model: openaisdk.EmbeddingModel(inputStruct.Model),
					Input: []string{inputStruct.Text},
				}
			} else {
				reqParams = openaisdk.EmbeddingRequest{
					Model:      openaisdk.EmbeddingModel(inputStruct.Model),
					Input:      []string{inputStruct.Text},
					Dimensions: inputStruct.Dimensions,
				}
			}

			resp, err := client.CreateEmbeddings(ctx, reqParams)
			if err != nil {
				return err
			}

			outputStruct := TextEmbeddingsOutput{
				Embedding: resp.Data[0].Embedding,
			}

			output, err := base.ConvertToStructpb(outputStruct)
			if err != nil {
				return err
			}
			outputs[batchIdx] = output

		case SpeechRecognitionTask:
			inputStruct := AudioTranscriptionInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return err
			}

			audioBytes, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(inputStruct.Audio))
			if err != nil {
				return err
			}

			request := openaisdk.AudioRequest{
				Reader:      bytes.NewReader(audioBytes),
				Model:       inputStruct.Model,
				Prompt:      inputStruct.Prompt,
				Language:    inputStruct.Prompt,
				Temperature: inputStruct.Temperature,
				Format:      openaisdk.AudioResponseFormatVerboseJSON,
			}
			resp, err := client.CreateTranscription(ctx, request)
			if err != nil {
				return err
			}

			output, err := base.ConvertToStructpb(resp)
			if err != nil {
				return err
			}
			outputs[batchIdx] = output

		case TextToSpeechTask:
			inputStruct := TextToSpeechInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return err
			}

			request := openaisdk.CreateSpeechRequest{
				Model:          openaisdk.SpeechModel(inputStruct.Model),
				Input:          inputStruct.Text,
				Voice:          openaisdk.SpeechVoice(inputStruct.Voice),
				ResponseFormat: openaisdk.SpeechResponseFormat(inputStruct.ResponseFormat),
				Speed:          inputStruct.Speed,
			}
			resp, err := client.CreateSpeech(ctx, request)
			if err != nil {
				return err
			}

			buf, err := io.ReadAll(resp)
			if err != nil {
				return err
			}

			audio := base64.StdEncoding.EncodeToString(buf)
			outputStruct := TextToSpeechOutput{
				Audio: fmt.Sprintf("data:audio/wav;base64,%s", audio),
			}

			output, err := base.ConvertToStructpb(outputStruct)
			if err != nil {
				return err
			}
			outputs[batchIdx] = output

		case TextToImageTask:

			inputStruct := ImagesGenerationInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return err
			}

			resp, err := client.CreateImage(ctx, openaisdk.ImageRequest{
				Model:          inputStruct.Model,
				Prompt:         inputStruct.Prompt,
				Quality:        inputStruct.Quality,
				Size:           inputStruct.Size,
				Style:          inputStruct.Style,
				N:              inputStruct.N,
				ResponseFormat: "b64_json",
			})
			if err != nil {
				return err
			}

			results := []ImageGenerationsOutputResult{}
			for _, data := range resp.Data {
				results = append(results, ImageGenerationsOutputResult{
					Image:         fmt.Sprintf("data:image/webp;base64,%s", data.B64JSON),
					RevisedPrompt: data.RevisedPrompt,
				})
			}
			outputStruct := ImageGenerationsOutput{
				Results: results,
			}

			output, err := base.ConvertToStructpb(outputStruct)
			if err != nil {
				return err
			}
			outputs[batchIdx] = output

		default:
			return errmsg.AddMessage(
				fmt.Errorf("not supported task: %s", e.Task),
				fmt.Sprintf("%s task is not supported.", e.Task),
			)
		}
	}

	return out.Write(ctx, outputs)
}

// Test checks the connector state.
func (c *component) Test(_ map[string]any, setup *structpb.Struct) error {

	return nil
}
