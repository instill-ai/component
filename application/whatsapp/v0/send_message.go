package whatsapp

import (
	"fmt"
	"strings"

	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

// this file is used to handle 4 send tasks

// Tasks:
// 1. Send Text Message
// 2. Send Media Message
// 3. Send Location Message
// 4. Send Contact Message

// types of messages: https://developers.facebook.com/docs/whatsapp/cloud-api/guides/send-messages

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

type contactsObject struct {
	Name     nameObject    `json:"name"`
	Phones   []phoneObject `json:"phones,omitempty"`
	Emails   []emailObject `json:"emails,omitempty"`
	Birthday string        `json:"birthday,omitempty"`
}

type nameObject struct {
	FormattedName string `json:"formatted_name"`
	FirstName     string `json:"first_name"`
	MiddleName    string `json:"middle_name,omitempty"`
	LastName      string `json:"last_name,omitempty"`
}

type emailObject struct {
	Email string `json:"email,omitempty"`
	Type  string `json:"type,omitempty"`
}

type phoneObject struct {
	Phone string `json:"phone,omitempty"`
	Type  string `json:"type,omitempty"`
}

// structs that are part of API response (used in multiple task)

type contact struct {
	Input string `json:"input"`
	WaID  string `json:"wa_id"`
}

type message struct {
	ID            string `json:"id"`
	MessageStatus string `json:"message_status,omitempty"`
}

// Task 1: Send Text Message Task

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
	inputStruct := TaskSendTextMessageInput{}
	err := base.ConvertFromStructpb(in, &inputStruct)

	if err != nil {
		return nil, fmt.Errorf("failed to convert input to struct: %v", err)
	}

	req := TaskSendTextMessageReq{
		MessagingProduct: "whatsapp",
		To:               inputStruct.To,
		Type:             "text",
		Text: textObject{
			Body:       inputStruct.Body,
			PreviewURL: inputStruct.PreviewURL == "true",
		},
	}

	resp, err := e.client.SendMessageAPI(&req, &TaskSendTextMessageResp{}, inputStruct.PhoneNumberID)

	if err != nil {
		return nil, fmt.Errorf("failed to do API request: %v", err)
	}

	respStruct := resp.(*TaskSendTextMessageResp)

	outputStruct := TaskSendTextMessageOutput{
		WaID: respStruct.Contacts[0].WaID,
		ID:   respStruct.Messages[0].ID,
	}

	output, err := base.ConvertToStructpb(outputStruct)

	if err != nil {
		return nil, fmt.Errorf("failed to convert output to struct: %v", err)
	}

	return output, nil
}

// Task 2: Send Media Message Task

type TaskSendMediaMessageInput struct {
	PhoneNumberID string `json:"phone-number-id"`
	To            string `json:"to"`
	MediaType     string `json:"media-type"`
	IDOrLink      string `json:"id-or-link"`
	Caption       string `json:"caption"`  //cannot be used in audio
	Filename      string `json:"filename"` //only for document

}

type TaskSendMediaMessageReq struct {
	MessagingProduct string      `json:"messaging_product"`
	To               string      `json:"to"`
	Type             string      `json:"type"`
	Document         mediaObject `json:"document,omitempty,"`
	Audio            mediaObject `json:"audio,omitempty"`
	Image            mediaObject `json:"image,omitempty"`
	Video            mediaObject `json:"video,omitempty"`
}

type TaskSendMediaMessageResp struct {
	MessagingProduct string    `json:"messaging_product"`
	Contacts         []contact `json:"contacts"`
	Messages         []message `json:"messages"`
}

type TaskSendMediaMessageOutput struct {
	WaID string `json:"recipient-wa-id"`
	ID   string `json:"message-id"`
}

