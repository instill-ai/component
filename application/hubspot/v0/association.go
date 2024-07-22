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
	GetThreadId(contactId string) (*TaskRetrieveAssociationThreadResp, error)
	GetCrmId(contactId string, objectType string) (*TaskRetrieveAssociationCrmResp, error)
}

type RetrieveAssociationServiceOp struct {
	retrieveCrmIdPath    string
	retrieveThreadIdPath string
	client               *hubspot.Client
}

func (s *RetrieveAssociationServiceOp) GetThreadId(contactId string) (*TaskRetrieveAssociationThreadResp, error) {
	resource := &TaskRetrieveAssociationThreadResp{}
	if err := s.client.Get(s.retrieveThreadIdPath+contactId, resource, nil); err != nil {
		return nil, err
	}
	return resource, nil
}

func (s *RetrieveAssociationServiceOp) GetCrmId(contactId string, objectType string) (*TaskRetrieveAssociationCrmResp, error) {
	resource := &TaskRetrieveAssociationCrmResp{}

	contactIdInput := TaskRetrieveAssociationCrmReqId{ContactId: contactId}

	req := &TaskRetrieveAssociationCrmReq{}
	req.Input = append(req.Input, contactIdInput)

	path := s.retrieveCrmIdPath + "/" + objectType + "/batch/read"

	if err := s.client.Post(path, req, resource); err != nil {
		return nil, err
	}
	return resource, nil
}

// Retrieve Association: use contact id to get the object ID associated with it

type TaskRetrieveAssociationInput struct {
	ContactId  string `json:"contact-id"`
	ObjectType string `json:"object-type"`
}

// Retrieve Association Task is mainly divided into two:
// 1. GetThreadId
// 2. GetCrmId
// Basically, these two will have seperate structs for handling request/response

// For GetThreadId

type TaskRetrieveAssociationThreadResp struct {
	Results []struct {
		Id string `json:"id"`
	} `json:"results"`
}

// For GetCrmId

type TaskRetrieveAssociationCrmReq struct {
	Input []TaskRetrieveAssociationCrmReqId `json:"inputs"`
}

type TaskRetrieveAssociationCrmReqId struct {
	ContactId string `json:"id"`
}

type TaskRetrieveAssociationCrmResp struct {
	Results []taskRetrieveAssociationCrmRespResult `json:"results"`
}

type taskRetrieveAssociationCrmRespResult struct {
	IdArray []struct {
		Id string `json:"id"`
	} `json:"to"`
}

// Retrieve Association Output

type TaskRetrieveAssociationOutput struct {
	ObjectIds []string `json:"object-ids"`
}

func (e *execution) RetrieveAssociation(input *structpb.Struct) (*structpb.Struct, error) {
	retrieveInput := TaskRetrieveAssociationInput{}
	err := base.ConvertFromStructpb(input, &retrieveInput)

	if err != nil {
		return nil, err
	}

	// API calls to retrieve association for Threads and CRM objects are different

	var objectIds []string
	if retrieveInput.ObjectType == "Threads" {
		// To handle Threads
		res, err := e.client.RetrieveAssociation.GetThreadId(retrieveInput.ContactId)

		if err != nil {
			return nil, err
		}

		objectIds = make([]string, len(res.Results))
		for index, value := range res.Results {
			objectIds[index] = value.Id
		}

	} else {

		// To handle CRM objects
		res, err := e.client.RetrieveAssociation.GetCrmId(retrieveInput.ContactId, retrieveInput.ObjectType)

		if err != nil {
			return nil, err
		}

		// only take the first Result, because the input is only one contact id
		objectIds = make([]string, len(res.Results))
		for index, value := range res.Results[0].IdArray {
			objectIds[index] = value.Id
		}

	}

	outputStruct := TaskRetrieveAssociationOutput{
		ObjectIds: objectIds,
	}

	output, err := base.ConvertToStructpb(outputStruct)

	if err != nil {
		return nil, err
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

func CreateAssociation(fromId *string, toIds *[]string, fromObjectType string, toObjectType string, e *execution) error {
	req := &CreateAssociationReq{
		Associations: make([]association, len(*toIds)),
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

	for index, toId := range *toIds {

		req.Associations[index] = association{
			From: struct {
				ID string `json:"id"`
			}{
				ID: *fromId,
			},
			To: struct {
				ID string `json:"id"`
			}{
				ID: toId,
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
