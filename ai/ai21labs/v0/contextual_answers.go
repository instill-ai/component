package ai21labs

//source https://docs.ai21.com/reference/contextual-answers-ref on 2024-07-21

const contextualAnswersEndpoint = "/studio/v1/answer"

type ContextualAnswersRequest struct {
	Context  string `json:"context"`
	Question string `json:"question"`
}

type ContextualAnswersResponse struct {
	ID              string `json:"id"`
	Answer          string `json:"answer"`
	AnswerInContext bool   `json:"answerInContext"`
}

func (c *AI21labsClient) ContextualAnswers(req ContextualAnswersRequest) (ContextualAnswersResponse, error) {
	resp := ContextualAnswersResponse{}
	httpReq := c.httpClient.R().SetResult(&resp).SetBody(req)
	if _, err := httpReq.Post(contextualAnswersEndpoint); err != nil {
		return resp, err
	}
	return resp, nil
}
