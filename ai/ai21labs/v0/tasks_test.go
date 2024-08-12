package ai21labs

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/ai"
	"github.com/instill-ai/component/base"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type MockAI21labsClient struct {
}

const apiKey = "### FAKE API KEY ###"

func (c *MockAI21labsClient) Chat(req ChatRequest) (ChatResponse, error) {
	outputMessage := "Hello from AI21labs, (last message: " + req.Messages[len(req.Messages)-1].Content + ", message count: " + fmt.Sprintf("%v", len(req.Messages)) + ")"
	promptTokens := 0
	for _, message := range req.Messages {
		promptTokens += len(message.Content)
	}
	return ChatResponse{
		ID: "98f17392-0dad-44cd-9eb8-a09a6ff2bbef",
		Choices: []ChatChoice{
			{
				Index: 0,
				Message: ChatMessage{
					Role:    "assistant",
					Content: outputMessage,
				},
				FinishReason: "stop",
			},
		},
		Usage: ChatUsage{
			PromptTokens:     promptTokens,
			CompletionTokens: len(outputMessage),
			TotalTokens:      promptTokens + len(outputMessage),
		},
	}, nil
}

func (c *MockAI21labsClient) Embeddings(req EmbeddingsRequest) (EmbeddingsResponse, error) {
	return EmbeddingsResponse{
		ID: "98f17392-0dad-44cd-9eb8-a09a6ff2bbef",
		Results: []EmbeddingResult{
			{
				Embedding: []float32{0.1, 0.2, 0.3},
			},
		},
	}, nil
}

func (c *MockAI21labsClient) ContextualAnswers(req ContextualAnswersRequest) (ContextualAnswersResponse, error) {
	if len(req.Question) == 0 {
		return ContextualAnswersResponse{
			ID:              "98f17392-0dad-44cd-9eb8-a09a6ff2bbef",
			Answer:          "Not found",
			AnswerInContext: false,
		}, nil
	}
	return ContextualAnswersResponse{
		ID:              "98f17392-0dad-44cd-9eb8-a09a6ff2bbef",
		Answer:          req.Question + " is a question",
		AnswerInContext: true,
	}, nil
}

func (c *MockAI21labsClient) GrammaticalErrorCorrections(req GrammaticalErrorCorrectionsRequest) (GrammaticalErrorCorrectionsResponse, error) {
	return GrammaticalErrorCorrectionsResponse{
		ID: "98f17392-0dad-44cd-9eb8-a09a6ff2bbef",
		Corrections: []GrammaticalErrorCorrection{
			{
				Suggestion:     "ABC",
				StartIndex:     0,
				EndIndex:       3,
				OriginalText:   req.Text[0:3],
				CorrectionType: CorrectionType("spelling"),
			},
		},
	}, nil
}

func (c *MockAI21labsClient) Paraphrase(req ParaphraseRequest) (ParaphraseResponse, error) {
	return ParaphraseResponse{
		ID: "98f17392-0dad-44cd-9eb8-a09a6ff2bbef",
		Suggestions: []ParaphraseSuggestion{
			{Text: "ABC"},
		},
	}, nil
}

func (c *MockAI21labsClient) Summarize(req SummarizeRequest) (SummarizeResponse, error) {
	return SummarizeResponse{
		ID:      "98f17392-0dad-44cd-9eb8-a09a6ff2bbef",
		Summary: "ABC",
	}, nil
}

func (c *MockAI21labsClient) SummarizeBySegment(req SummarizeRequest) (SummerizeBySegmentResponse, error) {
	return SummerizeBySegmentResponse{
		ID: "98f17392-0dad-44cd-9eb8-a09a6ff2bbef",
		Segements: []SummerizeSegment{
			{
				Summary:     "abc",
				SegmentText: "ABC",
				SegmentHTML: "<h1>ABC</h1>",
				SegmentType: "title",
				HasSummary:  true,
				Highlights:  []SummerizeHighlight{},
			},
		},
	}, nil

}

