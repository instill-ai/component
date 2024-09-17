package openaiv1

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/instill-ai/component/ai"
	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/util"
	"github.com/instill-ai/component/internal/util/httpclient"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	completionsPath = "/v1/chat/completions"
)

func (e *execution) ExecuteTextChat(input *structpb.Struct, job *base.Job, ctx context.Context) (*structpb.Struct, error) {
	inputStruct := ai.TextChatInput{}

	if err := base.ConvertFromStructpb(input, &inputStruct); err != nil {
		return nil, fmt.Errorf("failed to convert input to TextChatInput: %w", err)
	}

	return ExecuteTextChat(inputStruct, e.client, job, ctx)
}

func ExecuteTextChat(inputStruct ai.TextChatInput, client httpclient.IClient, job *base.Job, ctx context.Context) (*structpb.Struct, error) {

	requester := ModelRequesterFactory(inputStruct, client)

	return requester.SendChatRequest(job, ctx)
}

func ModelRequesterFactory(input ai.TextChatInput, client httpclient.IClient) IChatModelRequester {
	ChatModelList := []string{
		"gpt-3.5-turbo-16k",
		"gpt-4",
		"gpt-4-0314",
		"gpt-4-0613",
		"gpt-4-32k",
		"gpt-4-32k-0314",
		"gpt-4-32k-0613",
		"gpt-3.5-turbo-0301",
		"gpt-3.5-turbo-0125",
		"gpt-3.5-turbo-16k-0613",
	}

	model := input.Data.Model
	if util.InSlice(ChatModelList, model) {
		return &ChatModelRequester{
			Input:  input,
			Client: client,
		}
	}
	if model == "o1-preview" || model == "o1-mini" {
		return &O1ModelRequester{
			Input:  input,
			Client: client,
		}
	}
	return &SupportJSONOutputModelRequester{
		Input:  input,
		Client: client,
	}
}

type IChatModelRequester interface {
	SendChatRequest(*base.Job, context.Context) (*structpb.Struct, error)
}

// o1-preview or o1-mini
type O1ModelRequester struct {
	Input  ai.TextChatInput
	Client httpclient.IClient
}

// When it supports streaming, the job and ctx will be used.
func (r *O1ModelRequester) SendChatRequest(_ *base.Job, _ context.Context) (*structpb.Struct, error) {

	input := r.Input
	// Note: The o1-series models don't support streaming.
	input.Parameter.Stream = false

	chatReq := convertToTextChatReq(input)

	resp := textChatResp{}
	client := r.Client

	req := client.R().SetResult(&resp).SetBody(chatReq)

	if _, err := req.Post(completionsPath); err != nil {
		return nil, fmt.Errorf("failed to send chat request: %w", err)
	}

	outputStruct := ai.TextChatOutput{
		Data: ai.OutputData{
			Choices: make([]ai.Choice, len(resp.Choices)),
		},
		Metadata: ai.Metadata{
			Usage: ai.Usage{
				CompletionTokens: resp.Usage.ChatTokens,
				PromptTokens:     resp.Usage.PromptTokens,
				TotalTokens:      resp.Usage.TotalTokens,
			},
		},
	}

	for i, choice := range resp.Choices {
		outputStruct.Data.Choices[i] = ai.Choice{
			FinishReason: choice.FinishReason,
			Index:        choice.Index,
			Message: ai.OutputMessage{
				Content: choice.Message.Content,
				Role:    choice.Message.Role,
			},
			Created: resp.Created,
		}
	}

	return base.ConvertToStructpb(outputStruct)
}

// https://platform.openai.com/docs/api-reference/chat/create#chat-create-response_format
// Compatible with GPT-4o, GPT-4o mini, GPT-4 Turbo and all GPT-3.5 Turbo models newer than gpt-3.5-turbo-1106.
type SupportJSONOutputModelRequester struct {
	Input  ai.TextChatInput
	Client httpclient.IClient
}

func (r *SupportJSONOutputModelRequester) SendChatRequest(job *base.Job, ctx context.Context) (*structpb.Struct, error) {

	input := r.Input

	chatReq := convertToTextChatReq(input)

	// Note: Add response format to the request.
	// We will need to think about how to customize input & output for standardized AI components.

	output, err := sendRequest(chatReq, r.Client, job, ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to send chat request: %w", err)
	}

	return base.ConvertToStructpb(output)
}

type ChatModelRequester struct {
	Input  ai.TextChatInput
	Client httpclient.IClient
}

func (r *ChatModelRequester) SendChatRequest(job *base.Job, ctx context.Context) (*structpb.Struct, error) {

	input := r.Input

	chatReq := convertToTextChatReq(input)
	chatReq.Stream = true
	chatReq.StreamOptions = &streamOptions{
		IncludeUsage: true,
	}

	output, err := sendRequest(chatReq, r.Client, job, ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to send chat request: %w", err)
	}

	return base.ConvertToStructpb(output)
}

