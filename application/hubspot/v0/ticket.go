package hubspot

import (
	"strings"

	hubspot "github.com/belong-inc/go-hubspot"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

// following go-hubspot sdk format

// API functions for Ticket

type TicketService interface {
	Get(ticketId string) (*hubspot.ResponseResource, error)
	Create(ticket *TaskCreateTicketReq) (*hubspot.ResponseResource, error)
}

type TicketServiceOp struct {
	client     *hubspot.Client
	ticketPath string
}

var ticketProperties = []string{
	"hubspot_owner_id",
	"subject",
	"hs_pipeline_stage",
	"hs_pipeline",
	"hs_ticket_category",
	"hs_ticket_priority",
	"source_type",
	"hs_object_source_label",
	"createdate",
	"hs_lastmodifieddate",
}

func (s *TicketServiceOp) Get(ticketId string) (*hubspot.ResponseResource, error) {
	resource := &hubspot.ResponseResource{Properties: &TaskGetTicketResp{}}
	option := &hubspot.RequestQueryOption{Properties: ticketProperties, Associations: []string{"contacts"}}
	if err := s.client.Get(s.ticketPath+"/"+ticketId, resource, option); err != nil {
		return nil, err
	}

	return resource, nil
}

func (s *TicketServiceOp) Create(ticket *TaskCreateTicketReq) (*hubspot.ResponseResource, error) {
	req := &hubspot.RequestPayload{Properties: ticket}
	resource := &hubspot.ResponseResource{Properties: ticket}
	if err := s.client.Post(s.ticketPath, req, resource); err != nil {
		return nil, err
	}
	return resource, nil
}

// Get Ticket

type TaskGetTicketInput struct {
	TicketId string `json:"ticket-id"`
}

type TaskGetTicketResp struct {
	OwnerId          string `json:"hubspot_owner_id,omitempty"`
	TicketName       string `json:"subject"`
	TicketStatus     string `json:"hs_pipeline_stage"`
	Pipeline         string `json:"hs_pipeline"`
	Category         string `json:"hs_ticket_category,omitempty"`
	Priority         string `json:"hs_ticket_priority,omitempty"`
	Source           string `json:"source_type,omitempty"`
	RecordSource     string `json:"hs_object_source_label,omitempty"`
	CreateDate       string `json:"createdate"`
	LastModifiedDate string `json:"hs_lastmodifieddate"`
	TicketId         string `json:"hs_object_id"`
}

type TaskGetTicketOutput struct {
	OwnerId              string   `json:"owner-id,omitempty"`
	TicketName           string   `json:"ticket-name"`
	TicketStatus         string   `json:"ticket-status"`
	Pipeline             string   `json:"pipeline"`
	Category             []string `json:"category,omitempty"`
	Priority             string   `json:"priority,omitempty"`
	Source               string   `json:"source,omitempty"`
	RecordSource         string   `json:"record-source,omitempty"`
	CreateDate           string   `json:"create-date"`
	LastModifiedDate     string   `json:"last-modified-date"`
	AssociatedContactIds []string `json:"associated-contact-ids,omitempty"`
}

func (e *execution) GetTicket(input *structpb.Struct) (*structpb.Struct, error) {

	inputStruct := TaskGetTicketInput{}

	err := base.ConvertFromStructpb(input, &inputStruct)

	if err != nil {
		return nil, err
	}

	res, err := e.client.Ticket.Get(inputStruct.TicketId)
	if err != nil {
		return nil, err
	}

	ticketInfo := res.Properties.(*TaskGetTicketResp)

	// get contacts associated with ticket

	var ticketContactList []string
	if res.Associations != nil {
		ticketContactAssociation := res.Associations.Contacts.Results
		ticketContactList = make([]string, len(ticketContactAssociation))
		for index, value := range ticketContactAssociation {
			ticketContactList[index] = value.ID
		}
	}

	var categoryValues []string
	if ticketInfo.Category != "" {
		categoryValues = strings.Split(ticketInfo.Category, ";")
	}

	outputStruct := TaskGetTicketOutput{
		OwnerId:              ticketInfo.OwnerId,
		TicketName:           ticketInfo.TicketName,
		TicketStatus:         ticketInfo.TicketStatus,
		Pipeline:             ticketInfo.Pipeline,
		Category:             categoryValues,
		Priority:             ticketInfo.Priority,
		Source:               ticketInfo.Source,
		RecordSource:         ticketInfo.RecordSource,
		CreateDate:           ticketInfo.CreateDate,
		LastModifiedDate:     ticketInfo.LastModifiedDate,
		AssociatedContactIds: ticketContactList,
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// Create Ticket
type TaskCreateTicketInput struct {
	OwnerId                   string   `json:"owner-id"`
	TicketName                string   `json:"ticket-name"`
	TicketStatus              string   `json:"ticket-status"`
	Pipeline                  string   `json:"pipeline"`
	Category                  []string `json:"category"`
	Priority                  string   `json:"priority"`
	Source                    string   `json:"source"`
	CreateContactsAssociation []string `json:"create-contacts-association"`
}

type TaskCreateTicketReq struct {
	OwnerId      string `json:"hubspot_owner_id,omitempty"`
	TicketName   string `json:"subject"`
	TicketStatus string `json:"hs_pipeline_stage"`
	Pipeline     string `json:"hs_pipeline"`
	Category     string `json:"hs_ticket_category,omitempty"`
	Priority     string `json:"hs_ticket_priority,omitempty"`
	Source       string `json:"source_type,omitempty"`
	TicketId     string `json:"hs_object_id"`
}

type TaskCreateTicketOutput struct {
	TicketId string `json:"ticket-id"`
}

func (e *execution) CreateTicket(input *structpb.Struct) (*structpb.Struct, error) {

	inputStruct := TaskCreateTicketInput{}
	err := base.ConvertFromStructpb(input, &inputStruct)

	if err != nil {
		return nil, err
	}

	req := TaskCreateTicketReq{
		OwnerId:      inputStruct.OwnerId,
		TicketName:   inputStruct.TicketName,
		TicketStatus: inputStruct.TicketStatus,
		Pipeline:     inputStruct.Pipeline,
		Category:     strings.Join(inputStruct.Category, ";"),
		Priority:     inputStruct.Priority,
		Source:       inputStruct.Source,
	}

	res, err := e.client.Ticket.Create(&req)

	if err != nil {
		return nil, err
	}

	// get ticket Id
	ticketId := res.Properties.(*TaskCreateTicketReq).TicketId

	outputStruct := TaskCreateTicketOutput{TicketId: ticketId}

	output, err := base.ConvertToStructpb(outputStruct)

	if err != nil {
		return nil, err
	}

	// This section is for creating associations (ticket -> object)
	if len(inputStruct.CreateContactsAssociation) != 0 {
		err := CreateAssociation(&outputStruct.TicketId, &inputStruct.CreateContactsAssociation, "ticket", "contact", e)

		if err != nil {
			return nil, err
		}
	}

	return output, nil
}
