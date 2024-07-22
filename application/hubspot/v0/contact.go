package hubspot

import (
	"fmt"
	"strings"

	hubspot "github.com/belong-inc/go-hubspot"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

// Get Contact

type TaskGetContactInput struct {
	ContactIdOrEmail string `json:"contact-id-or-email"`
}

type TaskGetContactResp struct {
	OwnerId        string `json:"hubspot_owner_id,omitempty"`
	Email          string `json:"email,omitempty"`
	FirstName      string `json:"firstname,omitempty"`
	LastName       string `json:"lastname,omitempty"`
	PhoneNumber    string `json:"phone,omitempty"`
	Company        string `json:"company,omitempty"`
	JobTitle       string `json:"jobtitle,omitempty"`
	LifecycleStage string `json:"lifecyclestage,omitempty"`
	LeadStatus     string `json:"hs_lead_status,omitempty"`
	ContactId      string `json:"hs_object_id"`
}

type TaskGetContactOutput struct {
	OwnerId        string `json:"owner-id,omitempty"`
	Email          string `json:"email,omitempty"`
	FirstName      string `json:"first-name,omitempty"`
	LastName       string `json:"last-name,omitempty"`
	PhoneNumber    string `json:"phone-number,omitempty"`
	Company        string `json:"company,omitempty"`
	JobTitle       string `json:"job-title,omitempty"`
	LifecycleStage string `json:"lifecycle-stage,omitempty"`
	LeadStatus     string `json:"lead-status,omitempty"`
	ContactId      string `json:"contact-id"`
}

func (e *execution) GetContact(input *structpb.Struct) (*structpb.Struct, error) {
	inputStruct := TaskGetContactInput{}

	err := base.ConvertFromStructpb(input, &inputStruct)

	if err != nil {
		return nil, err
	}

	uniqueKey := inputStruct.ContactIdOrEmail

	// If user enter email instead of contact ID
	if strings.Contains(uniqueKey, "@") {
		uniqueKey += "?idProperty=email"
	}

	res, err := e.client.CRM.Contact.Get(uniqueKey, &TaskGetContactResp{}, &hubspot.RequestQueryOption{CustomProperties: []string{"phone"}})

	if err != nil {
		return nil, err
	}

	contactInfo := res.Properties.(*TaskGetContactResp)

	outputStruct := TaskGetContactOutput(*contactInfo)

	output, err := base.ConvertToStructpb(outputStruct)

	if err != nil {
		return nil, err
	}

	return output, nil
}

// Create Contact

// TODO: to future me, dont forget to create association feature in the future
type TaskCreateContactInput struct {
	OwnerId                    string   `json:"owner-id"`
	Email                      string   `json:"email"`
	FirstName                  string   `json:"first-name"`
	LastName                   string   `json:"last-name"`
	PhoneNumber                string   `json:"phone-number"`
	Company                    string   `json:"company"`
	JobTitle                   string   `json:"job-title"`
	LifecycleStage             string   `json:"lifecycle-stage"`
	LeadStatus                 string   `json:"lead-status"`
	CreateDealsAssociation     []string `json:"create-deals-association"`
	CreateCompaniesAssociation []string `json:"create-companies-association"`
	CreateTicketsAssociation   []string `json:"create-tickets-association"`
}

type TaskCreateContactReq struct {
	OwnerId        string `json:"hubspot_owner_id,omitempty"`
	Email          string `json:"email,omitempty"`
	FirstName      string `json:"firstname,omitempty"`
	LastName       string `json:"lastname,omitempty"`
	PhoneNumber    string `json:"phone,omitempty"`
	Company        string `json:"company,omitempty"`
	JobTitle       string `json:"jobtitle,omitempty"`
	LifecycleStage string `json:"lifecyclestage,omitempty"`
	LeadStatus     string `json:"hs_lead_status,omitempty"`
	ContactId      string `json:"hs_object_id"`
}

type TaskCreateContactOutput struct {
	ContactId string `json:"contact-id"`
}

func (e *execution) CreateContact(input *structpb.Struct) (*structpb.Struct, error) {
	inputStruct := TaskCreateContactInput{}

	err := base.ConvertFromStructpb(input, &inputStruct)
	fmt.Println(input)

	if err != nil {
		return nil, err
	}

	req := TaskCreateContactReq{
		OwnerId:        inputStruct.OwnerId,
		Email:          inputStruct.Email,
		FirstName:      inputStruct.FirstName,
		LastName:       inputStruct.LastName,
		PhoneNumber:    inputStruct.PhoneNumber,
		Company:        inputStruct.Company,
		JobTitle:       inputStruct.JobTitle,
		LifecycleStage: inputStruct.LifecycleStage,
		LeadStatus:     inputStruct.LeadStatus,
	}

	res, err := e.client.CRM.Contact.Create(&req)

	if err != nil {
		return nil, err
	}

	contactId := res.Properties.(*TaskCreateContactReq).ContactId

	outputStruct := TaskCreateContactOutput{ContactId: contactId}

	output, err := base.ConvertToStructpb(outputStruct)

	if err != nil {
		return nil, err
	}

	// This section is for creating associations (contact -> object)

	if len(inputStruct.CreateDealsAssociation) != 0 {
		err := CreateAssociation(&outputStruct.ContactId, &inputStruct.CreateDealsAssociation, "contact", "deal", e)

		if err != nil {
			return nil, err
		}
	}
	if len(inputStruct.CreateCompaniesAssociation) != 0 {
		err := CreateAssociation(&outputStruct.ContactId, &inputStruct.CreateCompaniesAssociation, "contact", "company", e)

		if err != nil {
			return nil, err
		}
	}
	if len(inputStruct.CreateTicketsAssociation) != 0 {
		err := CreateAssociation(&outputStruct.ContactId, &inputStruct.CreateTicketsAssociation, "contact", "ticket", e)

		if err != nil {
			return nil, err
		}
	}
	return output, nil
}
