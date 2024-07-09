package hubspot

import (
	hubspot "github.com/belong-inc/go-hubspot"
)

type RetrieveAssociationInput struct {
	ContactId  string `json:"contact-id"`
	ObjectType string `json:"object-type"`
}

type RetrieveAssociationService interface {
	GetThreadId(contactId string) (*RetrieveThreadIdResponse, error)
	GetCrmId(contactId string, objectType string) (*RetrieveCrmIdResponseHSFormat, error)
}

type RetrieveAssociationServiceOp struct {
	retrieveCrmIdPath    string
	retrieveThreadIdPath string
	client               *hubspot.Client
}

// structs for receiving threads ID

type RetrieveThreadIdResult struct {
	Id string `json:"id"`
}

type RetrieveThreadIdResponse struct {
	Results []RetrieveThreadIdResult `json:"results"`
}

// Used to do http request and get all the thread IDs associated with the contact
func (s *RetrieveAssociationServiceOp) GetThreadId(contactId string) (*RetrieveThreadIdResponse, error) {
	resource := &RetrieveThreadIdResponse{}
	if err := s.client.Get(s.retrieveThreadIdPath+contactId, resource, nil); err != nil {
		return nil, err
	}
	return resource, nil
}

// struct for POST request in order to obtain CRM objects ID

type id struct {
	ContactId string `json:"id"`
}

type RequestQueryCrmId struct {
	Input []id `json:"inputs"`
}

// structs for receiving CRM IDs

type RetrieveCrmIdResultHSFormat struct {
	IdArray []RetrieveCrmId `json:"to"`
}

type RetrieveCrmIdResponseHSFormat struct {
	Results []RetrieveCrmIdResultHSFormat `json:"results"`
}

type RetrieveCrmId struct {
	Id string `json:"id"`
}

// struct used for output (to convert to structpb)

type RetrieveCrmIdResultTaskFormat struct {
	IdArray []RetrieveCrmId `json:"results"`
}

func (s *RetrieveAssociationServiceOp) GetCrmId(contactId string, objectType string) (*RetrieveCrmIdResponseHSFormat, error) {
	resource := &RetrieveCrmIdResponseHSFormat{}

	contactIdInput := id{ContactId: contactId}

	req := &RequestQueryCrmId{}
	req.Input = append(req.Input, contactIdInput)

	path := s.retrieveCrmIdPath + "/" + objectType + "/batch/read"

	if err := s.client.Post(path, req, resource); err != nil {
		return nil, err
	}
	return resource, nil
}
