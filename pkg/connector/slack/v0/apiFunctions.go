package slack

import (
	"fmt"
	"strconv"
	"time"

	"github.com/slack-go/slack"
)

func loopChannelListApi(e *execution, isPublic bool, channelName string, targetChannelID *string) error {
	var apiParams slack.GetConversationsParameters
	setChannelType(&apiParams, isPublic)

	for {

		slackChannels, nextCur, err := e.client.GetConversations(&apiParams)
		if err != nil {
			return err
		}

		setChannelId(channelName, slackChannels, targetChannelID)

		if *targetChannelID != "" {
			break
		}

		if *targetChannelID == "" && nextCur == "" {
			err := fmt.Errorf("there is no match name in slack channel [%v]", channelName)
			return err
		}

		apiParams.Cursor = nextCur

	}

	return nil
}

// Todo: make it multiple options
func setChannelType(params *slack.GetConversationsParameters, isPublicChannel bool) {
	if !isPublicChannel {
		params.Types = append(params.Types, "private_channel")
	} else {
		params.Types = append(params.Types, "public_channel")
	}
}

func setChannelId(channelName string, channels []slack.Channel, targetChannelID *string) {
	for _, slackChannel := range channels {
		if channelName == slackChannel.Name {
			*targetChannelID = slackChannel.ID
			break
		}
	}
}

func getConversationHistory(e *execution, channelID string, nextCur string) (*slack.GetConversationHistoryResponse, error) {
	apiHistoryParams := slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Cursor:    nextCur,
	}

	historiesResp, err := e.client.GetConversationHistory(&apiHistoryParams)
	if err != nil {
		return nil, err
	}
	if !historiesResp.Ok {
		err := fmt.Errorf("slack api error: %v", historiesResp.Error)
		return nil, err
	}

	return historiesResp, nil
}

func getConversationReply(e *execution, channelID string, ts string) ([]slack.Message, error) {
	apiParams := slack.GetConversationRepliesParameters{
		ChannelID: channelID,
		Timestamp: ts,
	}
	msgs, _, _, err := e.client.GetConversationReplies(&apiParams)

	if err != nil {
		return nil, err
	}

	// TODO: deal with the nextCur
	// add one more input, nextCur string, in function
	// if nextCur != "" {

	// }

	return msgs, nil
}

func setApiRespToReadTaskResp(apiResp []slack.Message, readTaskResp *ReadTaskResp) error {

	for _, msg := range apiResp {
		formatedDate, err := transformTsToDate(msg.Timestamp, "2006-01-02")

		if err != nil {
			return err
		}

		conversation := Conversation{
			UserID:     msg.User,
			Message:    msg.Text,
			StartDate:  formatedDate,
			ReplyCount: msg.ReplyCount,
			Ts:         msg.Timestamp,
		}

		conversation.ThreadReplyMessage = []ThreadReplyMessage{}
		readTaskResp.Conversations = append(readTaskResp.Conversations, conversation)
	}
	return nil
}

func setRepliedToConversation(c *Conversation, replies []slack.Message) error {
	for _, msg := range replies {

		if c.Ts == msg.Timestamp {
			continue
		}

		formatedDateTime, err := transformTsToDate(msg.Timestamp, "2006-01-02 15:04:05")
		if err != nil {
			return err
		}
		reply := ThreadReplyMessage{
			UserID:   msg.User,
			DateTime: formatedDateTime,
			Message:  msg.Text,
		}
		c.ThreadReplyMessage = append(c.ThreadReplyMessage, reply)
	}
	return nil
}

func transformTsToDate(ts string, format string) (string, error) {

	tsFloat, err := strconv.ParseFloat(ts, 64)
	if err != nil {
		return "", err
	}

	timestamp := time.Unix(int64(tsFloat), int64((tsFloat-float64(int64(tsFloat)))*1e9))

	formatedTs := timestamp.Format(format)
	return formatedTs, nil
}
