package freshdesk

import (
	"fmt"

	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

// name, email, phone, mobile, description, job_title, tags, language, time_zone, company_id, unique_external_id, twitter_id, view_all_tickets, deletedc, other_companies, created_at, updated_at

const (
	ContactPath = "contacts"
)

// API functions for Contact

func (c *FreshdeskClient) GetContact(contactID int64) (*TaskGetContactResponse, error) {
	resp := &TaskGetContactResponse{}

	httpReq := c.httpclient.R().SetResult(resp)
	if _, err := httpReq.Get(fmt.Sprintf("/%s/%d", ContactPath, contactID)); err != nil {
		return nil, err
	}
	return resp, nil
}

// Task 1: Get Contact

type TaskGetContactInput struct {
	ContactID int64 `json:"contact-id"`
}

type TaskGetContactResponse struct {
	Name              string                                   `json:"name"`
	Email             string                                   `json:"email"`
	Phone             string                                   `json:"phone"`
	Mobile            string                                   `json:"mobile"`
	Description       string                                   `json:"description"`
	Address           string                                   `json:"address"`
	JobTitle          string                                   `json:"job_title"`
	Tags              []string                                 `json:"tags"`
	Language          string                                   `json:"language"`
	TimeZone          string                                   `json:"time_zone"`
	CompanyID         int64                                    `json:"company_id"`
	UniqueExternalID  string                                   `json:"unique_external_id"`
	TwitterID         string                                   `json:"twitter_id"`
	ViewAllTickets    bool                                     `json:"view_all_tickets"`
	Deleted           bool                                     `json:"deleted"`
	Active            bool                                     `json:"active"`
	OtherEmails       []string                                 `json:"other_emails"`
	OtherCompanies    []taskGetContactResponseOtherCompany     `json:"other_companies"`
	OtherPhoneNumbers []taskGetContactResponseOtherPhoneNumber `json:"other_phone_numbers"`
	CreatedAt         string                                   `json:"created_at"`
	UpdatedAt         string                                   `json:"updated_at"`
	CustomFields      map[string]interface{}                   `json:"custom_fields"`
}

type taskGetContactResponseOtherCompany struct {
	CompanyID int64 `json:"company_id"`
}

type taskGetContactResponseOtherPhoneNumber struct {
	PhoneNumber string `json:"value"`
}

type TaskGetContactOutput struct {
	Name              string                 `json:"name"`
	Email             string                 `json:"email,omitempty"`
	Phone             string                 `json:"phone,omitempty"`
	Mobile            string                 `json:"mobile,omitempty"`
	Description       string                 `json:"description,omitempty"`
	Address           string                 `json:"address,omitempty"`
	JobTitle          string                 `json:"job-title,omitempty"`
	Tags              []string               `json:"tags"`
	Language          string                 `json:"language,omitempty"`
	TimeZone          string                 `json:"time-zone,omitempty"`
	CompanyID         int64                  `json:"company-id,omitempty"`
	UniqueExternalID  string                 `json:"unique-external-id,omitempty"`
	TwitterID         string                 `json:"twitter-id,omitempty"`
	ViewAllTickets    bool                   `json:"view-all-tickets"`
	Deleted           bool                   `json:"deleted"`
	Active            bool                   `json:"active"`
	OtherEmails       []string               `json:"other-emails"`
	OtherCompaniesIDs []int64                `json:"other-companies-ids"`
	OtherPhoneNumbers []string               `json:"other-phone-numbers"`
	CreatedAt         string                 `json:"created-at"`
	UpdatedAt         string                 `json:"updated-at"`
	CustomFields      map[string]interface{} `json:"custom-fields,omitempty"`
}

func (e *execution) TaskGetContact(in *structpb.Struct) (*structpb.Struct, error) {

	inputStruct := TaskGetContactInput{}
	err := base.ConvertFromStructpb(in, &inputStruct)

	if err != nil {
		return nil, fmt.Errorf("failed to convert input to struct: %v", err)
	}

	resp, err := e.client.GetContact(inputStruct.ContactID)

	if err != nil {
		return nil, err
	}

	outputStruct := TaskGetContactOutput{
		Name:             resp.Name,
		Email:            resp.Email,
		Phone:            resp.Phone,
		Mobile:           resp.Mobile,
		Description:      resp.Description,
		Address:          resp.Address,
		JobTitle:         resp.JobTitle,
		Tags:             resp.Tags,
		Language:         resp.Language,
		TimeZone:         resp.TimeZone,
		CompanyID:        resp.CompanyID,
		UniqueExternalID: resp.UniqueExternalID,
		TwitterID:        resp.TwitterID,
		ViewAllTickets:   resp.ViewAllTickets,
		Deleted:          resp.Deleted,
		Active:           resp.Active,
		OtherEmails:      *checkForNilString(&resp.OtherEmails),
		CreatedAt:        resp.CreatedAt,
		UpdatedAt:        resp.UpdatedAt,
	}

	if len(resp.OtherCompanies) > 0 {
		outputStruct.OtherCompaniesIDs = make([]int64, len(resp.OtherCompanies))
		for index, company := range resp.OtherCompanies {
			outputStruct.OtherCompaniesIDs[index] = company.CompanyID
		}
	} else {
		outputStruct.OtherCompaniesIDs = []int64{}
	}

	if len(resp.OtherPhoneNumbers) > 0 {
		outputStruct.OtherPhoneNumbers = make([]string, len(resp.OtherPhoneNumbers))
		for index, phone := range resp.OtherPhoneNumbers {
			outputStruct.OtherPhoneNumbers[index] = phone.PhoneNumber
		}
	} else {
		outputStruct.OtherPhoneNumbers = []string{}
	}

	if len(resp.CustomFields) > 0 {
		outputStruct.CustomFields = resp.CustomFields
	}

	output, err := base.ConvertToStructpb(outputStruct)

	if err != nil {
		return nil, fmt.Errorf("failed to convert output to struct: %v", err)
	}

	return output, nil
}
