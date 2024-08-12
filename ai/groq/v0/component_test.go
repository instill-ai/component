package groq

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gojuno/minimock/v3"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	qt "github.com/frankban/quicktest"

	"github.com/instill-ai/component/base"
)

const (
	MockAPIKey = "### Mock API Key ###"
)

func TestComponent_Execute(t *testing.T) {
	c := qt.New(t)

	bc := base.Component{Logger: zap.NewNop()}
	cmp := Init(bc)

	c.Run("ok - supported task", func(c *qt.C) {
		task := TaskTextGenerationChat

		_, err := cmp.CreateExecution(base.ComponentExecution{
			Component: cmp,
			Task:      task,
		})
		c.Check(err, qt.IsNil)
	})

	c.Run("nok - unsupported task", func(c *qt.C) {
		task := "FOOBAR"

		_, err := cmp.CreateExecution(base.ComponentExecution{
			Component: cmp,
			Task:      task,
		})
		c.Check(err, qt.ErrorMatches, "unsupported task")
	})
}

func TestComponent_Tasks(t *testing.T) {
	mc := minimock.NewController(t)
	c := qt.New(t)
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)
	ctx := context.Background()

	GroqClientMock := NewGroqClientInterfaceMock(mc)
	GroqClientMock.ChatMock.
		When(ChatRequest{
			Model: "llama-3.1-405b-reasoning",
			Messages: []GroqChatMessageInterface{
				GroqChatMessage{
					Role: "user",
					Content: []GroqChatContent{
						{
							Text: "Tell me a joke",
							Type: GroqChatContentTypeText,
						},
					},
				},
			},
			N:    1,
			Stop: []string{},
		}).
		Then(ChatResponse{
			ID:      "34a9110d-c39d-423b-9ab9-9c748747b204",
			Object:  "chat.completion",
			Model:   "llama-3.1-405b-reasoning",
			Created: 1708045122,
			Usage: GroqUsage{
				PromptTokens:     24,
				CompletionTokens: 377,
				TotalTokens:      401,
				PromptTime:       0.009,
				CompletionTime:   0.774,
				TotalTime:        0.783,
			},
			Choices: []GroqChoice{
				{
					Index:        0,
					FinishReason: "stop",
					Message: GroqResponseMessage{
						Role:    "assistant",
						Content: "\nWhy did the tomato turn red?\nAnswer: Because it saw the salad dressing",
					},
				},
			},
		}, nil)
	GroqClientMock.ChatMock.
		When(ChatRequest{
			Model: "gemini",
			Messages: []GroqChatMessageInterface{
				GroqChatMessage{
					Role: "user",
					Content: []GroqChatContent{
						{
							Text: "Tell me a joke",
							Type: GroqChatContentTypeText,
						},
					},
				},
			},
			N:    1,
			Stop: []string{},
		}).
		Then(ChatResponse{}, fmt.Errorf("error when sending chat request %s", `no access to "gemini"`))

	c.Run("ok - task text generation", func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"api-key": MockAPIKey,
		})
		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskTextGenerationChat},
			client:             GroqClientMock,
		}
		e.execute = e.TaskTextGenerationChat

		pbIn, err := base.ConvertToStructpb(map[string]any{"model": "llama-3.1-405b-reasoning", "prompt": "Tell me a joke"})
		c.Assert(err, qt.IsNil)

		got, err := e.Execute(ctx, []*structpb.Struct{pbIn})
		c.Assert(err, qt.IsNil)

		wantJSON, err := json.Marshal(TaskTextGenerationChatOuput{Text: "\nWhy did the tomato turn red?\nAnswer: Because it saw the salad dressing", Usage: TaskTextGenerationChatUsage{InputTokens: 24, OutputTokens: 377}})
		c.Assert(err, qt.IsNil)
		c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
	})

	c.Run("nok - task text generation", func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"api-key": MockAPIKey,
		})
		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskTextGenerationChat},
			client:             GroqClientMock,
		}
		e.execute = e.TaskTextGenerationChat

		pbIn, err := base.ConvertToStructpb(map[string]any{"model": "gemini", "prompt": "Tell me a joke"})
		c.Assert(err, qt.IsNil)

		_, err = e.Execute(ctx, []*structpb.Struct{pbIn})
		c.Assert(err, qt.ErrorMatches, `error when sending chat request no access to "gemini"`)
	})

}