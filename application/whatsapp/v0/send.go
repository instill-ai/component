package whatsapp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

// Reference: https://developers.facebook.com/docs/whatsapp/cloud-api/reference/messages

// this file is used to handle all send related tasks for WhatsApp, which include
// send template task: https://developers.facebook.com/docs/whatsapp/cloud-api/guides/send-message-templates
// send catalog template task: https://developers.facebook.com/docs/whatsapp/cloud-api/guides/send-message-templates/mpm-template-messages
// send message task: https://developers.facebook.com/docs/whatsapp/cloud-api/guides/send-messages

// Objects (structs that are part of request struct; can be used in more than one task)

type templateObject struct {
	Name       string            `json:"name"`
	Language   languageObject    `json:"language"`
	Components []componentObject `json:"components,omitempty"`
}

type languageObject struct {
	Code string `json:"code"`
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

// Component parameters

// used when the header type is text (also used for body)
type textParameter struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// used when the header type is image
type imageParameter struct {
	Type  string      `json:"type"`
	Image mediaObject `json:"image"`
}

// used when the header type is video
type videoParameter struct {
	Type  string      `json:"type"`
	Video mediaObject `json:"video"`
}

// used when the header type is document
type documentParameter struct {
	Type     string      `json:"type"`
	Document mediaObject `json:"document"`
}

// used when the header type is location
type locationParameter struct {
	Type     string         `json:"type"`
	Location locationObject `json:"location"`
}

// used for button component
type buttonParameter struct {
	Type    string `json:"type"`
	Payload string `json:"payload,omitempty"`
	Text    string `json:"text,omitempty"`
}

// structs that are part of response (used in multiple task)

type contact struct {
	Input string `json:"input"`
	WaID  string `json:"wa_id"`
}

type message struct {
	ID            string `json:"id"`
	MessageStatus string `json:"message_status"`
}

// Send Template task

type TaskSendTemplateInput struct {
	PhoneNumberID    string   `json:"phone-number-id"`
	To               string   `json:"to"`
	HeaderType       string   `json:"header-type"`
	TemplateName     string   `json:"name"`
	LanguageCode     string   `json:"language-code"`
	HeaderParameters []string `json:"header-parameters"`
	BodyParameters   []string `json:"body-parameters"`
	ButtonParameters []string `json:"button-parameters"`
}

type TaskSendTemplateReq struct {
	MessagingProduct string         `json:"messaging_product"`
	To               string         `json:"to"`
	Type             string         `json:"type"`
	Template         templateObject `json:"template"`
}

type TaskSendTemplateResp struct {
	MessagingProduct string    `json:"messaging_product"`
	Contacts         []contact `json:"contacts"`
	Messages         []message `json:"messages"`
}

type TaskSendTemplateOutput struct {
	WaID          string `json:"recipient-wa-id"`
	ID            string `json:"message-id"`
	MessageStatus string `json:"message-status,omitempty"`
}

func (e *execution) SendTemplate(in *structpb.Struct) (*structpb.Struct, error) {

	inputStruct := TaskSendTemplateInput{}
	err := base.ConvertFromStructpb(in, &inputStruct)

	if err != nil {
		return nil, err
	}

	req := TaskSendTemplateReq{
		MessagingProduct: "whatsapp",
		To:               inputStruct.To,
		Type:             "template",
		Template: templateObject{
			Name: inputStruct.TemplateName,
			Language: languageObject{
				Code: inputStruct.LanguageCode,
			},
		},
	}

	// to assign HeaderParameters
	// Header can have 6 types: none, text, image, video, document & location.

	// create a header component

	switch inputStruct.HeaderType {
	case "text":
		// the text header type can have 0 parameter, so there is no need to have an error message if there is no parameter
		if len(inputStruct.HeaderParameters) != 0 {
			component := componentObject{
				Type:       "header",
				Parameters: make([]interface{}, len(inputStruct.HeaderParameters)),
			}

			for index, value := range inputStruct.HeaderParameters {
				component.Parameters[index] = textParameter{
					Type: "text",
					Text: value,
				}
			}
			req.Template.Components = append(req.Template.Components, component)
		}

	case "image":

		// the image header type only has 1 parameter, which is the id/link of the image
		if len(inputStruct.HeaderParameters) == 1 {

			component := componentObject{
				Type:       "header",
				Parameters: make([]interface{}, 1),
			}

			firstParam := inputStruct.HeaderParameters[0]
			if strings.Contains(firstParam, "http") {
				component.Parameters[0] = imageParameter{
					Type: "image",
					Image: mediaObject{
						Link: firstParam,
					},
				}
			} else {
				component.Parameters[0] = imageParameter{
					Type: "image",
					Image: mediaObject{
						ID: firstParam,
					},
				}
			}
			req.Template.Components = append(req.Template.Components, component)
		} else {
			return nil, fmt.Errorf("the image header type requires one parameter (in the header parameters), which is the id/link of the image. format: [id/link]")
		}

	case "video":
		// the video header type only has 1 parameter, which is the id/link of the video

		if len(inputStruct.HeaderParameters) == 1 {
			component := componentObject{
				Type:       "header",
				Parameters: make([]interface{}, 1),
			}

			firstParam := inputStruct.HeaderParameters[0]
			if strings.Contains(firstParam, "http") {
				component.Parameters[0] = videoParameter{
					Type: "video",
					Video: mediaObject{
						Link: firstParam,
					},
				}
			} else {
				component.Parameters[0] = videoParameter{
					Type: "video",
					Video: mediaObject{
						ID: firstParam,
					},
				}
			}
			req.Template.Components = append(req.Template.Components, component)
		} else {
			return nil, fmt.Errorf("the video header type requires one parameter (in the header parameters), which is the id/link of the video. format: [id/link]")
		}

	case "document":

		if len(inputStruct.HeaderParameters) == 1 || len(inputStruct.HeaderParameters) == 2 {

			component := componentObject{
				Type: "header",
			}

			firstParam := inputStruct.HeaderParameters[0]

			var secondParam string
			if len(inputStruct.HeaderParameters) > 1 {
				secondParam = inputStruct.HeaderParameters[1]
				component.Parameters = make([]interface{}, 2)
			} else {
				component.Parameters = make([]interface{}, 1)
			}

			if strings.Contains(firstParam, "http") {
				component.Parameters[0] = documentParameter{
					Type: "document",
					Document: mediaObject{
						Link:     firstParam,
						Filename: secondParam,
					},
				}
			} else {
				component.Parameters[0] = documentParameter{
					Type: "document",
					Document: mediaObject{
						ID:       firstParam,
						Filename: secondParam,
					},
				}
			}
			req.Template.Components = append(req.Template.Components, component)
		} else {
			return nil, fmt.Errorf("the document header type requires one or two parameter (in the header parameters). The first parameter is the id/link of the document (required), and the second parameter is the filename which can be used to specify the extension of the file as well (optional). format: [id/link, filename]")
		}

	case "location":

		if len(inputStruct.HeaderParameters) == 4 {

			component := componentObject{
				Type:       "header",
				Parameters: make([]interface{}, 1),
			}

			component.Parameters[0] =
				locationParameter{
					Type: "location",
					Location: locationObject{
						Latitude:  inputStruct.HeaderParameters[0],
						Longitude: inputStruct.HeaderParameters[1],
						Name:      inputStruct.HeaderParameters[2],
						Address:   inputStruct.HeaderParameters[3],
					},
				}

			req.Template.Components = append(req.Template.Components, component)
		} else {
			return nil, fmt.Errorf("the location header type requires 4 parameters which are: latitude, longitude, name, address. format: [latitude, longitude, name, address]")
		}

	}

	// create a body component if there is any body parameters

	if len(inputStruct.BodyParameters) != 0 {
		component := componentObject{
			Type:       "body",
			Parameters: make([]interface{}, len(inputStruct.BodyParameters)),
		}

		for index, value := range inputStruct.BodyParameters {
			component.Parameters[index] = textParameter{
				Type: "text",
				Text: value,
			}
		}

		req.Template.Components = append(req.Template.Components, component)
	}

	// create button component if there is any
	// one parameter -> one button component

	for index, value := range inputStruct.ButtonParameters {
		splitParam := strings.Split(value, ";")

		if len(splitParam) != 2 {
			return nil, fmt.Errorf("format is wrong, it must be 'button_type;value_of_the_parameter'. Example: quick_reply;randomvalue")
		}

		var param buttonParameter
		if splitParam[0] == "quick_reply" || splitParam[0] == "copy_code" {
			param = buttonParameter{
				Type:    "payload",
				Payload: splitParam[1],
			}

		} else if splitParam[0] == "url" {
			param = buttonParameter{
				Type: "text",
				Text: splitParam[1],
			}

		} else {
			return nil, fmt.Errorf("wrong button_type. button_type is either 'quick_reply', 'url' or 'copy_code'")
		}

		component := componentObject{
			Type:          "button",
			ButtonSubType: splitParam[0],
			ButtonIndex:   strconv.Itoa(index),
		}

		component.Parameters = append(component.Parameters, param)

		req.Template.Components = append(req.Template.Components, component)
	}

	resp, err := e.client.SendMessageAPI(&req, &TaskSendTemplateResp{}, inputStruct.PhoneNumberID)

	if err != nil {
		return nil, err
	}

	respStruct := resp.(*TaskSendTemplateResp)

	// only take the first index because we are sending a template to an individual, so there will only be one contact and one message.
	outputStruct := TaskSendTemplateOutput{
		WaID:          respStruct.Contacts[0].WaID,
		ID:            respStruct.Messages[0].ID,
		MessageStatus: respStruct.Messages[0].MessageStatus,
	}

	output, err := base.ConvertToStructpb(outputStruct)

	if err != nil {
		return nil, err
	}

	return output, nil
}

// Send Catalog Template task
