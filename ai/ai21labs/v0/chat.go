package ai21labs

// Source https://docs.ai21.com/reference/jamba-instruct-api on 2024-07-21

const chatEndpoint = "/studio/v1/chat/completions"

type ChatMessage struct {
	// enum: "user" "assistant" "system"
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature float32       `json:"temperature"`
	TopP        float32       `json:"top_p"`
	N           int           `json:"n"`
	// not supported
	// Stop   []string `json:"stop"` will trigger 500 internal server error
	Stream bool `json:"stream"`
}

type ChatChoice struct {
	Index   int         `json:"index"`
	Message ChatMessage `json:"message"`
	// could be "stop" or "length"
	FinishReason string `json:"finish_reason"`
}

type ChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatResponse struct {
	ID      string       `json:"id"`
	Choices []ChatChoice `json:"choices"`
	Usage   ChatUsage    `json:"usage"`
}

func (c *AI21labsClient) Chat(req ChatRequest) (ChatResponse, error) {
	resp := ChatResponse{}
	httpReq := c.httpClient.R().SetResult(&resp).SetBody(req)

	if _, err := httpReq.Post(chatEndpoint); err != nil {
		return resp, err
	}
	return resp, nil
}
