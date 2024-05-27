package slack

import "github.com/slack-go/slack"

type MockSlackClient struct{}

func (m *MockSlackClient) GetConversations(params *slack.GetConversationsParameters) ([]slack.Channel, string, error) {

	var channels []slack.Channel
	nextCursor := ""
	fakeChannel := slack.Channel{
		GroupConversation: slack.GroupConversation{
			Conversation: slack.Conversation{
				ID: "G0AKFJBEU",
			},
			Name: "test_channel",
		},
	}
	channels = append(channels, fakeChannel)

	return channels, nextCursor, nil
}

func (m *MockSlackClient) PostMessage(channelID string, options ...slack.MsgOption) (string, string, error) {

	return "", "", nil
}

func (m *MockSlackClient) GetConversationHistory(params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error) {

	fakeResp := slack.GetConversationHistoryResponse{
		SlackResponse: slack.SlackResponse{
			Ok: true,
		},
		Messages: []slack.Message{
			{
				Msg: slack.Msg{
					Timestamp:  "1715159446.644219",
					User:       "user123",
					Text:       "Hello, world!",
					ReplyCount: 1,
				},
			},
		},
	}

	return &fakeResp, nil
}

func (m *MockSlackClient) GetConversationReplies(params *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error) {

	fakeMessages := []slack.Message{
		{
			Msg: slack.Msg{
				Timestamp: "1715159446.644219",
				User:      "user123",
				Text:      "Hello, world!",
			},
		},
		{
			Msg: slack.Msg{
				Timestamp: "1715159449.399879",
				User:      "user456",
				Text:      "Hello, how are you",
			},
		},
	}
	hasMore := false
	nextCursor := ""
	return fakeMessages, hasMore, nextCursor, nil
}
