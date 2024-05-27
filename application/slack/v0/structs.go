package slack

type UserInputWriteTask struct {
	ChannelName     string `json:"channel_name"`
	Message         string `json:"message"`
	IsPublicChannel bool   `json:"is_public_channel"`
}

type WriteTaskResp struct {
	Result string `json:"result"`
}

// TODO: Read Task
type UserInputReadTask struct {
	ChannelName     string `json:"channel_name"`
	StartToReadDate string `json:"start_to_read_date"`
	IsPublicChannel bool   `json:"is_public_channel"`
}

type ReadTaskResp struct {
	Conversations []Conversation `json:"conversations"`
}

type ThreadReplyMessage struct {
	UserID   string `json:"user_id"`
	DateTime string `json:"datetime"`
	Message  string `json:"message"`
}

type Conversation struct {
	UserID             string               `json:"user_id"`
	Message            string               `json:"message"`
	StartDate          string               `json:"start_date"`
	LastDate           string               `json:"last_date"`
	TS                 string               `json:"ts"`
	ReplyCount         int                  `json:"reply_count"`
	ThreadReplyMessage []ThreadReplyMessage `json:"thread_reply_messages"`
}
