package whatsapp

import "google.golang.org/protobuf/types/known/structpb"

// this file is used to handle all send tasks except for send template message task
// types of messages: https://developers.facebook.com/docs/whatsapp/cloud-api/guides/send-messages

// Tasks:
// send text massage: https://developers.facebook.com/docs/whatsapp/cloud-api/messages/text-messages

// Objects (structs that are part of request struct; can be used in more than one task)

type languageObject struct {
	Code string `json:"code"`
}

type textObject struct {
	Body       string `json:"body"`
	PreviewURL bool   `json:"preview_url"`
}

type locationObject struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
	Name      string `json:"name"`
	Address   string `json:"address"`
}

type mediaObject struct {
	ID       string `json:"id,omitempty"`
	Link     string `json:"link,omitempty"`     // if id is used, no need link.
	Filename string `json:"filename,omitempty"` // only for document. This will specify the format of the file displayed in WhatsApp
	Caption  string `json:"caption,omitempty"`  // cannot be used in template message
}

type templateObject struct {
	Name       string            `json:"name"`
	Language   languageObject    `json:"language"`
	Components []componentObject `json:"components,omitempty"`
}

// Component type is either header, body or button
// Note:
// footer cannot have any parameters.
// header can have various parameters: text, image, location, document and video
// body support text parameters (along with currency and date time), but for our implementation, we will only do text parameter (for simplification). User can actually just input currency and date time as text as well, so it is not a big deal.
// button type is quick reply and call to action. Can either specify payload or text as the parameter for button.

type componentObject struct {
	Type          string        `json:"type"`
	Parameters    []interface{} `json:"parameters"`
	ButtonSubType string        `json:"sub_type,omitempty"` // only used when the type is button. Type of button
	ButtonIndex   string        `json:"index,omitempty"`    // only used when the type is button. Refers to the position index of the button
}

// structs that are part of API response (used in multiple task)

type contact struct {
	Input string `json:"input"`
	WaID  string `json:"wa_id"`
}

type message struct {
	ID            string `json:"id"`
	MessageStatus string `json:"message_status"`
}

// Send Text Message Task

type TaskSendTextMessageInput struct {
	PhoneNumberID string `json:"phone-number-id"`
	To            string `json:"to"`
	Body          string `json:"body"`
	PreviewURL    string `json:"preview-url"`
}

type TaskSendTextMessageReq struct {
	MessagingProduct string     `json:"messaging_product"`
	To               string     `json:"to"`
	Type             string     `json:"type"`
	Text             textObject `json:"text"`
}

type TaskSendTextMessageResp struct {
	MessagingProduct string    `json:"messaging_product"`
	Contacts         []contact `json:"contacts"`
	Messages         []message `json:"messages"`
}

// Note: no message status in the output struct because it is not returned in the response

type TaskSendTextMessageOutput struct {
	WaID string `json:"recipient-wa-id"`
	ID   string `json:"message-id"`
}

func (e *execution) SendTextMessage(in *structpb.Struct) (*structpb.Struct, error) {
	return nil, nil
}
