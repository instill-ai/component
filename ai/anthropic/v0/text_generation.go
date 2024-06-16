package anthropic

// reference: https://docs.anthropic.com/en/api/messages
type messagesResp struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Role       string    `json:"role"`
	Content    []content `json:"content"`
	Model      string    `json:"model"`
	StopReason string    `json:"stop_reason,omitempty"`
	Usage      usage     `json:"usage"`
}

type messagesReq struct {
	Model         string      `json:"model"`
	Messages      []message   `json:"messages"`
	MaxTokens     int         `json:"max_tokens"`
	Metadata      interface{} `json:"metadata"`
	StopSequences []string    `json:"stop_sequences,omitempty"`
	Stream        bool        `json:"stream,omitempty"`
	System        string      `json:"system,omitempty"`
	Temperature   float32     `json:"temperature,omitempty"`
	TopK          int         `json:"top_k,omitempty"`
	TopP          float32     `json:"top_p,omitempty"`
}

type messagesOutput struct {
	Text string `json:"text"`
}

type message struct {
	Role    string    `json:"role"`
	Content []content `json:"content"`
}

type usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// doesn't support anthropic tools at the moment
type content struct {
	Type   string  `json:"type"`
	Text   string  `json:"text,omitempty"`
	Source *source `json:"source,omitempty"`
}

type source struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}
