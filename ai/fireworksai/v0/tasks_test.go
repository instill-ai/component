package fireworksai

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/gojuno/minimock/v3"
	"github.com/instill-ai/component/base"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestComponent_Tasks(t *testing.T) {
	mc := minimock.NewController(t)
	c := qt.New(t)
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)
	ctx := context.Background()

	FireworksClientMock := NewFireworksClientInterfaceMock(mc)
	FireworksClientMock.ChatMock.
		When(ChatRequest{
			Model: "accounts/fireworks/models/llama-v3p1-405b-instruct",
			N:     1,
			Messages: []FireworksChatRequestMessage{
				{
					Role: "user",
					Content: []FireworksMultiModalContent{
						{
							Type: FireworksContentTypeText,
							Text: "Tell me a joke",
						}}},
			}}).
		Then(ChatResponse{
			Model:   "accounts/fireworks/models/llama-v3p1-405b-instruct",
			Object:  FireworksResponseObjectChatCompletion,
			Created: 0,
			Choices: []FireWorksChoice{},
			Usage:   FireworksChatUsage{PromptTokens: 10, CompletionTokens: 18, TotalTokens: 28},
		}, nil)
	FireworksClientMock.ChatMock.
		When(ChatRequest{
			Model: "accounts/fireworks/models/gemini-1.5-pro",
			N:     1,
			Messages: []FireworksChatRequestMessage{
				{
					Role: "user",
					Content: []FireworksMultiModalContent{
						{
							Type: FireworksContentTypeText,
							Text: "Tell me a joke",
						}}},
			}}).
		Then(ChatResponse{}, fmt.Errorf("error when sending chat request %s", "unsuccessful HTTP response"))
	FireworksClientMock.EmbedMock.
		When(EmbedRequest{
			Model: "nomic-ai/nomic-embed-text-v1.5",
			Input: "The United Kingdom, made up of England, Scotland, Wales and Northern Ireland, is an island nation in northwestern Europe.",
		}).
		Then(EmbedResponse{
			Model:  "nomic-ai/nomic-embed-text-v1.5",
			Data:   []FireworksEmbedData{{Index: 0, Embedding: []float32{0.1, 0.2, 0.3}, Object: FireworksResponseObjectEmbedding}},
			Usage:  FireworksEmbedUsage{TotalTokens: 10},
			Object: FireworksObjectList}, nil)
	FireworksClientMock.EmbedMock.
		When(EmbedRequest{
			Model: "nomic-ai/nomic-embed-text-v1.87",
			Input: "The United Kingdom, made up of England, Scotland, Wales and Northern Ireland, is an island nation in northwestern Europe.",
		}).
		Then(EmbedResponse{}, fmt.Errorf("error when sending embeddings request %s", "unsuccessful HTTP response"))

	c.Run("ok - task text generation", func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{})
		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskTextGenerationChat},
			client:             FireworksClientMock,
		}
		e.execute = e.TaskTextGenerationChat
		exec := &base.ExecutionWrapper{Execution: e}

		pbIn, err := base.ConvertToStructpb(map[string]any{"model-name": "llama-v3p1-405b-instruct", "prompt": "Tell me a joke"})
		c.Assert(err, qt.IsNil)

		got, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})
		c.Assert(err, qt.IsNil)

		wantJSON, err := json.Marshal(TaskTextGenerationChatOuput{Text: "\nWhy did the tomato turn red?\nAnswer: Because it saw the salad dressing", Usage: TaskTextGenerationChatUsage{InputTokens: 10, OutputTokens: 18}})
		c.Assert(err, qt.IsNil)
		c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
	})

	c.Run("nok - task text generation", func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{})
		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskTextGenerationChat},
			client:             FireworksClientMock,
		}
		e.execute = e.TaskTextGenerationChat
		exec := &base.ExecutionWrapper{Execution: e}

		pbIn, err := base.ConvertToStructpb(map[string]any{"model-name": "gemini-1.5-pro", "prompt": "Tell me a joke"})
		c.Assert(err, qt.IsNil)

		_, err = exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})
		c.Assert(err, qt.ErrorMatches, `error when sending chat request model unsuccessful HTTP response`)
	})

	c.Run("ok - task embedding", func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{})
		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskTextEmbeddings},
			client:             FireworksClientMock,
		}
		e.execute = e.TaskTextEmbeddings
		exec := &base.ExecutionWrapper{Execution: e}

		pbIn, err := base.ConvertToStructpb(map[string]any{"model-name": "snowflake-arctic-embed:22m", "text": "The United Kingdom, made up of England, Scotland, Wales and Northern Ireland, is an island nation in northwestern Europe."})
		c.Assert(err, qt.IsNil)

		got, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})
		c.Assert(err, qt.IsNil)

		wantJSON, err := json.Marshal(TaskTextEmbeddingsOutput{Embedding: []float32{0.1, 0.2, 0.3, 0.4, 0.5}})
		c.Assert(err, qt.IsNil)
		c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
	})

	c.Run("nok - task embedding", func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{})
		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskTextEmbeddings},
			client:             FireworksClientMock,
		}
		e.execute = e.TaskTextEmbeddings
		exec := &base.ExecutionWrapper{Execution: e}

		pbIn, err := base.ConvertToStructpb(map[string]any{"model-name": "snowflake-arctic-embed:23m", "text": "The United Kingdom, made up of England, Scotland, Wales and Northern Ireland, is an island nation in northwestern Europe."})
		c.Assert(err, qt.IsNil)

		_, err = exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})
		c.Assert(err, qt.ErrorMatches, `error when sending embeddings unsuccessful HTTP response`)
	})

}
