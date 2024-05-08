package slack

import (
	"sync"
	"time"

	"github.com/instill-ai/component/pkg/base"
	"github.com/slack-go/slack"
	"google.golang.org/protobuf/types/known/structpb"
)

func (e *execution) readMessage(in *structpb.Struct) (*structpb.Struct, error) {

	params := UserInputReadTask{}

	if err := base.ConvertFromStructpb(in, &params); err != nil {
		return nil, err
	}
	var targetChannelID string
	err := loopChannelListApi(e, params.IsPublicChannel, params.ChannelName, &targetChannelID)

	if err != nil {
		return nil, err
	}

	resp, err := getConversationHistory(e, targetChannelID, "")
	if err != nil {
		return nil, err
	}

	// TODO: discussed if only collect X days ago as default.
	if params.StartToReadDate == "" {
		currentTime := time.Now()
		sevenDaysAgo := currentTime.AddDate(0, 0, -7)
		sevenDaysAgoString := sevenDaysAgo.Format("2006-01-02")
		params.StartToReadDate = sevenDaysAgoString
	}

	var readTaskResp ReadTaskResp
	err = setApiRespToReadTaskResp(resp.Messages, &readTaskResp, params.StartToReadDate)
	if err != nil {
		return nil, err
	}

	// TODO: fetch historyAPI first if there are more conversations.
	// if resp.ResponseMetaData.NextCursor != "" {

	// }

	var mu sync.Mutex
	var wg sync.WaitGroup

	for i, conversation := range readTaskResp.Conversations {
		if conversation.ReplyCount > 0 {
			wg.Add(1)
			go func(readTaskResp *ReadTaskResp, idx int) {
				defer wg.Done()
				replies, _ := getConversationReply(e, targetChannelID, readTaskResp.Conversations[idx].Ts)
				// TODO: to be discussed about this error handdling
				// fail? or not fail?
				// if err != nil {
				// }

				// TODO: fetch further replies if there are

				mu.Lock()
				setRepliedToConversation(readTaskResp, replies, idx)
				mu.Unlock()

			}(&readTaskResp, i)
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

	var targetChannelID string
	err := loopChannelListApi(e, params.IsPublicChannel, params.ChannelName, &targetChannelID)
	if err != nil {
		return nil, err
	}

	_, _, err = e.client.PostMessage(targetChannelID, slack.MsgOptionText(params.Message, false))

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
