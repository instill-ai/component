package slack

import (
	"github.com/slack-go/slack"
)

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

// TODO: Read Task
// func setGetParams(params any) map[string]string {

// 	v := reflect.ValueOf(params)
// 	typ := v.Type()

// 	keyValueMap := make(map[string]string)

// 	// TODO: make it extendable
// 	for i := 0; i < v.NumField(); i++ {
// 		if typ.Field(i).Name == "Types" {
// 			keyValueMap["types"] = v.Field(i).String()

// 		} else if typ.Field(i).Name == "ChannelID" {
// 			keyValueMap["channel"] = v.Field(i).String()
// 		} else if typ.Field(i).Name == "ThreadTs" {
// 			keyValueMap["ts"] = v.Field(i).String()
// 		}
// 	}
// 	return keyValueMap
// }

// func appendToReadTaskResp(resp ConversationReplyApiResp, readTaskResp *ReadTaskResp) {

// }
