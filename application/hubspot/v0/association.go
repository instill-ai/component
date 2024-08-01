package hubspot

import (
	"fmt"

	hubspot "github.com/belong-inc/go-hubspot"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

// Retrieve Association is a custom feature
// Will implement it following go-hubspot sdk format

// API functions for Retrieve Association

type RetrieveAssociationService interface {
	GetThreadID(contactID string) (*TaskRetrieveAssociationThreadResp, error)
	GetCrmID(contactID string, objectType string) (*TaskRetrieveAssociationCrmResp, error)
}

type RetrieveAssociationServiceOp struct {
	retrieveCrmIDPath    string
	retrieveThreadIDPath string
	client               *hubspot.Client
}

func (s *RetrieveAssociationServiceOp) GetThreadID(contactID string) (*TaskRetrieveAssociationThreadResp, error) {
	resource := &TaskRetrieveAssociationThreadResp{}
	if err := s.client.Get(s.retrieveThreadIDPath+contactID, resource, nil); err != nil {
		return nil, err
	}
	return resource, nil
}

func (s *RetrieveAssociationServiceOp) GetCrmID(contactID string, objectType string) (*TaskRetrieveAssociationCrmResp, error) {
	resource := &TaskRetrieveAssociationCrmResp{}

	contactIDInput := TaskRetrieveAssociationCrmReqID{ContactID: contactID}

	req := &TaskRetrieveAssociationCrmReq{}
	req.Input = append(req.Input, contactIDInput)

	path := s.retrieveCrmIDPath + "/" + objectType + "/batch/read"

	if err := s.client.Post(path, req, resource); err != nil {
		return nil, err
	}
	return resource, nil
}

// Retrieve Association: use contact id to get the object ID associated with it

type TaskRetrieveAssociationInput struct {
	ContactID  string `json:"contact-id"`
	ObjectType string `json:"object-type"`
}

// Retrieve Association Task is mainly divided into two:
// 1. GetThreadID
// 2. GetCrmID
// Basically, these two will have seperate structs for handling request/response

// For GetThreadID

type TaskRetrieveAssociationThreadResp struct {
	Results []struct {
		ID string `json:"id"`
	} `json:"results"`
}

// For GetCrmID

type TaskRetrieveAssociationCrmReq struct {
	Input []TaskRetrieveAssociationCrmReqID `json:"inputs"`
}

type TaskRetrieveAssociationCrmReqID struct {
	ContactID string `json:"id"`
}

type TaskRetrieveAssociationCrmResp struct {
	Results []taskRetrieveAssociationCrmRespResult `json:"results"`
}

type taskRetrieveAssociationCrmRespResult struct {
	IDArray []struct {
		ID string `json:"id"`
	} `json:"to"`
}

// Retrieve Association Output

type TaskRetrieveAssociationOutput struct {
	ObjectIDs []string `json:"object-ids"`
}

func (e *execution) RetrieveAssociation(input *structpb.Struct) (*structpb.Struct, error) {
	inputStruct := TaskRetrieveAssociationInput{}
	err := base.ConvertFromStructpb(input, &inputStruct)

	if err != nil {
		return nil, fmt.Errorf("failed to convert input to struct: %v", err)
	}

	// API calls to retrieve association for Threads and CRM objects are different

	var objectIDs []string
	if inputStruct.ObjectType == "Threads" {

		// To handle Threads
		res, err := e.client.RetrieveAssociation.GetThreadID(inputStruct.ContactID)

		if err != nil {
			return nil, err
		}

		if len(res.Results) == 0 {
			return nil, fmt.Errorf("no object ID found")
		}

		objectIDs = make([]string, len(res.Results))
		for index, value := range res.Results {
			objectIDs[index] = value.ID
		}

	} else {

		// To handle CRM objects
		res, err := e.client.RetrieveAssociation.GetCrmID(inputStruct.ContactID, inputStruct.ObjectType)

		if err != nil {
			return nil, err
		}

		if len(res.Results) == 0 {
			return nil, fmt.Errorf("no object ID found")
		}

		// only take the first Result, because the input is only one contact id
		objectIDs = make([]string, len(res.Results[0].IDArray))
		for index, value := range res.Results[0].IDArray {
			objectIDs[index] = value.ID
		}

	}

	outputStruct := TaskRetrieveAssociationOutput{
		ObjectIDs: objectIDs,
	}

	output, err := base.ConvertToStructpb(outputStruct)

	if err != nil {
		return nil, fmt.Errorf("failed to convert output to struct: %v", err)
	}

	return output, nil
}

// Create Association (not a task)
// This section (create association) is used in:
// create contact task to create contact -> objects (company, ticket, deal) association
// create company task to create company -> contact association
// create deal task to create deal -> contact association
// create ticket task to create ticket -> contact association

type CreateAssociationReq struct {
	Associations []association `json:"inputs"`
}

type association struct {
	From struct {
		ID string `json:"id"`
	} `json:"from"`
	To struct {
		ID string `json:"id"`
	} `json:"to"`
	Type string `json:"type"`
}

type CreateAssociationResponse struct {
	Status string `json:"status"`
}

// CreateAssociation is used to create batch associations between objects

func CreateAssociation(fromID *string, toIDs *[]string, fromObjectType string, toObjectType string, e *execution) error {
	req := &CreateAssociationReq{
		Associations: make([]association, len(*toIDs)),
	}

	//for any association created related to company, it will use non-primary label.
	//for more info: https://developers.hubspot.com/beta-docs/guides/api/crm/associations#association-type-id-values

	var associationType string
	if toObjectType == "company" {
		switch fromObjectType { //use switch here in case other association of object -> company want to be created in the future
		case "contact":
			associationType = "279"
		}

	} else if fromObjectType == "company" {
		switch toObjectType {
		case "contact":
			associationType = "280"
		}
	} else {
		associationType = fmt.Sprintf("%s_to_%s", fromObjectType, toObjectType)
	}

	for index, toID := range *toIDs {

		req.Associations[index] = association{
			From: struct {
				ID string `json:"id"`
			}{
				ID: *fromID,
			},
			To: struct {
				ID string `json:"id"`
			}{
				ID: toID,
			},
			Type: associationType,
		}
	}

	createAssociationPath := fmt.Sprintf("crm/v3/associations/%s/%s/batch/create", fromObjectType, toObjectType)

	resp := &CreateAssociationResponse{}

	if err := e.client.Post(createAssociationPath, req, resp); err != nil {
		return err
	}

	if resp.Status != "COMPLETE" {
		return fmt.Errorf("failed to create association")
	}

	return nil
}
