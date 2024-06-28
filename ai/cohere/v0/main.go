//go:generate compogen readme ./config ./README.mdx
package cohere

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"slices"
	"sync"

	cohereSDK "github.com/cohere-ai/cohere-go/v2"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	textGenerationTask = "TASK_TEXT_GENERATION_CHAT"
	textEmbeddingTask  = "TASK_TEXT_EMBEDDINGS"
	textRerankTask     = "TASK_TEXT_RERANKING"
	cfgAPIKey          = "api-key"
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

type component struct {
	base.Component
}

type CohereClient interface {
	generateTextChat(request cohereSDK.ChatRequest) (cohereSDK.NonStreamedChatResponse, error)
	generateEmbedding(request cohereSDK.EmbedRequest) (cohereSDK.EmbedResponse, error)
	generateRerank(request cohereSDK.RerankRequest) (cohereSDK.RerankResponse, error)
}

// These structs are used to send the request /  parse the response from the API, this following their naming convension.
// reference: https://docs.anthropic.com/en/api/messages

type ciatation struct {
	Start int    `json:"start"`
	End   int    `json:"end"`
	Text  string `json:"text"`
}

type commandOutput struct {
	Text       string      `json:"text"`
	Ciatations []ciatation `json:"citations"`
}

type embedOutput struct {
	Embedding []float64 `json:"embedding"`
}

type rerankOutput struct {
	Ranking []string `json:"ranking"`
}

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

type execution struct {
	base.ComponentExecution
	execute func(*structpb.Struct) (*structpb.Struct, error)
	client  CohereClient
}

func (c *component) CreateExecution(sysVars map[string]any, setup *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	e := &execution{
		ComponentExecution: base.ComponentExecution{Component: c, SystemVariables: sysVars, Task: task, Setup: setup},
		client:             newClient(getAPIKey(setup), c.GetLogger()),
	}
	switch task {
	case textGenerationTask:
		e.execute = e.taskCommand
	case textEmbeddingTask:
		e.execute = e.taskEmbedding
	case textRerankTask:
		e.execute = e.taskRerank
	default:
		return nil, fmt.Errorf("unsupported task")
	}
	return &base.ExecutionWrapper{Execution: e}, nil
}
func (e *execution) Execute(_ context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := make([]*structpb.Struct, len(inputs))

	// The execution takes a array of inputs and returns an array of outputs. The execution is done sequentially.
	for i, input := range inputs {
		output, err := e.execute(input)
		if err != nil {
			return nil, err
		}

		outputs[i] = output
	}

	return outputs, nil
}

func (e *execution) taskCommand(in *structpb.Struct) (*structpb.Struct, error) {

	prompt := in.Fields["prompt"].GetStringValue()

	messages := []*cohereSDK.Message{}

	systemPrompt := in.Fields["system-prompt"].GetStringValue()

	if systemPrompt != "" {
		message := cohereSDK.Message{}
		message.Role = "SYSTEM"
		message.Chatbot = &cohereSDK.ChatMessage{Message: systemPrompt}
		messages = slices.Insert(messages, 0, &message)
	}

	documents := []map[string]string{}
	if in.Fields["documents"] != nil {
		for _, doc := range in.Fields["documents"].GetListValue().Values {
			document := map[string]string{}
			document["text"] = doc.GetStringValue()
			documents = append(documents, document)
		}
	}

	modelName := in.Fields["model-name"].GetStringValue()
	maxTokens := int(in.Fields["max-new-tokens"].GetNumberValue())
	temperature := float64(in.Fields["temperature"].GetNumberValue())
	topK := int(in.Fields["top-k"].GetNumberValue())
	seed := int(in.Fields["seed"].GetNumberValue())

	req := cohereSDK.ChatRequest{
		Message:     prompt,
		Model:       &modelName,
		ChatHistory: messages,
		MaxTokens:   &maxTokens,
		Temperature: &temperature,
		K:           &topK,
		Seed:        &seed,
		Documents:   documents,
	}

	resp, err := e.client.generateTextChat(req)

	if err != nil {
		return nil, err
	}

	citations := []ciatation{}

	for _, c := range resp.Citations {
		citation := ciatation{
			Start: c.Start,
			End:   c.End,
			Text:  c.Text,
		}
		citations = append(citations, citation)
	}

	outputStruct := commandOutput{
		Text:       resp.Text,
		Ciatations: citations,
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
	return &output, nil
}

func (e *execution) taskEmbedding(in *structpb.Struct) (*structpb.Struct, error) {
	text := in.Fields["text"].GetStringValue()
	texts := []string{text}
	modelName := in.Fields["model-name"].GetStringValue()
	inputType := in.Fields["input-type"].GetStringValue()
	req := cohereSDK.EmbedRequest{
		Texts:     texts,
		Model:     &modelName,
		InputType: (*cohereSDK.EmbedInputType)(&inputType),
	}
	resp, err := e.client.generateEmbedding(req)

	if err != nil {
		return nil, err
	}

	outputStruct := embedOutput{
		Embedding: resp.EmbeddingsFloats.Embeddings[0],
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
	return &output, nil
}

func (e *execution) taskRerank(in *structpb.Struct) (*structpb.Struct, error) {
	query := in.Fields["query"].GetStringValue()
	documents := []*cohereSDK.RerankRequestDocumentsItem{}
	for _, doc := range in.Fields["documents"].GetListValue().Values {
		document := cohereSDK.RerankRequestDocumentsItem{
			String: doc.GetStringValue(),
		}
		documents = append(documents, &document)
	}
	modelName := in.Fields["model-name"].GetStringValue()
	returnDocument := true
	rankFields := []string{"text"}
	req := cohereSDK.RerankRequest{
		Model:           &modelName,
		Query:           query,
		Documents:       documents,
		RankFields:      rankFields,
		ReturnDocuments: &returnDocument,
	}
	resp, err := e.client.generateRerank(req)
	if err != nil {
		return nil, err
	}
	newRanking := []string{}
	for _, rankResult := range resp.Results {
		newRanking = append(newRanking, rankResult.Document.Text)
	}
	outputStruct := rerankOutput{
		Ranking: newRanking,
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
	return &output, nil

}
