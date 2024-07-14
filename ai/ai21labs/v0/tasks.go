package ai21labs

import (
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

// pricing info: https://www.ai21.com/pricing
// note: task specific models are billed based on API calls, not generated tokens

type TaskTextGenerationChatInput struct {
	base.TemplateTextGenerationInput
	TopP float64 `json:"top-p"`
	N    int     `json:"n"`
}

type TaskTextGenerationChatOutput struct {
	base.TemplateTextGenerationOutput
}

func (e *execution) TaskTextGenerationChat(in *structpb.Struct) (*structpb.Struct, error) {
	input := TaskTextGenerationChatInput{}
	if err := base.ConvertFromStructpb(in, &input); err != nil {
		return nil, err
	}

	messages := []ChatMessage{}

	messages = append(messages, ChatMessage{
		Role:    "system",
		Content: input.SystemMsg,
	})

	for _, message := range input.ChatHistory {
		contents := ""
		for _, content := range message.Content {
			if content.Type == "text" {
				contents += content.Text
			}
		}
		messages = append(messages, ChatMessage{
			Role:    message.Role,
			Content: contents,
		})
	}

	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: input.Prompt,
	})

	req := ChatRequest{
		Model:       input.ModelName,
		Messages:    messages,
		MaxTokens:   input.MaxNewTokens,
		Temperature: float32(input.Temperature),
		TopP:        float32(input.TopP),
		N:           input.N,
	}

	resp, err := e.client.Chat(req)
	if err != nil {
		return nil, err
	}

	outputStruct := base.TemplateTextGenerationOutput{
		Text: resp.Choices[0].Message.Content,
		Usage: base.GenerativeTextModelUsage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
		},
	}

	return base.ConvertToStructpb(outputStruct)
}

type TaskContextualAnsweringInput struct {
	ContextualAnswersRequest
}

type TaskContextualAnsweringOutput struct {
	Answer          string `json:"answer"`
	AnswerInContext bool   `json:"answer-in-context"`
}

func (e *execution) TaskContextualAnswering(in *structpb.Struct) (*structpb.Struct, error) {
	input := TaskContextualAnsweringInput{}
	if err := base.ConvertFromStructpb(in, &input); err != nil {
		return nil, err
	}

	resp, err := e.client.ContextualAnswers(input.ContextualAnswersRequest)
	if err != nil {
		return nil, err
	}

	output := TaskContextualAnsweringOutput{
		Answer:          resp.Answer,
		AnswerInContext: resp.AnswerInContext,
	}

	return base.ConvertToStructpb(output)
}

type TaskTextEmbeddingsInput struct {
	Text  string `json:"text"`
	Style string `json:"style"`
}

type TaskTextEmbeddingsOutput struct {
	Embedding []float32                    `json:"embedding"`
	Usage     base.EmbeddingTextModelUsage `json:"usage"`
}

func (e *execution) TaskTextEmbeddings(in *structpb.Struct) (*structpb.Struct, error) {
	input := TaskTextEmbeddingsInput{}
	if err := base.ConvertFromStructpb(in, &input); err != nil {
		return nil, err
	}

	req := EmbeddingsRequest{
		Texts: []string{input.Text},
		Type:  EmbeddingsType(input.Style),
	}

	resp, err := e.client.Embeddings(req)
	if err != nil {
		return nil, err
	}

	output := TaskTextEmbeddingsOutput{
		Embedding: resp.Results[0].Embedding,
		Usage: base.EmbeddingTextModelUsage{
			Tokens: len(input.Text) / 2, // IMPORTANT: this is a rough estimate, but the embedding API does not return token counts for now (2024-07-21)
		},
	}
	return base.ConvertToStructpb(output)
}

type TaskTextImprovementInput struct {
	Text string `json:"text"`
}

type TaskTextImprovementOutput struct {
	Suggestions []string `json:"suggestions"`
	StartIndexs []int    `json:"start-indexs"`
	EndIndexs   []int    `json:"end-indexs"`
	Types       []string `json:"types"`
}

func (e *execution) TaskTextImprovement(in *structpb.Struct) (*structpb.Struct, error) {
	input := TaskTextImprovementInput{}
	if err := base.ConvertFromStructpb(in, &input); err != nil {
		return nil, err
	}

	// Default to all improvement types
	ImprovementTypes := []ImprovementType{
		Fluency,
		Specificity,
		Variety,
		ShortSentences,
		Conciseness,
	}

	req := TextImprovementsRequest{
		Text:  input.Text,
		Types: ImprovementTypes,
	}

	resp, err := e.client.TextImprovements(req)
	if err != nil {
		return nil, err
	}

	output := TaskTextImprovementOutput{}

	for _, improvement := range resp.Improvements {
		output.Suggestions = append(output.Suggestions, improvement.Suggestions...)
		output.StartIndexs = append(output.StartIndexs, improvement.StartIndex)
		output.EndIndexs = append(output.EndIndexs, improvement.EndIndex)
		output.Types = append(output.Types, string(improvement.ImprovementType))
	}
	return base.ConvertToStructpb(output)
}

type TaskTextParaphrasingInput struct {
	Text       string `json:"text"`
	Style      string `json:"style"`
	StartIndex int    `json:"start-index"`
	EndIndex   int    `json:"end-index"`
}

type TaskTextParaphrasingOutput struct {
	Suggestions []string `json:"suggestions"`
}

