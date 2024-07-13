package hubspot

import (
	"strings"

	hubspot "github.com/belong-inc/go-hubspot"
)

type TicketService interface {
	Get(ticketId string) (*hubspot.ResponseResource, error)
	Create(ticket *TicketInfoHSFormat) (*hubspot.ResponseResource, error)
}

type TicketServiceOp struct {
	client     *hubspot.Client
	ticketPath string
}

type TicketInfoHSFormat struct {
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

type TicketInfoTaskFormat struct {
	OwnerId          string   `json:"owner-id,omitempty"`
	TicketName       string   `json:"ticket-name"`
	TicketStatus     string   `json:"ticket-status"`
	Pipeline         string   `json:"pipeline"`
	Category         []string `json:"category,omitempty"`
	Priority         string   `json:"priority,omitempty"`
	Source           string   `json:"source,omitempty"`
	RecordSource     string   `json:"record-source,omitempty"`
	CreateDate       string   `json:"create-date"`
	LastModifiedDate string   `json:"last-modified-date"`
	TicketId         string   `json:"ticket-id"`
}

// for Get function
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
	"hs_object_id",
}

func (s *TicketServiceOp) Get(ticketId string) (*hubspot.ResponseResource, error) {
	resource := &hubspot.ResponseResource{Properties: &TicketInfoHSFormat{}}
	option := &hubspot.RequestQueryOption{Properties: ticketProperties, Associations: []string{"contacts"}}
	if err := s.client.Get(s.ticketPath+"/"+ticketId, resource, option); err != nil {
		return nil, err
	}

	return resource, nil
}

func (s *TicketServiceOp) Create(ticket *TicketInfoHSFormat) (*hubspot.ResponseResource, error) {
	req := &hubspot.RequestPayload{Properties: ticket}
	resource := &hubspot.ResponseResource{Properties: ticket}
	if err := s.client.Post(s.ticketPath, req, resource); err != nil {
		return nil, err
	}
	return resource, nil
}

func TicketConvertToTaskFormat(res *TicketInfoHSFormat) *TicketInfoTaskFormat {

	var categoryValues []string
	if res.Category != "" {
		categoryValues = strings.Split(res.Category, ";")
	}

	ret := &TicketInfoTaskFormat{
		OwnerId:          res.OwnerId,
		TicketName:       res.TicketName,
		TicketStatus:     res.TicketStatus,
		Pipeline:         res.Pipeline,
		Priority:         res.Priority,
		Category:         categoryValues,
		Source:           res.Source,
		RecordSource:     res.RecordSource,
		CreateDate:       res.CreateDate,
		LastModifiedDate: res.LastModifiedDate,
		TicketId:         res.TicketId,
	}

	return ret
}

func TicketConvertToHSFormat(res *TicketInfoTaskFormat) *TicketInfoHSFormat {

	combinedCategoryValues := strings.Join(res.Category, ";")

	ret := &TicketInfoHSFormat{
		TicketName:   res.TicketName,
		TicketStatus: res.TicketStatus,
		Pipeline:     res.Pipeline,
		Priority:     res.Priority,
		Category:     combinedCategoryValues,
		Source:       res.Source,
		RecordSource: res.RecordSource,
		CreateDate:   res.CreateDate,
	}

	return ret

}