func sendRequest(chatReq textChatReq, client httpclient.IClient, job *base.Job, ctx context.Context) (ai.TextChatOutput, error) {

	req := client.SetDoNotParseResponse(true).R().SetBody(chatReq)

	outputStruct := ai.TextChatOutput{}
	restyResp, err := req.Post(completionsPath)

	if err != nil {
		return outputStruct, fmt.Errorf("failed to send chat request: %w", err)
	}

	if restyResp.StatusCode() != 200 {
		res := restyResp.Body()
		return outputStruct, fmt.Errorf("send request to openai error with error code: %d, msg %s", restyResp.StatusCode(), res)
	}

	scanner := bufio.NewScanner(restyResp.RawResponse.Body)

	var u ai.Usage

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
				return outputStruct, inErr
			}
			output := &structpb.Struct{}
			inErr = protojson.Unmarshal(outputJSON, output)
			if inErr != nil {
				return outputStruct, inErr
			}
			err = job.Output.Write(ctx, output)
			if err != nil {
				return outputStruct, err
			}
			if res == "[DONE]" {
				break
			}
			count = 0
		}

		count += 1
		response := &textChatStreamResp{}
		err = json.Unmarshal([]byte(res), response)

		if err != nil {
			return outputStruct, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		if len(outputStruct.Data.Choices) == 0 {
			outputStruct.Data.Choices = make([]ai.Choice, len(response.Choices))
		}
		for idx, c := range response.Choices {
			outputStruct.Data.Choices[idx].Message.Content += c.Delta.Content

			if c.Delta.Role != "" {
				outputStruct.Data.Choices[idx].Message.Role = c.Delta.Role
			}
		}

		if response.Usage.TotalTokens > 0 {
			u = ai.Usage{
				CompletionTokens: response.Usage.ChatTokens,
				PromptTokens:     response.Usage.PromptTokens,
				TotalTokens:      response.Usage.TotalTokens,
			}
		}
	}

	outputStruct.Metadata.Usage = u

	return outputStruct, nil
}

// Build the vendor-specific request structure
func convertToTextChatReq(input ai.TextChatInput) textChatReq {
	messages := buildMessages(input)

	temp := float32(input.Parameter.Temperature)
	topP := float32(input.Parameter.TopP)
	n := input.Parameter.N
	maxTokens := input.Parameter.MaxTokens
	seed := input.Parameter.Seed

	return textChatReq{
		Model:       input.Data.Model,
		Messages:    messages,
		MaxTokens:   &maxTokens,
		Temperature: &temp,
		N:           &n,
		TopP:        &topP,
		Seed:        &seed,
		Stream:      input.Parameter.Stream,
	}
}

func buildMessages(input ai.TextChatInput) []interface{} {
	messages := make([]interface{}, len(input.Data.Messages))
	for i, msg := range input.Data.Messages {
		content := make([]map[string]interface{}, len(msg.Contents))
		for j, c := range msg.Contents {
			content[j] = map[string]interface{}{
				"type": c.Type,
			}
			if c.Type == "text" {
				content[j]["text"] = c.Text
			}
			if c.Type == "image-url" {
				content[j]["image_url"] = c.ImageURL
			}
			if c.Type == "image-base64" {
				content[j]["image_url"] = util.GetDataURL(c.ImageBase64)
			}
		}

		messages[i] = map[string]interface{}{
			"role":    msg.Role,
			"name":    msg.Name,
			"content": content,
		}
	}

	return messages
}

// API request and response structures
type textChatReq struct {
	Model            string                   `json:"model"`
	Messages         []interface{}            `json:"messages"`
	Temperature      *float32                 `json:"temperature,omitempty"`
	TopP             *float32                 `json:"top_p,omitempty"`
	N                *int                     `json:"n,omitempty"`
	Stop             *string                  `json:"stop,omitempty"`
	Seed             *int                     `json:"seed,omitempty"`
	MaxTokens        *int                     `json:"max_tokens,omitempty"`
	PresencePenalty  *float32                 `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float32                 `json:"frequency_penalty,omitempty"`
	ResponseFormat   *responseFormatReqStruct `json:"response_format,omitempty"`
	Stream           bool                     `json:"stream"`
	StreamOptions    *streamOptions           `json:"stream_options,omitempty"`
}

type streamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type responseFormatReqStruct struct {
	Type       string         `json:"type,omitempty"`
	JSONSchema map[string]any `json:"json_schema,omitempty"`
}

type textChatStreamResp struct {
	ID      string          `json:"id"`
	Object  string          `json:"object"`
	Created int             `json:"created"`
	Choices []streamChoices `json:"choices"`
	Usage   usageOpenAI     `json:"usage"`
}

type textChatResp struct {
	ID      string      `json:"id"`
	Object  string      `json:"object"`
	Created int         `json:"created"`
	Choices []choice    `json:"choices"`
	Usage   usageOpenAI `json:"usage"`
}

type outputMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type streamChoices struct {
	Index        int           `json:"index"`
	FinishReason string        `json:"finish_reason"`
	Delta        outputMessage `json:"delta"`
}

type choice struct {
	Index        int           `json:"index"`
	FinishReason string        `json:"finish_reason"`
	Message      outputMessage `json:"message"`
}

type usageOpenAI struct {
	PromptTokens int `json:"prompt_tokens"`
	ChatTokens   int `json:"completion_tokens"`
	TotalTokens  int `json:"total_tokens"`
}
