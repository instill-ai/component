package mistralai

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"
	mistralSDK "github.com/gage-technologies/mistral-go"
	"github.com/instill-ai/component/base"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type MockMistralClient struct {
}

func (m *MockMistralClient) Embeddings(model string, input []string) (*mistralSDK.EmbeddingResponse, error) {
	return &mistralSDK.EmbeddingResponse{
		ID:     "embd-aad6fc62b17349b192ef09225058bc45",
		Object: "list",
		Data: []mistralSDK.EmbeddingObject{
			{
				Object:    "embedding",
				Embedding: []float64{1.0, 2.0, 3.0},
				Index:     0,
			},
		},
		Model: model,
		Usage: mistralSDK.UsageInfo{
			PromptTokens: len(input[0]),
			TotalTokens:  len(input[0]),
		},
	}, nil
}

func (m *MockMistralClient) Chat(model string, messages []mistralSDK.ChatMessage, params *mistralSDK.ChatRequestParams) (*mistralSDK.ChatCompletionResponse, error) {
	message := fmt.Sprintf("Hello Mistral! message count: %d", len(messages))
	promptToken := 0
	for _, m := range messages {
		promptToken += len(m.Content)
	}
	return &mistralSDK.ChatCompletionResponse{
		ID:      "cmpl-e5cc70bb28c444948073e77776eb30ef",
		Object:  "chat.completion",
		Created: 1702256327,
		Model:   model,
		Usage: mistralSDK.UsageInfo{
			PromptTokens:     promptToken,
			CompletionTokens: len(message),
			TotalTokens:      promptToken + len(message),
		},
		Choices: []mistralSDK.ChatCompletionResponseChoice{
			{
				Index: 0,
				Message: mistralSDK.ChatMessage{
					Role:    "assistant",
					Content: message,
				},
				FinishReason: mistralSDK.FinishReasonStop,
			},
		},
	}, nil
}

const (
	apiKey = "### MOCK API KEY ###"
)

func TestComponent_Execute(t *testing.T) {
	c := qt.New(t)

	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	c.Run("ok - supported task", func(c *qt.C) {
		task := TextGenerationTask

		_, err := connector.CreateExecution(nil, nil, task)
		c.Check(err, qt.IsNil)
	})
	c.Run("ok - supported task", func(c *qt.C) {
		task := TextEmbeddingTask

		_, err := connector.CreateExecution(nil, nil, task)
		c.Check(err, qt.IsNil)
	})

	c.Run("nok - unsupported task", func(c *qt.C) {
		task := "FOOBAR"

		_, err := connector.CreateExecution(nil, nil, task)
		c.Check(err, qt.ErrorMatches, "unsupported task")
	})
}

func TestComponent_Tasks(t *testing.T) {
	c := qt.New(t)

	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)
	ctx := context.Background()

	chatTc := struct {
		input    map[string]any
		wantResp TextGenerationOutput
	}{
		input:    map[string]any{"model-name": "open-mixtral-8x22b", "prompt": "Hello World"},
		wantResp: TextGenerationOutput{Text: "Hello Mistral! message count: 1", Usage: chatUsage{InputTokens: len("Hello World"), OutputTokens: len("Hello Mistral! message count: 1")}},
	}

	c.Run("ok - task text generation", func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"api-key": apiKey,
		})
		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TextGenerationTask},
			client:             MistralClient{sdkClient: &MockMistralClient{}, logger: nil},
		}
		e.execute = e.taskTextGeneration

		pbIn, err := base.ConvertToStructpb(chatTc.input)
		c.Assert(err, qt.IsNil)

		got, err := e.Execute(ctx, []*structpb.Struct{pbIn})
		c.Assert(err, qt.IsNil)

		wantJSON, err := json.Marshal(chatTc.wantResp)
		c.Assert(err, qt.IsNil)
		c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
	})

	embeddingTc := struct {
		input    map[string]any
		wantResp TextEmbeddingOutput
	}{
		input:    map[string]any{"model-name": "mistral-embed", "text": "Hello World"},
		wantResp: TextEmbeddingOutput{Embedding: []float64{1.0, 2.0, 3.0}, Usage: textEmbeddingUsage{Tokens: len("Hello World")}},
	}

	c.Run("ok - task embedding", func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"api-key": apiKey,
		})
		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TextEmbeddingTask},
			client:             MistralClient{sdkClient: &MockMistralClient{}, logger: nil},
		}
		e.execute = e.taskTextEmbedding

		pbIn, err := base.ConvertToStructpb(embeddingTc.input)
		c.Assert(err, qt.IsNil)

		got, err := e.Execute(ctx, []*structpb.Struct{pbIn})
		c.Assert(err, qt.IsNil)

		wantJSON, err := json.Marshal(embeddingTc.wantResp)
		c.Assert(err, qt.IsNil)
		c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
	})

}
