package slack

import "reflect"

// Todo: make it multiple options
func setChannelType(params *ConversationsListParams, isPublicChannel bool) {
	if !isPublicChannel {
		params.Types = "private_channel"
	} else {
		params.Types = "public_channel"
	}
}

func getChannelId(channelName string, channels []SlackChannel) string {
	for _, slackChannel := range channels {
		if slackChannel.Name == channelName {
			return slackChannel.ID
		}
	}
	return ""
}

func setGetParams(params any) map[string]string {

	v := reflect.ValueOf(params)
	typ := v.Type()

	keyValueMap := make(map[string]string)

	// TODO: make it extendable
	for i := 0; i < v.NumField(); i++ {
		if typ.Field(i).Name == "Types" {
			keyValueMap["types"] = v.Field(i).String()

		} else if typ.Field(i).Name == "ChannelID" {
			keyValueMap["channel"] = v.Field(i).String()
		} else if typ.Field(i).Name == "ThreadTs" {
			keyValueMap["ts"] = v.Field(i).String()
		}
	}
	return keyValueMap
}

func appendToReadTaskResp(resp ConversationReplyApiResp, readTaskResp *ReadTaskResp) {
	
}