func (e *execution) TaskTextParaphrasing(in *structpb.Struct) (*structpb.Struct, error) {
	input := TaskTextParaphrasingInput{}
	if err := base.ConvertFromStructpb(in, &input); err != nil {
		return nil, err
	}

	req := ParaphraseRequest{
		Text:       input.Text,
		Style:      ParaphraseStyle(input.Style),
		StartIndex: input.StartIndex,
		EndIndex:   input.EndIndex,
	}

	resp, err := e.client.Paraphrase(req)
	if err != nil {
		return nil, err
	}

	output := TaskTextParaphrasingOutput{
		Suggestions: []string{},
	}

	for _, suggestion := range resp.Suggestions {
		output.Suggestions = append(output.Suggestions, suggestion.Text)
	}
	return base.ConvertToStructpb(output)
}

type TaskTextSummarizationInput struct {
	Text   string `json:"text"`
	Focus  string `json:"focus"`
	Source string `json:"source"`
}

type TaskTextSummarizationOutput struct {
	Summary string `json:"summary"`
}

func (e *execution) TaskTextSummarization(in *structpb.Struct) (*structpb.Struct, error) {
	input := TaskTextSummarizationInput{}
	if err := base.ConvertFromStructpb(in, &input); err != nil {
		return nil, err
	}

	req := SummarizeRequest{
		Source:     input.Text,
		SourceType: SourceType(input.Source),
		Focus:      input.Focus,
	}

	resp, err := e.client.Summarize(req)
	if err != nil {
		return nil, err
	}

	output := TaskTextSummarizationOutput{
		Summary: resp.Summary,
	}

	return base.ConvertToStructpb(output)
}

type TaskTextSummarizationBySegmentInput struct {
	Text   string `json:"text"`
	Focus  string `json:"focus"`
	Source string `json:"source"`
}

type TaskTextSummarizationBySegmentOutput struct {
	Summerizations []string `json:"summerizations"`
	SegmentTexts   []string `json:"segments"`
	SegmentHtmls   []string `json:"segment-htmls"`
	Types          []string `json:"types"`
}

func (e *execution) TaskTextSummarizationBySegment(in *structpb.Struct) (*structpb.Struct, error) {
	input := TaskTextSummarizationBySegmentInput{}
	if err := base.ConvertFromStructpb(in, &input); err != nil {
		return nil, err
	}

	req := SummarizeRequest{
		Source:     input.Text,
		SourceType: SourceType(input.Source),
		Focus:      input.Focus,
	}

	resp, err := e.client.SummarizeBySegment(req)
	if err != nil {
		return nil, err
	}

	output := TaskTextSummarizationBySegmentOutput{
		Summerizations: []string{},
		SegmentTexts:   []string{},
		SegmentHtmls:   []string{},
		Types:          []string{},
	}

	for _, segment := range resp.Segements {
		output.Summerizations = append(output.Summerizations, segment.Summary)
		output.SegmentTexts = append(output.SegmentTexts, segment.SegmentText)
		output.SegmentHtmls = append(output.SegmentHtmls, segment.SegmentHTML)
		output.Types = append(output.Types, string(segment.SegmentType))
	}

	return base.ConvertToStructpb(output)
}

type TaskTextSegmentationInput struct {
	Text   string `json:"text"`
	Source string `json:"source"`
}

type TaskTextSegmentationOutput struct {
	SegmentTexts []string `json:"segments"`
	Types        []string `json:"types"`
}

func (e *execution) TaskTextSegmentation(in *structpb.Struct) (*structpb.Struct, error) {
	input := TaskTextSegmentationInput{}
	if err := base.ConvertFromStructpb(in, &input); err != nil {
		return nil, err
	}

	req := TextSegmentationRequest{
		Source:     input.Text,
		SourceType: SourceType(input.Source),
	}

	resp, err := e.client.TextSegmentation(req)
	if err != nil {
		return nil, err
	}

	output := TaskTextSegmentationOutput{
		SegmentTexts: []string{},
		Types:        []string{},
	}

	for _, segment := range resp.Segments {
		output.SegmentTexts = append(output.SegmentTexts, segment.SegementText)
		output.Types = append(output.Types, string(segment.SegmentType))
	}

	return base.ConvertToStructpb(output)
}

type TaskGrammarCheckInput struct {
	GrammaticalErrorCorrectionsRequest
}

type TaskGrammarCheckOutput struct {
	Suggestions []string `json:"suggestions"`
	StartIndexs []int    `json:"start-indexs"`
	EndIndexs   []int    `json:"end-indexs"`
	Types       []string `json:"types"`
}

func (e *execution) TaskGrammarCheck(in *structpb.Struct) (*structpb.Struct, error) {
	input := TaskGrammarCheckInput{}
	if err := base.ConvertFromStructpb(in, &input); err != nil {
		return nil, err
	}

	resp, err := e.client.GrammaticalErrorCorrections(input.GrammaticalErrorCorrectionsRequest)
	if err != nil {
		return nil, err
	}
	output := TaskGrammarCheckOutput{}
	for _, correction := range resp.Corrections {
		output.Suggestions = append(output.Suggestions, correction.Suggestion)
		output.StartIndexs = append(output.StartIndexs, correction.StartIndex)
		output.EndIndexs = append(output.EndIndexs, correction.EndIndex)
		output.Types = append(output.Types, string(correction.CorrectionType))
	}
	return base.ConvertToStructpb(output)
}
