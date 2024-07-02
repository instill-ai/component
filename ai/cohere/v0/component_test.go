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
		task := textGenerationTask

		_, err := connector.CreateExecution(nil, nil, task)
		c.Check(err, qt.IsNil)
	})

	c.Run("nok - unsupported task", func(c *qt.C) {
		task := "FOOBAR"

		_, err := connector.CreateExecution(nil, nil, task)
		c.Check(err, qt.ErrorMatches, "unsupported task")
	})

}

func TestComponent_Generation(t *testing.T) {
	c := qt.New(t)

	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)
	ctx := context.Background()

	commandTc := struct {
		input    map[string]any
		wantResp commandOutput
	}{
		input:    map[string]any{"model-name": "command-r-plus"},
		wantResp: commandOutput{Text: "Hi! My name is command-r-plus.", Ciatations: []ciatation{}},
	}

	c.Run("ok - task command", func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"api-key": apiKey,
		})
		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: textGenerationTask},
			client:             &MockCohereClient{},
		}
		e.execute = e.taskCommand
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
		wantResp embedOutput
	}{
		input:    map[string]any{"text": "abcde"},
		wantResp: embedOutput{Embedding: []float64{0.1, 0.2, 0.3, 0.4, 0.5}},
	}

	c.Run("ok - task embed", func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"api-key": apiKey,
		})
		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: textEmbeddingTask},
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
		wantResp: rerankOutput{Ranking: []string{"d", "c", "b", "a"}},
	}
	c.Run("ok - task rerank", func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"api-key": apiKey,
		})
		c.Assert(err, qt.IsNil)
		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: textRerankTask},
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
	return cohereSDK.NonStreamedChatResponse{
		Citations: cia,
		Text:      tx,
	}, nil
}

func (m *MockCohereClient) generateEmbedding(request cohereSDK.EmbedRequest) (cohereSDK.EmbedResponse, error) {
	embedding := cohereSDK.EmbedFloatsResponse{
		Embeddings: [][]float64{{0.1, 0.2, 0.3, 0.4, 0.5}},
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
	return cohereSDK.RerankResponse{
		Results: result,
	}, nil
}