func (c *MockAI21labsClient) TextImprovements(req TextImprovementsRequest) (TextImprovementsResponse, error) {
	return TextImprovementsResponse{
		ID: "98f17392-0dad-44cd-9eb8-a09a6ff2bbef",
		Improvements: []TextImprovement{{
			Suggestions:     []string{"ABC"},
			StartIndex:      0,
			EndIndex:        3,
			OriginalText:    req.Text[0:3],
			ImprovementType: ImprovementType("fluency"),
		}},
	}, nil
}

func (c *MockAI21labsClient) TextSegmentation(req TextSegmentationRequest) (TextSegmentationResponse, error) {
	return TextSegmentationResponse{
		ID: "98f17392-0dad-44cd-9eb8-a09a6ff2bbef",
		Segments: []TextSegmentationSegment{
			{
				SegementText: "ABC",
				SegmentType:  SegmentType("title"),
			},
		},
	}, nil
}

func (c *MockAI21labsClient) BaseURL() string {
	return ""
}

func TestTasks(t *testing.T) {

	c := qt.New(t)

	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)
	ctx := context.Background()
	c.Run("ok - task text generation", func(c *qt.C) {
		tc := struct {
			input    map[string]any
			wantResp TaskTextGenerationChatOutput
		}{
			input: map[string]any{"prompt": "Hello World!"},
			wantResp: TaskTextGenerationChatOutput{
				ai.TemplateTextGenerationOutput{
					Text: "Hello from AI21labs, (last message: Hello World!, message count: 2)",
					Usage: ai.GenerativeTextModelUsage{
						InputTokens:  12,
						OutputTokens: 67,
					},
				},
			},
		}
		setup, err := structpb.NewStruct(map[string]any{
			"api-key": apiKey,
		})
		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution:     base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskTextGenerationChat},
			client:                 &MockAI21labsClient{},
			usesInstillCredentials: false,
		}
		e.execute = e.TaskTextGenerationChat

		pbIn, err := base.ConvertToStructpb(tc.input)
		c.Assert(err, qt.IsNil)

		got, err := e.Execute(ctx, []*structpb.Struct{pbIn})
		c.Assert(err, qt.IsNil)

		wantJSON, err := json.Marshal(tc.wantResp)
		c.Assert(err, qt.IsNil)
		c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
	})

	c.Run("ok - task embedding", func(c *qt.C) {
		tc := struct {
			input    map[string]any
			wantResp TaskTextEmbeddingsOutput
		}{
			input: map[string]any{"text": "Hello World!"},
			wantResp: TaskTextEmbeddingsOutput{
				Embedding: []float32{0.1, 0.2, 0.3},
				Usage: ai.EmbeddingTextModelUsage{
					Tokens: len("Hello World!") / 2, // IMPORTANT: The vendor's API does not return the actual token count, so we are using a dummy value here.
				},
			},
		}
		setup, err := structpb.NewStruct(map[string]any{
			"api-key": apiKey,
		})
		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution:     base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskTextEmbeddings},
			client:                 &MockAI21labsClient{},
			usesInstillCredentials: false,
		}
		e.execute = e.TaskTextEmbeddings

		pbIn, err := base.ConvertToStructpb(tc.input)
		c.Assert(err, qt.IsNil)

		got, err := e.Execute(ctx, []*structpb.Struct{pbIn})
		c.Assert(err, qt.IsNil)

		wantJSON, err := json.Marshal(tc.wantResp)
		c.Assert(err, qt.IsNil)
		c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
	})

	c.Run("ok - task contextual answers without answers", func(c *qt.C) {

		tc := struct {
			input    map[string]any
			wantResp TaskContextualAnsweringOutput
		}{
			input: map[string]any{"question": ""},
			wantResp: TaskContextualAnsweringOutput{
				Answer:          "Not found",
				AnswerInContext: false,
			},
		}
		setup, err := structpb.NewStruct(map[string]any{
			"api-key": apiKey,
		})
		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution:     base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskContextualAnswering},
			client:                 &MockAI21labsClient{},
			usesInstillCredentials: false,
		}
		e.execute = e.TaskContextualAnswering

		pbIn, err := base.ConvertToStructpb(tc.input)
		c.Assert(err, qt.IsNil)

		got, err := e.Execute(ctx, []*structpb.Struct{pbIn})
		c.Assert(err, qt.IsNil)

		wantJSON, err := json.Marshal(tc.wantResp)
		c.Assert(err, qt.IsNil)
		c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())

	})

	c.Run("ok - task contextual answers with answers", func(c *qt.C) {

		tc := struct {
			input    map[string]any
			wantResp TaskContextualAnsweringOutput
		}{
			input: map[string]any{"question": "How's the weather today?"},
			wantResp: TaskContextualAnsweringOutput{
				Answer:          "How's the weather today?" + " is a question",
				AnswerInContext: true,
			},
		}
		setup, err := structpb.NewStruct(map[string]any{
			"api-key": apiKey,
		})
		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution:     base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskContextualAnswering},
			client:                 &MockAI21labsClient{},
			usesInstillCredentials: false,
		}
		e.execute = e.TaskContextualAnswering

		pbIn, err := base.ConvertToStructpb(tc.input)
		c.Assert(err, qt.IsNil)

		got, err := e.Execute(ctx, []*structpb.Struct{pbIn})
		c.Assert(err, qt.IsNil)

		wantJSON, err := json.Marshal(tc.wantResp)
		c.Assert(err, qt.IsNil)
		c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
	})

	c.Run("ok - task text summarization", func(c *qt.C) {
		tc := struct {
			input    map[string]any
			wantResp TaskTextSummarizationOutput
		}{
			input: map[string]any{},
			wantResp: TaskTextSummarizationOutput{
				Summary: "ABC",
			},
		}
		setup, err := structpb.NewStruct(map[string]any{
			"api-key": apiKey,
		})
		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution:     base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskTextSummarization},
			client:                 &MockAI21labsClient{},
			usesInstillCredentials: false,
		}
		e.execute = e.TaskTextSummarization

		pbIn, err := base.ConvertToStructpb(tc.input)
		c.Assert(err, qt.IsNil)

		got, err := e.Execute(ctx, []*structpb.Struct{pbIn})
		c.Assert(err, qt.IsNil)

		wantJSON, err := json.Marshal(tc.wantResp)
		c.Assert(err, qt.IsNil)
		c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
	})

	c.Run("ok - task text summarization by segment", func(c *qt.C) {
		tc := struct {
			input    map[string]any
			wantResp TaskTextSummarizationBySegmentOutput
		}{
			input: map[string]any{},
			wantResp: TaskTextSummarizationBySegmentOutput{
				Summerizations: []TextSegmentSummarization{
					{
						Text: "ABC",
						HTML: "<h1>ABC</h1>",
						Type: "title",
					},
				},
			},
		}
		setup, err := structpb.NewStruct(map[string]any{
			"api-key": apiKey,
		})
		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution:     base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskTextSummarizationBySegment},
			client:                 &MockAI21labsClient{},
			usesInstillCredentials: false,
		}
		e.execute = e.TaskTextSummarizationBySegment
		pbIn, err := base.ConvertToStructpb(tc.input)
		c.Assert(err, qt.IsNil)

		got, err := e.Execute(ctx, []*structpb.Struct{pbIn})
		c.Assert(err, qt.IsNil)

		wantJSON, err := json.Marshal(tc.wantResp)
		c.Assert(err, qt.IsNil)
		c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
	})

	c.Run("ok - task text paraphrasing", func(c *qt.C) {
		tc := struct {
			input    map[string]any
			wantResp TaskTextParaphrasingOutput
		}{
			input: map[string]any{},
			wantResp: TaskTextParaphrasingOutput{
				Suggestions: []string{"ABC"},
			},
		}
		setup, err := structpb.NewStruct(map[string]any{
			"api-key": apiKey,
		})
		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution:     base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskTextParaphrasing},
			client:                 &MockAI21labsClient{},
			usesInstillCredentials: false,
		}
		e.execute = e.TaskTextParaphrasing
		pbIn, err := base.ConvertToStructpb(tc.input)
		c.Assert(err, qt.IsNil)

		got, err := e.Execute(ctx, []*structpb.Struct{pbIn})
		c.Assert(err, qt.IsNil)

		wantJSON, err := json.Marshal(tc.wantResp)
		c.Assert(err, qt.IsNil)
		c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
	})

	c.Run("ok - task grammar check", func(c *qt.C) {
		tc := struct {
			input    map[string]any
			wantResp TaskGrammarCheckOutput
		}{
			input: map[string]any{"text": "Hello world!`"},
			wantResp: TaskGrammarCheckOutput{
				Suggestions: []GrammerSuggestion{
					{
						Text:       "ABC",
						StartIndex: 0,
						EndIndex:   3,
						Type:       "spelling",
					},
				},
			},
		}
		setup, err := structpb.NewStruct(map[string]any{
			"api-key": apiKey,
		})

		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution:     base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskGrammarCheck},
			client:                 &MockAI21labsClient{},
			usesInstillCredentials: false,
		}
		e.execute = e.TaskGrammarCheck
		pbIn, err := base.ConvertToStructpb(tc.input)
		c.Assert(err, qt.IsNil)

		got, err := e.Execute(ctx, []*structpb.Struct{pbIn})
		c.Assert(err, qt.IsNil)

		wantJSON, err := json.Marshal(tc.wantResp)
		c.Assert(err, qt.IsNil)
		c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
	})

	c.Run("ok - task text improvement", func(c *qt.C) {
		tc := struct {
			input    map[string]any
			wantResp TaskTextImprovementOutput
		}{
			input: map[string]any{"text": "Hello world!`"},
			wantResp: TaskTextImprovementOutput{
				Suggestions: []Improvement{
					{
						Texts:      []string{"ABC"},
						StartIndex: 0,
						EndIndex:   3,
						Type:       "fluency",
					},
				},
			},
		}
		setup, err := structpb.NewStruct(map[string]any{
			"api-key": apiKey,
		})

		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution:     base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskTextImprovement},
			client:                 &MockAI21labsClient{},
			usesInstillCredentials: false,
		}
		e.execute = e.TaskTextImprovement
		pbIn, err := base.ConvertToStructpb(tc.input)
		c.Assert(err, qt.IsNil)

		got, err := e.Execute(ctx, []*structpb.Struct{pbIn})
		c.Assert(err, qt.IsNil)

		wantJSON, err := json.Marshal(tc.wantResp)
		c.Assert(err, qt.IsNil)
		c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
	})

	c.Run("ok - task text segmentation", func(c *qt.C) {
		tc := struct {
			input    map[string]any
			wantResp TaskTextSegmentationOutput
		}{
			input: map[string]any{"text": "Hello world!`"},
			wantResp: TaskTextSegmentationOutput{
				Segments: []TextSegment{
					{
						Text: "ABC",
						Type: "title",
					},
				},
			},
		}
		setup, err := structpb.NewStruct(map[string]any{
			"api-key": apiKey,
		})

		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution:     base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TaskTextSegmentation},
			client:                 &MockAI21labsClient{},
			usesInstillCredentials: false,
		}
		e.execute = e.TaskTextSegmentation
		pbIn, err := base.ConvertToStructpb(tc.input)
		c.Assert(err, qt.IsNil)

		got, err := e.Execute(ctx, []*structpb.Struct{pbIn})
		c.Assert(err, qt.IsNil)

		wantJSON, err := json.Marshal(tc.wantResp)
		c.Assert(err, qt.IsNil)
		c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
	})
}
