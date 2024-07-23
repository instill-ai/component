package hubspot

import (
	hubspot "github.com/belong-inc/go-hubspot"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

// following go-hubspot sdk format

// API functions for Thread

type ThreadService interface {
	Get(threadID string) (*TaskGetThreadResp, error)
}

type ThreadServiceOp struct {
	threadPath string
	client     *hubspot.Client
}

func (s *ThreadServiceOp) Get(threadID string) (*TaskGetThreadResp, error) {
	resource := &TaskGetThreadResp{}
	if err := s.client.Get(s.threadPath+"/"+threadID+"/messages", resource, nil); err != nil {
		return nil, err
	}
	return resource, nil
}

// Get Thread

// Get Thread Input

type TaskGetThreadInput struct {
	ThreadID string `json:"thread-id"`
}

// Get Thread Reponse structs

type TaskGetThreadResp struct {
	Results []taskGetThreadRespResult `json:"results"`
}

type taskGetThreadRespResult struct {
	CreatedAt  string                      `json:"createdAt"`
	Client     taskGetThreadRespUserClient `json:"client,omitempty"`
	Senders    []taskGetThreadRespUser     `json:"senders,omitempty"`
	Recipients []taskGetThreadRespUser     `json:"recipients,omitempty"`
	Text       string                      `json:"text,omitempty"`
	Subject    string                      `json:"subject,omitempty"`
}

type taskGetThreadRespUserClient struct {
	ClientType string `json:"clientType,omitempty"`
}

type taskGetThreadRespUser struct {
	Name               string                      `json:"name,omitempty"`
	DeliveryIDentifier taskGetThreadRespIDentifier `json:"deliveryIDentifier,omitempty"`
}

type taskGetThreadRespIDentifier struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

// Get Thread Output structs

type TaskGetThreadOutput struct {
	Results []taskGetThreadOutputResult `json:"results"`
}

type taskGetThreadOutputResult struct {
	CreatedAt  string                    `json:"created-at"`
	Senders    []taskGetThreadOutputUser `json:"senders,omitempty"`
	Recipients []taskGetThreadOutputUser `json:"recipients,omitempty"`
	Text       string                    `json:"text"`
	Subject    string                    `json:"subject"`
}

type taskGetThreadOutputUser struct {
	Name  string `json:"name,omitempty"`
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

func (e *execution) GetThread(input *structpb.Struct) (*structpb.Struct, error) {

	inputStruct := TaskGetThreadInput{}
	err := base.ConvertFromStructpb(input, &inputStruct)

	if err != nil {
		return nil, err
	}

	res, err := e.client.Thread.Get(inputStruct.ThreadID)

	if err != nil {
		return nil, err
	}

	// convert to output struct

	outputStruct := TaskGetThreadOutput{}

	for _, value1 := range res.Results {
		// this will make the output to not contain any system message
		if value1.Client.ClientType == "SYSTEM" {
			continue
		}

		resultOutput := taskGetThreadOutputResult{
			CreatedAt: value1.CreatedAt,
			Text:      value1.Text,
			Subject:   value1.Subject,
		}

		// sender
		for _, value2 := range value1.Senders {
			userOutput := taskGetThreadOutputUser{
				Name:  value2.Name,
				Type:  value2.DeliveryIDentifier.Type,
				Value: value2.DeliveryIDentifier.Value,
			}

			resultOutput.Senders = append(resultOutput.Senders, userOutput)
		}

		// recipient
		for _, value3 := range value1.Recipients {
			userOutput := taskGetThreadOutputUser{
				Name:  value3.Name,
				Type:  value3.DeliveryIDentifier.Type,
				Value: value3.DeliveryIDentifier.Value,
			}

			resultOutput.Recipients = append(resultOutput.Recipients, userOutput)

		}

		outputStruct.Results = append(outputStruct.Results, resultOutput)

	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}

	return output, nil
}
