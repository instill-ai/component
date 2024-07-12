package ai21labs

import (
	"github.com/instill-ai/component/internal/util/httpclient"
	"go.uber.org/zap"
)

type AI21labsClient struct {
	httpClient *httpclient.Client
}

type AI21labsClientInterface interface {
	Chat(ChatRequest) (ChatResponse, error)
	Embeddings(EmbeddingsRequest) (EmbeddingsResponse, error)
	ContextualAnswers(ContextualAnswersRequest) (ContextualAnswersResponse, error)
	GrammaticalErrorCorrections(GrammaticalErrorCorrectionsRequest) (GrammaticalErrorCorrectionsResponse, error)
	Paraphrase(ParaphraseRequest) (ParaphraseResponse, error)
	Summarize(SummarizeRequest) (SummarizeResponse, error)
	SummarizeBySegment(SummarizeRequest) (SummerizeBySegmentResponse, error)
	TextImprovements(TextImprovementsRequest) (TextImprovementsResponse, error)
	TextSegmentation(TextSegmentationRequest) (TextSegmentationResponse, error)
	BaseURL() string
}

func newClient(apiKey string, baseURL string, logger *zap.Logger) *AI21labsClient {
	c := httpclient.New("AI21labs", baseURL,
		httpclient.WithLogger(logger),
	)
	c.SetAuthToken(apiKey)

	return &AI21labsClient{httpClient: c}
}

func (c *AI21labsClient) BaseURL() string {
	return c.httpClient.BaseURL
}

// common types

type SourceType string

const (
	PlainText SourceType = "TEXT"
	Link      SourceType = "URL"
)

type SegmentType string

const (
	NormalText      SegmentType = "normal_text"
	Title           SegmentType = "title"
	Header1         SegmentType = "h1"
	Header2         SegmentType = "h2"
	Header3         SegmentType = "h3"
	Header4         SegmentType = "h4"
	Header5         SegmentType = "h5"
	Footnote        SegmentType = "footnote"
	Other           SegmentType = "other"
	NormalTextShort SegmentType = "normal_text_short"
	NormalTextLong  SegmentType = "normal_text_long"
	NonEnglish      SegmentType = "non_english"
)
