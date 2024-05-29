package slack

import (
	"fmt"
	"sync"
	"time"

	"github.com/instill-ai/component/base"
	"github.com/slack-go/slack"
	"google.golang.org/protobuf/types/known/structpb"
)

type UserInputReadTask struct {
	ChannelName     string `json:"channel_name"`
	StartToReadDate string `json:"start_to_read_date"`
}

type ReadTaskResp struct {
	Conversations []Conversation `json:"conversations"`
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

type ThreadReplyMessage struct {
	UserID   string `json:"user_id"`
	DateTime string `json:"datetime"`
	Message  string `json:"message"`
}

type UserInputWriteTask struct {
	ChannelName     string `json:"channel_name"`
	Message         string `json:"message"`
}

type WriteTaskResp struct {
	Result string `json:"result"`
}

func (e *execution) readMessage(in *structpb.Struct) (*structpb.Struct, error) {

	params := UserInputReadTask{}

	if err := base.ConvertFromStructpb(in, &params); err != nil {
		return nil, err
	}

	targetChannelID, err := loopChannelListAPI(e, params.ChannelName)

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
	err = setAPIRespToReadTaskResp(resp.Messages, &readTaskResp, params.StartToReadDate)
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
				replies, _ := getConversationReply(e, targetChannelID, readTaskResp.Conversations[idx].TS)
				// TODO: to be discussed about this error handdling
				// fail? or not fail?
				// if err != nil {
				// }

				// TODO: fetch further replies if there are

				mu.Lock()
				err := setRepliedToConversation(readTaskResp, replies, idx)
				// TODO: think a better way to pass lint, maybe use channel
				if err != nil {
					fmt.Println("error when set the output: ", err)
				}
				mu.Unlock()

			}(&readTaskResp, i)
		}
	}
	wg.Wait()

	if readTaskResp.Conversations == nil {
		readTaskResp.Conversations = []Conversation{}
	}

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

	targetChannelID, err := loopChannelListAPI(e, params.ChannelName)
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
