package slack

import (
	"fmt"
	"sync"

	"github.com/instill-ai/component/pkg/base"
	"google.golang.org/protobuf/types/known/structpb"
)

func (e *execution) readMessage(in *structpb.Struct) (*structpb.Struct, error) {

	params := UserInputReadTask{}

	if err := base.ConvertFromStructpb(in, &params); err != nil {
		return nil, err
	}

	var apiListParams ConversationsListParams

	setChannelType(&apiListParams, params.IsPublicChannel)

	slackChannels, err := fetchChannelInfo(e.client, apiListParams)
	if err != nil {
		return nil, err
	}

	channelId := getChannelId(params.ChannelName, *slackChannels)

	if channelId == "" {
		err := fmt.Errorf("there is no match name in slack channel [%v]", params.ChannelName)
		return nil, err
	}

	apiHistoryParams := ConversationsHistoryParams{
		ChannelID: channelId,
	}

	conversations, err := fetchConversations(e.client, apiHistoryParams)
	if err != nil {
		return nil, err
	}

	// TODO: new thread to fetch next conversations
	if conversations.ResponseMetadata.NextCursor != "" {
	}

	// TODO: fetch historyAPI first if there are more conversations.
	if conversations.HasMore {
	}

	var responses []ConversationReplyApiResp
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(conversations.Messages))

	var readTaskResp ReadTaskResp

	for _, message := range conversations.Messages {
		if message.ReplyCount > 0 {
			go func(message Message) {
				defer wg.Done()
				apiReplyParams := ConversationReplyParams{
					ThreadTs:  message.Ts,
					ChannelID: channelId,
				}
				replies, err := fetchReplies(e.client, apiReplyParams)

				// TODO: to be discussed about this error handdling
				// fail? or not fail?
				if err != nil {
					// fmt.Printf("replies %v:  %v \n", i, replies)
				}

				// TODO: fetch further replies if there are

				// !!! TODO next week: 
				// 1. transform replies to ThreadReplyMessage
				// 2. Get the data from message
				// 2-1. Transform ts to timestamp and compare with start-date
				// 3. Make it as a conversation
				// 4. lock and append and unlock
				mu.Lock()
				responses = append(responses, *replies)
				mu.Unlock()

			}(message)
		}
	}
	wg.Wait()

	out, err := base.ConvertToStructpb(readTaskResp)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (e *execution) sendMessage(in *structpb.Struct) (*structpb.Struct, error) {
	params := UserInputWriteTask{}

	if err := base.ConvertFromStructpb(in, &params); err != nil {
		return nil, err
	}

	var apiParams ConversationsListParams

	setChannelType(&apiParams, params.IsPublicChannel)

	slackChannels, err := fetchChannelInfo(e.client, apiParams)
	if err != nil {
		return nil, err
	}

	sendingData := ChatPostMessageParams{
		Text:    params.Message,
		Channel: getChannelId(params.ChannelName, *slackChannels),
	}

	if sendingData.Channel == "" {
		err := fmt.Errorf("there is no match name in slack channel [%v]", params.ChannelName)
		return nil, err
	}

	err = postMessageToSlackChannel(e.client, sendingData)
	if err != nil {
		return nil, err
	}

	out, err := base.ConvertToStructpb(WriteTaskResp{
		Result: "succeed",
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}