func (e *execution) TaskSendMediaMessage(in *structpb.Struct) (*structpb.Struct, error) {
	inputStruct := TaskSendMediaMessageInput{}
	err := base.ConvertFromStructpb(in, &inputStruct)

	if err != nil {
		return nil, fmt.Errorf("failed to convert input to struct: %v", err)
	}

	req := TaskSendMediaMessageReq{
		MessagingProduct: "whatsapp",
		To:               inputStruct.To,
	}

	var id string
	var link string
	if strings.Contains(inputStruct.IDOrLink, "http") {
		link = inputStruct.IDOrLink
	} else {
		id = inputStruct.IDOrLink
	}

	switch inputStruct.MediaType {
	case "document":
		req.Type = "document"
		req.Document = mediaObject{
			ID:       id,
			Link:     link,
			Caption:  inputStruct.Caption,
			Filename: inputStruct.Filename,
		}
	case "audio":
		req.Type = "audio"
		req.Audio = mediaObject{
			ID:   id,
			Link: link,
		}
	case "image":
		req.Type = "image"
		req.Image = mediaObject{
			ID:      id,
			Link:    link,
			Caption: inputStruct.Caption,
		}
	case "video":
		req.Type = "video"
		req.Video = mediaObject{
			ID:      id,
			Link:    link,
			Caption: inputStruct.Caption,
		}
	default:
		return nil, fmt.Errorf("unsupported media type")
	}

	resp, err := e.client.SendMessageAPI(&req, &TaskSendMediaMessageResp{}, inputStruct.PhoneNumberID)

	if err != nil {
		return nil, fmt.Errorf("failed to do API request: %v", err)
	}

	respStruct := resp.(*TaskSendMediaMessageResp)

	outputStruct := TaskSendMediaBasedTemplateMessageOutput{
		WaID: respStruct.Contacts[0].WaID,
		ID:   respStruct.Messages[0].ID,
	}

	output, err := base.ConvertToStructpb(outputStruct)

	if err != nil {
		return nil, fmt.Errorf("failed to convert output to struct: %v", err)
	}

	return output, nil
}

