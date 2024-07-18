package whatsapp

// Reference: https://developers.facebook.com/docs/whatsapp/cloud-api/reference/messages

//structs for request

type MessageObjectReq struct {
	MessagingProduct string          `json:"messaging_product"`
	To               string          `json:"to"`
	Type             string          `json:"type"`
	Template         *TemplateObject `json:"template,omitempty"`
}

// TemplateObject is part of MessageObjectReq

type TemplateObject struct {
	Name       string             `json:"name"`
	Language   LanguageObject     `json:"language"`
	Components []*ComponentObject `json:"components,omitempty"`
}

type LanguageObject struct {
	Code string `json:"code"`
}

// Component type is either header, body or button
// Note:
// footer cannot have any parameters.
// header can have various parameters: text, image, location, document and video
// body support text parameters (along with currency and date time), but for our implementation, we will only do text parameter (for simplification). User can actually just input currency and date time as text as well, so it is not a big deal.
// button type is quick reply and call to action. Can either specify payload or text as the parameter for button.

type ComponentObject struct {
	Type          string        `json:"type"`
	Parameters    []interface{} `json:"parameters"`
	ButtonSubType string        `json:"sub_type,omitempty"` // only used when the type is button. Type of button
	ButtonIndex   string        `json:"index,omitempty"`    // only used when the type is button. Refers to the position index of the button
}

// used when the header type is text (also used for body)
type TextParameter struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// used when the header type is image
type ImageParameter struct {
	Type  string      `json:"type"`
	Image MediaObject `json:"image"`
}

// used when the header type is video
type VideoParameter struct {
	Type  string      `json:"type"`
	Video MediaObject `json:"video"`
}

// used when the header type is document
type DocumentParameter struct {
	Type     string      `json:"type"`
	Document MediaObject `json:"document"`
}

type MediaObject struct {
	Id       string `json:"id,omitempty"`
	Link     string `json:"link,omitempty"`     // if id is used, no need link.
	Filename string `json:"filename,omitempty"` // only for document. This will specify the format of the file displayed in WhatsApp
	Caption  string `json:"caption,omitempty"`  // cannot be used in template message
}

// used when the header type is location
type LocationParameter struct {
	Type     string         `json:"type"`
	Location LocationObject `json:"location"`
}

type LocationObject struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
	Name      string `json:"name"`
	Address   string `json:"address"`
}

// used for button component
type ButtonParameter struct {
	Type    string `json:"type"`
	Payload string `json:"payload,omitempty"`
	Text    string `json:"text,omitempty"`
}

// structs for response

type Contact struct {
	Input string `json:"input"`
	WaId  string `json:"wa_id"`
}

type Message struct {
	Id            string `json:"id"`
	MessageStatus string `json:"message_status"`
}

type MessageObjectResp struct {
	MessagingProduct string    `json:"messaging_product"`
	Contacts         []Contact `json:"contacts"`
	Messages         []Message `json:"messages"`
}

// struct for input (send template task)

type TemplateInput struct {
	PhoneNumberId    string   `json:"phone-number-id"`
	To               string   `json:"to"`
	HeaderType       string   `json:"header-type"`
	TemplateName     string   `json:"name"`
	LanguageCode     string   `json:"language-code"`
	HeaderParameters []string `json:"header-parameters"`
	BodyParameters   []string `json:"body-parameters"`
	ButtonParameters []string `json:"button-parameters"`
}

// struct for output (send template task)

type SendTemplateOutput struct {
	WaId          string `json:"recipient-wa-id"`
	Id            string `json:"message-id"`
	MessageStatus string `json:"message-status,omitempty"`
}
