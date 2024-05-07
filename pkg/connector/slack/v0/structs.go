package slack

type UserInputWriteTask struct {
	ChannelName     string `json:"channel_name"`
	Message         string `json:"message"`
	IsPublicChannel bool   `json:"is_public_channel"`
}
// 	OK      bool   `json:"ok"`
// 	Channel string `json:"channel"`
// 	Ts      string `json:"ts"`
// 	Error   string `json:"error"`
// 	Needed  string `json:"needed"`
// 	Message struct {
// 		User       string `json:"user"`
// 		Type       string `json:"type"`
// 		Ts         string `json:"ts"`
// 		BotID      string `json:"bot_id"`
// 		AppID      string `json:"app_id"`
// 		Text       string `json:"text"`
// 		Team       string `json:"team"`
// 		BotProfile struct {
// 			ID    string `json:"id"`
// 			AppID string `json:"app_id"`
// 			Name  string `json:"name"`
// 			Icons struct {
// 				Image36 string `json:"image_36"`
// 				Image48 string `json:"image_48"`
// 				Image72 string `json:"image_72"`
// 			} `json:"icons"`
// 			Deleted bool   `json:"deleted"`
// 			Updated int    `json:"updated"`
// 			TeamID  string `json:"team_id"`
// 		} `json:"bot_profile"`
// 	} `json:"message"`
// 	Blocks []struct {
// 		Type     string `json:"type"`
// 		BlockID  string `json:"block_id"`
// 		Elements []struct {
// 			Type     string `json:"type"`
// 			Elements []struct {
// 				Type string `json:"type"`
// 				Text string `json:"text"`
// 			} `json:"elements"`
// 		} `json:"elements"`
// 	} `json:"blocks"`
// }

type WriteTaskResp struct {
	Result string `json:"result"`
}

// TODO: Read Task
type UserInputReadTask struct {
	ChannelName     string `json:"channel_name"`
	StartToReadDate string `json:"start_to_read_date"`
	IsPublicChannel bool   `json:"is_public_channel"`
}

type ConversationsHistoryParams struct {
	ChannelID string `json:"channel,omitempty"`
	Cursor    string `json:"cursor,omitempty"`
	Inclusive bool   `json:"inclusive,omitempty"`
	Latest    string `json:"latest,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Oldest    string `json:"oldest,omitempty"`
}

type ConversationsHistoryApiResp struct {
	Ok               bool      `json:"ok"`
	Error            string    `json:"error"`
	Messages         []Message `json:"messages"`
	HasMore          bool      `json:"has_more"`
	PinCount         int       `json:"pin_count"`
	ResponseMetadata struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata"`
}

type MessageAttachment struct {
	ServiceName string `json:"service_name"`
	Text        string `json:"text"`
	Fallback    string `json:"fallback"`
	ThumbURL    string `json:"thumb_url"`
	ThumbWidth  int    `json:"thumb_width"`
	ThumbHeight int    `json:"thumb_height"`
	ID          int    `json:"id"`
}

type Message struct {
	Type            string              `json:"type"`
	User            string              `json:"user"`
	Text            string              `json:"text"`
	Attachments     []MessageAttachment `json:"attachments,omitempty"`
	Ts              string              `json:"ts"`
	ThreadTs        string              `json:"thread_ts,omitempty"`
	ReplyCount      int                 `json:"reply_count,omitempty"`
	ReplyUsersCount int                 `json:"reply_users_count,omitempty"`
	LatestReply     string              `json:"latest_reply,omitempty"`
	ReplyUsers      []string            `json:"reply_users,omitempty"`
}

type ConversationReplyParams struct {
	ChannelID string `json:"channel"`
	ThreadTs  string `json:"ts"`
	Cursor    string `json:"cursor,omitempty"`
	Inclusive bool   `json:"inclusive,omitempty"`
	Latest    string `json:"latest,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Oldest    string `json:"oldest,omitempty"`
}

type ConversationReplyApiResp struct {
	Ok               bool             `json:"ok"`
	Error            string           `json:"error"`
	Messages         []ReplyMessage   `json:"messages"`
	HasMore          bool             `json:"has_more"`
	ResponseMetadata ResponseMetadata `json:"response_metadata"`
}

type ReplyMessage struct {
	Type         string `json:"type"`
	User         string `json:"user"`
	Text         string `json:"text"`
	ThreadTs     string `json:"thread_ts"`
	ReplyCount   int    `json:"reply_count"`
	Subscribed   bool   `json:"subscribed"`
	LastRead     string `json:"last_read"`
	UnreadCount  int    `json:"unread_count"`
	Ts           string `json:"ts"`
	ParentUserID string `json:"parent_user_id,omitempty"`
}

type ResponseMetadata struct {
	NextCursor string `json:"next_cursor"`
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
	ThreadReplyMessage []ThreadReplyMessage `json:"thread_reply_messages"`
}