// Task 3: Send Location Message Task
type TaskSendLocationMessageInput struct {
	PhoneNumberID string  `json:"phone-number-id"`
	To            string  `json:"to"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	LocationName  string  `json:"location-name"`
	Address       string  `json:"address"`
}

type TaskSendLocationMessageReq struct {
	MessagingProduct string         `json:"messaging_product"`
	To               string         `json:"to"`
	Type             string         `json:"type"`
	Location         locationObject `json:"location"`
}

type TaskSendLocationMessageResp struct {
	MessagingProduct string    `json:"messaging_product"`
	Contacts         []contact `json:"contacts"`
	Messages         []message `json:"messages"`
}

type TaskSendLocationMessageOutput struct {
	WaID string `json:"recipient-wa-id"`
	ID   string `json:"message-id"`
}

func (e *execution) TaskSendLocationMessage(in *structpb.Struct) (*structpb.Struct, error) {
	inputStruct := TaskSendLocationMessageInput{}
	err := base.ConvertFromStructpb(in, &inputStruct)

	if err != nil {
		return nil, fmt.Errorf("failed to convert input to struct: %v", err)
	}

	req := TaskSendLocationMessageReq{
		MessagingProduct: "whatsapp",
		To:               inputStruct.To,
		Type:             "location",
		Location: locationObject{
			Latitude:  fmt.Sprintf("%f", inputStruct.Latitude),
			Longitude: fmt.Sprintf("%f", inputStruct.Longitude),
			Name:      inputStruct.LocationName,
			Address:   inputStruct.Address,
		},
	}

	resp, err := e.client.SendMessageAPI(&req, &TaskSendLocationMessageResp{}, inputStruct.PhoneNumberID)

	if err != nil {
		return nil, fmt.Errorf("failed to do API request: %v", err)
	}

	respStruct := resp.(*TaskSendLocationMessageResp)

	outputStruct := TaskSendLocationMessageOutput{
		WaID: respStruct.Contacts[0].WaID,
		ID:   respStruct.Messages[0].ID,
	}

	output, err := base.ConvertToStructpb(outputStruct)

	if err != nil {
		return nil, fmt.Errorf("failed to convert output to struct: %v", err)
	}

	return output, nil
}

// Task 4: Send Contact Message Task
type TaskSendContactMessageInput struct {
	PhoneNumberID   string `json:"phone-number-id"`
	To              string `json:"to"`
	FirstName       string `json:"first-name"`
	MiddleName      string `json:"middle-name"`
	LastName        string `json:"last-name"`
	PhoneNumber     string `json:"phone-number"`
	PhoneNumberType string `json:"phone-number-type"`
	Email           string `json:"email"`
	EmailType       string `json:"email-type"`
	Birthdate       string `json:"birthday"`
}

type TaskSendContactMessageReq struct {
	MessagingProduct string         `json:"messaging_product"`
	To               string         `json:"to"`
	Type             string         `json:"type"`
	Contacts         contactsObject `json:"contacts"`
}

type TaskSendContactMessageResp struct {
	MessagingProduct string    `json:"messaging_product"`
	Contacts         []contact `json:"contacts"`
	Messages         []message `json:"messages"`
}

type TaskSendContactMessageOutput struct {
	WaID string `json:"recipient-wa-id"`
	ID   string `json:"message-id"`
}

func (e *execution) TaskSendContactMessage(in *structpb.Struct) (*structpb.Struct, error) {
	inputStruct := TaskSendContactMessageInput{}
	err := base.ConvertFromStructpb(in, &inputStruct)

	if err != nil {
		return nil, fmt.Errorf("failed to convert input to struct: %v", err)
	}

	var formattedName string
	if inputStruct.MiddleName == "" && inputStruct.LastName == "" {
		formattedName = inputStruct.FirstName
	} else if inputStruct.MiddleName == "" {
		formattedName = fmt.Sprintf("%s %s", inputStruct.FirstName, inputStruct.LastName)
	} else if inputStruct.LastName == "" {
		formattedName = fmt.Sprintf("%s %s", inputStruct.FirstName, inputStruct.MiddleName)
	}

	req := TaskSendContactMessageReq{
		MessagingProduct: "whatsapp",
		To:               inputStruct.To,
		Type:             "contacts",
		Contacts: contactsObject{
			Name: nameObject{
				FormattedName: formattedName,
				FirstName:     inputStruct.FirstName,
				MiddleName:    inputStruct.MiddleName,
				LastName:      inputStruct.LastName,
			},
			Emails: []emailObject{
				{
					Email: inputStruct.Email,
					Type:  inputStruct.EmailType,
				},
			},
			Birthday: inputStruct.Birthdate,
		},
	}

	if inputStruct.PhoneNumber != "" {
		if inputStruct.PhoneNumberType == "none" {
			return nil, fmt.Errorf("you forgot to set the phone number type")
		}

		req.Contacts.Phones = append(req.Contacts.Phones, phoneObject{
			Phone: inputStruct.PhoneNumber,
			Type:  inputStruct.PhoneNumberType,
		})

	}

	if inputStruct.Email != "" {
		if inputStruct.EmailType == "none" {
			return nil, fmt.Errorf("you forgot to set the email type")
		}

		req.Contacts.Emails = append(req.Contacts.Emails, emailObject{
			Email: inputStruct.Email,
			Type:  inputStruct.EmailType,
		})
	}

	resp, err := e.client.SendMessageAPI(&req, &TaskSendContactMessageResp{}, inputStruct.PhoneNumberID)

	if err != nil {
		return nil, fmt.Errorf("failed to do API request: %v", err)
	}

	respStruct := resp.(*TaskSendContactMessageResp)

	outputStruct := TaskSendContactMessageOutput{
		WaID: respStruct.Contacts[0].WaID,
		ID:   respStruct.Messages[0].ID,
	}

	output, err := base.ConvertToStructpb(outputStruct)

	if err != nil {
		return nil, fmt.Errorf("failed to convert output to struct: %v", err)
	}

	return output, nil

}
