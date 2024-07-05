package hubspot

import (
	hubspot "github.com/belong-inc/go-hubspot"
)

type RetrieveAssociationInput struct {
	ContactId  string `json:"contactId"`
	ObjectType string `json:"objectType"`
}

type RetrieveAssociationService interface {
	GetThreadId(contactId string) (*RetrieveThreadIdResponse, error)
	GetCrmId(contactId string, objectType string) (*RetrieveCrmIdResponse, error)
}

type RetrieveAssociationServiceOp struct {
	retrieveCrmIdPath    string
	retrieveThreadIdPath string
	client               *hubspot.Client
}

// structs for receiving threads ID

type RetrieveThreadIdResult struct {
	ID string `json:"id"`
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

type RetrieveCrmId struct {
	Id string `json:"id"`
}

type RetrieveCrmIdResult struct {
	IdArray []RetrieveCrmId `json:"to"`
}

type RetrieveCrmIdResponse struct {
	Results []RetrieveCrmIdResult `json:"results"`
}

// structs used for CRM IDs task format (to utilize base.ConvertToStructpb)
type RetrieveCrmIdResultTaskFormat struct {
	IdArray []RetrieveCrmId `json:"results"`
}

func (s *RetrieveAssociationServiceOp) GetCrmId(contactId string, objectType string) (*RetrieveCrmIdResponse, error) {
	resource := &RetrieveCrmIdResponse{}

	contactIdInput := id{ContactId: contactId}

	req := &RequestQueryCrmId{}
	req.Input = append(req.Input, contactIdInput)

	path := s.retrieveCrmIdPath + "/" + objectType + "/batch/read"

	if err := s.client.Post(path, req, resource); err != nil {
		return nil, err
	}
	return resource, nil
}
