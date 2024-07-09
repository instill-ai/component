package hubspot

import (
	hubspot "github.com/belong-inc/go-hubspot"
)

type ThreadService interface {
	Get(threadId string) (*ThreadResponseHSFormat, error)
}

type ThreadServiceOp struct {
	threadPath string
	client     *hubspot.Client
}

// The structs used to receive response from HubSpot API

// Give more information about sender/receiver. If it is email, the value is email of the sender/receiver.
type ThreadDeliveryIdentifier struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

type ThreadUserHSFormat struct {
	Name               string                   `json:"name,omitempty"`
	DeliveryIdentifier ThreadDeliveryIdentifier `json:"deliveryIdentifier,omitempty"`
}

type ThreadClientHSFormat struct {
	ClientType string `json:"clientType,omitempty"`
}

type ThreadResultHSFormat struct {
	CreatedAt  string               `json:"createdAt"`
	Client     ThreadClientHSFormat `json:"client,omitempty"`
	Senders    []ThreadUserHSFormat `json:"senders,omitempty"`
	Recipients []ThreadUserHSFormat `json:"recipients,omitempty"`
	Text       string               `json:"text,omitempty"`
	Subject    string               `json:"subject,omitempty"`
}

// The response received from HubSpot when requesting thread.
type ThreadResponseHSFormat struct {
	Results []ThreadResultHSFormat `json:"results"`
}

// Structs used for task format

type ThreadUserTaskFormat struct {
	Name  string `json:"name,omitempty"`
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

type ThreadResultTaskFormat struct {
	CreatedAt  string                 `json:"created-at"`
	Senders    []ThreadUserTaskFormat `json:"senders,omitempty"`
	Recipients []ThreadUserTaskFormat `json:"recipients,omitempty"`
	Text       string                 `json:"text"`
	Subject    string                 `json:"subject"`
}

type ThreadResponseTaskFormat struct {
	Results []ThreadResultTaskFormat `json:"results"`
}

// Used to do http request and get thread
func (s *ThreadServiceOp) Get(threadId string) (*ThreadResponseHSFormat, error) {
	resource := &ThreadResponseHSFormat{}
	if err := s.client.Get(s.threadPath+"/"+threadId+"/messages", resource, nil); err != nil {
		return nil, err
	}
	return resource, nil
}

// Used to convert struct response to struct for task. Also used to collapse ThreadDeliveryIdentifier and ThreadUser into one struct.
func ThreadConvertToTaskFormat(res *ThreadResponseHSFormat) *ThreadResponseTaskFormat {
	responseOutput := ThreadResponseTaskFormat{}

	for _, value1 := range res.Results {
		// this will make the output to not contain any system message
		if value1.Client.ClientType == "SYSTEM" {
			continue
		}

		resultOutput := ThreadResultTaskFormat{
			CreatedAt: value1.CreatedAt,
			Text:      value1.Text,
			Subject:   value1.Subject,
		}

		// sender
		for _, value2 := range value1.Senders {
			userOutput := ThreadUserTaskFormat{
				Name:  value2.Name,
				Type:  value2.DeliveryIdentifier.Type,
				Value: value2.DeliveryIdentifier.Value,
			}

			resultOutput.Senders = append(resultOutput.Senders, userOutput)
		}

		// recipient
		for _, value3 := range value1.Recipients {
			userOutput := ThreadUserTaskFormat{
				Name:  value3.Name,
				Type:  value3.DeliveryIdentifier.Type,
				Value: value3.DeliveryIdentifier.Value,
			}

			resultOutput.Recipients = append(resultOutput.Recipients, userOutput)

		}

		responseOutput.Results = append(responseOutput.Results, resultOutput)

	}

	return &responseOutput
}
