//go:generate compogen readme ./config ./README.mdx
package openai

import (
	"bufio"
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/gabriel-vasile/mimetype"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/util/httpclient"
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
	if setup == nil || setup.Fields == nil {
		setup = &structpb.Struct{Fields: map[string]*structpb.Value{}}
	}
	if v, ok := setup.GetFields()[cfgAPIKey]; ok {
		apiKey := v.GetStringValue()
		if apiKey != "" && apiKey != base.SecretKeyword {
			return setup, false, nil
		}
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

type workerResult struct {
	batchIdx int
	err      error
	output   *structpb.Struct
	done     bool
}

func (e *execution) worker(ctx context.Context, client *httpclient.Client, batchIdx int, input *structpb.Struct, result chan<- *workerResult) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("panic: %+v", r)
			result <- &workerResult{batchIdx: batchIdx, err: fmt.Errorf("panic: %+v", r)}
		}
		result <- &workerResult{done: true}
	}()

	switch e.Task {
	case TextGenerationTask:
		inputStruct := TextCompletionInput{}
		err := base.ConvertFromStructpb(input, &inputStruct)
		if err != nil {
			result <- &workerResult{batchIdx: batchIdx, err: err}
			return
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
				result <- &workerResult{batchIdx: batchIdx, err: err}
				return
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
			Stream:           true,
			StreamOptions: streamOptions{
				IncludeUsage: true,
			},
		}

		// workaround, the OpenAI service can not accept this param
		if inputStruct.Model != "gpt-4-vision-preview" {
			if inputStruct.ResponseFormat != nil {
				body.ResponseFormat = &responseFormatReqStruct{
					Type: inputStruct.ResponseFormat.Type,
				}
				if inputStruct.ResponseFormat.Type == "json_schema" {
					if inputStruct.Model == "gpt-4o-mini" || inputStruct.Model == "gpt-4o-2024-08-06" {
						sch := map[string]any{}
						if inputStruct.ResponseFormat.JSONSchema != "" {
							err = json.Unmarshal([]byte(inputStruct.ResponseFormat.JSONSchema), &sch)
							if err != nil {
								result <- &workerResult{batchIdx: batchIdx, err: err}

								return
							}
							body.ResponseFormat = &responseFormatReqStruct{
								Type:       inputStruct.ResponseFormat.Type,
								JSONSchema: sch,
							}
						}

					} else {
						result <- &workerResult{batchIdx: batchIdx, err: fmt.Errorf("this model doesn't support response format: json_schema")}

						return
					}

				}
			}

		}

		req := client.SetDoNotParseResponse(true).R().SetBody(body)
		restyResp, err := req.Post(completionsPath)
		if err != nil {
			result <- &workerResult{batchIdx: batchIdx, err: err}
			return
		}

		if restyResp.StatusCode() != 200 {
			res := restyResp.Body()
			result <- &workerResult{batchIdx: batchIdx, err: fmt.Errorf("send request to openai error with error code: %d, msg %s", restyResp.StatusCode(), res)}
			return
		}
		scanner := bufio.NewScanner(restyResp.RawResponse.Body)

		outputStruct := TextCompletionOutput{}

		u := usage{}
		count := 0
		for scanner.Scan() {

			res := scanner.Text()

			if len(res) == 0 {
				continue
			}
			res = strings.Replace(res, "data: ", "", 1)

			// Note: Since we haven’t provided delta updates for the
			// messages, we’re reducing the number of event streams by
			// returning the response every ten iterations.
			if count == 10 || res == "[DONE]" {
				outputJSON, inErr := json.Marshal(outputStruct)
				if inErr != nil {
					result <- &workerResult{batchIdx: batchIdx, err: inErr}
					return
				}
				output := &structpb.Struct{}
				inErr = protojson.Unmarshal(outputJSON, output)
				if inErr != nil {
					result <- &workerResult{batchIdx: batchIdx, err: inErr}
					return
				}
				result <- &workerResult{
					batchIdx: batchIdx,
					output:   output,
				}
				if res == "[DONE]" {
					break
				}
				count = 0
			}

			count += 1
			response := &textCompletionResp{}
			err = json.Unmarshal([]byte(res), response)
			if err != nil {
				result <- &workerResult{batchIdx: batchIdx, err: err}
				return
			}

			if outputStruct.Texts == nil {
				outputStruct.Texts = make([]string, len(response.Choices))
			}
			for idx, c := range response.Choices {
				outputStruct.Texts[idx] += c.Delta.Content

			}

			u = usage{
				PromptTokens:     response.Usage.PromptTokens,
				CompletionTokens: response.Usage.CompletionTokens,
				TotalTokens:      response.Usage.TotalTokens,
			}

		}

		outputStruct.Usage = u
		outputJSON, err := json.Marshal(outputStruct)
		if err != nil {
			result <- &workerResult{batchIdx: batchIdx, err: err}
			return
		}
		output := &structpb.Struct{}
		err = protojson.Unmarshal(outputJSON, output)
		if err != nil {
			result <- &workerResult{batchIdx: batchIdx, err: err}
			return
		}
		result <- &workerResult{
			batchIdx: batchIdx,
			output:   output,
		}

	case TextEmbeddingsTask:
		inputStruct := TextEmbeddingsInput{}
		err := base.ConvertFromStructpb(input, &inputStruct)
		if err != nil {
			result <- &workerResult{batchIdx: batchIdx, err: err}
			return
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
			result <- &workerResult{batchIdx: batchIdx, err: err}
			return
		}

		outputStruct := TextEmbeddingsOutput{
			Embedding: resp.Data[0].Embedding,
		}

		output, err := base.ConvertToStructpb(outputStruct)
		if err != nil {
			result <- &workerResult{batchIdx: batchIdx, err: err}
			return
		}
		result <- &workerResult{
			batchIdx: batchIdx,
			output:   output,
		}

	case SpeechRecognitionTask:
		inputStruct := AudioTranscriptionInput{}
		err := base.ConvertFromStructpb(input, &inputStruct)
		if err != nil {
			result <- &workerResult{batchIdx: batchIdx, err: err}

			return
		}

		audioBytes, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(inputStruct.Audio))
		if err != nil {
			result <- &workerResult{batchIdx: batchIdx, err: err}
			return
		}

		data, ct, err := getBytes(AudioTranscriptionReq{
			File:        audioBytes,
			Model:       inputStruct.Model,
			Prompt:      inputStruct.Prompt,
			Language:    inputStruct.Language,
			Temperature: inputStruct.Temperature,

			// Verbosity is passed to extract result duration.
			ResponseFormat: "verbose_json",
		})
		if err != nil {
			result <- &workerResult{batchIdx: batchIdx, err: err}
			return
		}

		resp := AudioTranscriptionResp{}
		req := client.R().SetBody(data).SetResult(&resp).SetHeader("Content-Type", ct)
		if _, err := req.Post(transcriptionsPath); err != nil {
			result <- &workerResult{batchIdx: batchIdx, err: err}
			return
		}

		output, err := base.ConvertToStructpb(resp)
		if err != nil {
			result <- &workerResult{batchIdx: batchIdx, err: err}
			return
		}
		result <- &workerResult{
			batchIdx: batchIdx,
			output:   output,
		}

	case TextToSpeechTask:
		inputStruct := TextToSpeechInput{}
		err := base.ConvertFromStructpb(input, &inputStruct)
		if err != nil {
			result <- &workerResult{batchIdx: batchIdx, err: err}
			return
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
			result <- &workerResult{batchIdx: batchIdx, err: err}
			return
		}

		audio := base64.StdEncoding.EncodeToString(resp.Body())
		outputStruct := TextToSpeechOutput{
			Audio: fmt.Sprintf("data:audio/wav;base64,%s", audio),
		}

		output, err := base.ConvertToStructpb(outputStruct)
		if err != nil {
			result <- &workerResult{batchIdx: batchIdx, err: err}
			return
		}
		result <- &workerResult{
			batchIdx: batchIdx,
			output:   output,
		}

	case TextToImageTask:

		inputStruct := ImagesGenerationInput{}
		err := base.ConvertFromStructpb(input, &inputStruct)
		if err != nil {
			result <- &workerResult{batchIdx: batchIdx, err: err}
			return
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
			result <- &workerResult{batchIdx: batchIdx, err: err}
			return
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
			result <- &workerResult{batchIdx: batchIdx, err: err}
			return
		}
		result <- &workerResult{
			batchIdx: batchIdx,
			output:   output,
		}

	default:
		result <- &workerResult{batchIdx: batchIdx, err: errmsg.AddMessage(
			fmt.Errorf("not supported task: %s", e.Task),
			fmt.Sprintf("%s task is not supported.", e.Task),
		)}
		return
	}
}

func (e *execution) Execute(ctx context.Context, in base.InputReader, out base.OutputWriter) error {
	inputs, err := in.Read(ctx)
	if err != nil {
		return err
	}
	client := newClient(e.Setup, e.GetLogger())
	outputs := make([]*structpb.Struct, len(inputs))

	result := make(chan *workerResult)
	defer close(result)

	for batchIdx, input := range inputs {
		go e.worker(ctx, client, batchIdx, input, result)
	}

	// TODO: Optimize the concurrent implementation by
	// 1. Creating InputReader/OutputWriter for single data instead of batch
	// data
	// 2. Moving concurrent implementation to the component base execution
	count := 0
	for {
		r := <-result
		if r.done {
			count += 1
			if count == len(inputs) {
				break
			}

		} else if r.err != nil {
			err = r.err
		} else {
			outputs[r.batchIdx] = r.output
			err = out.Write(ctx, outputs)
		}
	}

	return err
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
