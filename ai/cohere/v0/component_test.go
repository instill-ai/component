package cohere

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	cohereSDK "github.com/cohere-ai/cohere-go/v2"
	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	apiKey = "cohere-api-key"
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

	commandTc := struct {
		input    map[string]any
		wantResp textGenerationOutput
	}{
		input:    map[string]any{"model-name": "command-r-plus"},
		wantResp: textGenerationOutput{Text: "Hi! My name is command-r-plus.", Ciatations: []ciatation{}, Usage: commandUsage{InputTokens: 20, OutputTokens: 30}},
	}

	c.Run("ok - task command", func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"api-key": apiKey,
		})
		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TextGenerationTask},
			client:             &MockCohereClient{},
		}
		e.execute = e.taskTextGeneration
		exec := &base.ExecutionWrapper{Execution: e}

		pbIn, err := base.ConvertToStructpb(commandTc.input)
		c.Assert(err, qt.IsNil)

		got, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})
		c.Assert(err, qt.IsNil)

		wantJSON, err := json.Marshal(commandTc.wantResp)
		c.Assert(err, qt.IsNil)
		c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())

	})

	embedTc := struct {
		input    map[string]any
		wantResp embeddingOutput
	}{
		input:    map[string]any{"text": "abcde"},
		wantResp: embeddingOutput{Embedding: []float64{0.1, 0.2, 0.3, 0.4, 0.5}, Usage: embedUsage{Tokens: 20}},
	}

	c.Run("ok - task embed", func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"api-key": apiKey,
		})
		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TextEmbeddingTask},
			client:             &MockCohereClient{},
		}
		e.execute = e.taskEmbedding
		exec := &base.ExecutionWrapper{Execution: e}

		pbIn, err := base.ConvertToStructpb(embedTc.input)
		c.Assert(err, qt.IsNil)

		got, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})
		c.Assert(err, qt.IsNil)

		wantJSON, err := json.Marshal(embedTc.wantResp)
		c.Assert(err, qt.IsNil)
		c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
	})

	rerankTc := struct {
		input    map[string]any
		wantResp rerankOutput
	}{
		input:    map[string]any{"documents": []string{"a", "b", "c", "d"}},
		wantResp: rerankOutput{Ranking: []string{"d", "c", "b", "a"}, Usage: rerankUsage{Search: 5}},
	}
	c.Run("ok - task rerank", func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"api-key": apiKey,
		})
		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: TextRerankTask},
			client:             &MockCohereClient{},
		}
		e.execute = e.taskRerank
		exec := &base.ExecutionWrapper{Execution: e}

		pbIn, err := base.ConvertToStructpb(rerankTc.input)
		c.Assert(err, qt.IsNil)

		got, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})
		c.Assert(err, qt.IsNil)

		wantJSON, err := json.Marshal(rerankTc.wantResp)
		c.Assert(err, qt.IsNil)
		c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
	})

}

type MockCohereClient struct{}

func (m *MockCohereClient) generateTextChat(request cohereSDK.ChatRequest) (cohereSDK.NonStreamedChatResponse, error) {
	tx := fmt.Sprintf("Hi! My name is %s.", *request.Model)
	cia := []*cohereSDK.ChatCitation{}
	inputToken := float64(20)
	outputToken := float64(30)
	bill := cohereSDK.ApiMetaBilledUnits{InputTokens: &inputToken, OutputTokens: &outputToken}
	meta := cohereSDK.ApiMeta{BilledUnits: &bill}
	return cohereSDK.NonStreamedChatResponse{
		Citations: cia,
		Text:      tx,
		Meta:      &meta,
	}, nil
}

func (m *MockCohereClient) generateEmbedding(request cohereSDK.EmbedRequest) (cohereSDK.EmbedResponse, error) {
	inputToken := float64(20)
	bill := cohereSDK.ApiMetaBilledUnits{InputTokens: &inputToken}
	meta := cohereSDK.ApiMeta{BilledUnits: &bill}
	embedding := cohereSDK.EmbedFloatsResponse{
		Embeddings: [][]float64{{0.1, 0.2, 0.3, 0.4, 0.5}},
		Meta:       &meta,
	}
	return cohereSDK.EmbedResponse{
		EmbeddingsFloats: &embedding,
	}, nil
}

func (m *MockCohereClient) generateRerank(request cohereSDK.RerankRequest) (cohereSDK.RerankResponse, error) {
	documents := []cohereSDK.RerankResponseResultsItemDocument{
		{Text: request.Documents[3].String},
		{Text: request.Documents[2].String},
		{Text: request.Documents[1].String},
		{Text: request.Documents[0].String},
	}
	result := []*cohereSDK.RerankResponseResultsItem{
		{Document: &documents[0]},
		{Document: &documents[1]},
		{Document: &documents[2]},
		{Document: &documents[3]},
	}
	searchCnt := float64(5)
	bill := cohereSDK.ApiMetaBilledUnits{SearchUnits: &searchCnt}
	meta := cohereSDK.ApiMeta{BilledUnits: &bill}
	return cohereSDK.RerankResponse{
		Results: result,
		Meta:    &meta,
	}, nil
}
