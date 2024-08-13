package ai21labs

import (
	"github.com/instill-ai/component/ai"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

// pricing info: https://www.ai21.com/pricing
// note: task specific models are billed based on API calls, not generated tokens

const (
	TaskTextGenerationChat         = "TASK_TEXT_GENERATION_CHAT"
	TaskContextualAnswering        = "TASK_CONTEXTUAL_ANSWERING"
	TaskTextEmbeddings             = "TASK_TEXT_EMBEDDINGS"
	TaskTextImprovement            = "TASK_TEXT_IMPROVEMENT"
	TaskTextParaphrasing           = "TASK_TEXT_PARAPHRASING"
	TaskTextSummarization          = "TASK_TEXT_SUMMARIZATION"
	TaskTextSummarizationBySegment = "TASK_TEXT_SUMMARIZATION_SEGMENT"
	TaskTextSegmentation           = "TASK_TEXT_SEGMENTATION"
	TaskGrammarCheck               = "TASK_GRAMMAR_CHECK"
)

type TaskTextGenerationChatInput struct {
	ai.TemplateTextGenerationInput
	TopP float64 `json:"top-p"`
}

type TaskTextGenerationChatOutput struct {
	ai.TemplateTextGenerationOutput
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
		N:           1,
	}

	resp, err := e.client.Chat(req)
	if err != nil {
		return nil, err
	}

	outputStruct := ai.TemplateTextGenerationOutput{
		Text: resp.Choices[0].Message.Content,
		Usage: ai.GenerativeTextModelUsage{
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
	Embedding []float32                  `json:"embedding"`
	Usage     ai.EmbeddingTextModelUsage `json:"usage"`
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
		Usage: ai.EmbeddingTextModelUsage{
			Tokens: len(input.Text) / 2, // IMPORTANT: this is a rough estimate, but the embedding API does not return token counts for now (2024-07-21)
		},
	}
	return base.ConvertToStructpb(output)
}

type TaskTextImprovementInput struct {
	Text string `json:"text"`
}

type Improvement struct {
	Texts      []string `json:"texts"`
	StartIndex int      `json:"start-index"`
	EndIndex   int      `json:"end-index"`
	Type       string   `json:"type"`
}

type TaskTextImprovementOutput struct {
	Suggestions []Improvement `json:"suggestions"`
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
		output.Suggestions = append(output.Suggestions, Improvement{
			Texts:      improvement.Suggestions,
			StartIndex: improvement.StartIndex,
			EndIndex:   improvement.EndIndex,
			Type:       string(improvement.ImprovementType),
		})
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

type TextSegmentSummarization struct {
	Summary string `json:"sammery"`
	Text    string `json:"text"`
	HTML    string `json:"html"`
	Type    string `json:"type"`
}

type TaskTextSummarizationBySegmentOutput struct {
	Summerizations []TextSegmentSummarization `json:"summerizations"`
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

	output := TaskTextSummarizationBySegmentOutput{}

	for _, segment := range resp.Segements {
		output.Summerizations = append(output.Summerizations, TextSegmentSummarization{
			Summary: segment.Summary,
			Text:    segment.SegmentText,
			HTML:    segment.SegmentHTML,
			Type:    string(segment.SegmentType),
		})
	}

	return base.ConvertToStructpb(output)
}

type TaskTextSegmentationInput struct {
	Text   string `json:"text"`
	Source string `json:"source"`
}

type TextSegment struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type TaskTextSegmentationOutput struct {
	Segments []TextSegment `json:"segments"`
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

	output := TaskTextSegmentationOutput{}

	for _, segment := range resp.Segments {
		output.Segments = append(output.Segments, TextSegment{
			Text: segment.SegementText,
			Type: string(segment.SegmentType),
		})
	}

	return base.ConvertToStructpb(output)
}

type TaskGrammarCheckInput struct {
	GrammaticalErrorCorrectionsRequest
}

type GrammerSuggestion struct {
	Text       string `json:"text"`
	StartIndex int    `json:"start-index"`
	EndIndex   int    `json:"end-index"`
	Type       string `json:"type"`
}

type TaskGrammarCheckOutput struct {
	Suggestions []GrammerSuggestion `json:"suggestions"`
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
		output.Suggestions = append(output.Suggestions, GrammerSuggestion{
			Text:       correction.Suggestion,
			StartIndex: correction.StartIndex,
			EndIndex:   correction.EndIndex,
			Type:       string(correction.CorrectionType),
		})
	}
	return base.ConvertToStructpb(output)
}
