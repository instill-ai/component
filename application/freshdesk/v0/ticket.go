package freshdesk

import (
	"fmt"
	"strings"

	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	TicketPath = "tickets"
)

// API functions for Ticket

func (c *FreshdeskClient) GetTicket(ticketID int64) (*TaskGetTicketResponse, error) {

	resp := &TaskGetTicketResponse{}

	httpReq := c.httpclient.R().SetResult(resp)
	if _, err := httpReq.Get(fmt.Sprintf("/%s/%d", TicketPath, ticketID)); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *FreshdeskClient) CreateTicket(req *TaskCreateTicketReq) (*TaskCreateTicketResponse, error) {
	resp := &TaskCreateTicketResponse{}
	httpReq := c.httpclient.R().SetBody(req).SetResult(resp)
	if _, err := httpReq.Post("/" + TicketPath); err != nil {
		return nil, err
	}
	return resp, nil
}

//Task 1: Get Ticket

type TaskGetTicketInput struct {
	TicketID int64 `json:"ticket-id"`
}

type TaskGetTicketResponse struct {
	Subject                string                          `json:"subject"`
	DescriptionText        string                          `json:"description_text"`
	Source                 int                             `json:"source"`
	Status                 int                             `json:"status"`
	Priority               int                             `json:"priority"`
	TicketType             string                          `json:"type"`
	AssociationType        int                             `json:"association_type"`
	AssociatedTicketList   []int                           `json:"associated_tickets_list"`
	Tags                   []string                        `json:"tags"`
	CCEmails               []string                        `json:"cc_emails"`
	ForwardEmails          []string                        `json:"fwd_emails"`
	ReplyCCEmails          []string                        `json:"reply_cc_emails"`
	RequesterID            int64                           `json:"requester_id"`
	ResponderID            int64                           `json:"responder_id"`
	CompanyID              int64                           `json:"company_id"`
	GroupID                int64                           `json:"group_id"`
	ProductID              int64                           `json:"product_id"`
	SupportEmail           string                          `json:"support_email"`
	ToEmails               []string                        `json:"to_emails"`
	Spam                   bool                            `json:"spam"`
	IsEscalated            bool                            `json:"is_escalated"`
	DueBy                  string                          `json:"due_by"`
	FirstResponseDueBy     string                          `json:"fr_due_by"`
	FirstResponseEscalated bool                            `json:"fr_escalated"`
	NextResponseDueBy      string                          `json:"nr_due_by"`
	NextResponseEscalated  bool                            `json:"nr_escalated"`
	CreatedAt              string                          `json:"created_at"`
	UpdatedAt              string                          `json:"updated_at"`
	Attachments            []taskGetTicketOutputAttachment `json:"attachments"`
	SentimentScore         int                             `json:"sentiment_score"`
	InitialSentimentScore  int                             `json:"initial_sentiment_score"`
}

type TaskGetTicketOutput struct {
	Subject                string                          `json:"subject"`
	DescriptionText        string                          `json:"description-text"`
	Source                 string                          `json:"source"`
	Status                 string                          `json:"status"`
	Priority               string                          `json:"priority"`
	TicketType             string                          `json:"ticket-type,omitempty"`
	AssociationType        string                          `json:"association-type"`
	AssociatedTicketList   []int                           `json:"associated-ticket-list,omitempty"`
	Tags                   []string                        `json:"tags"`
	CCEmails               []string                        `json:"cc-emails"`
	ForwardEmails          []string                        `json:"forward-emails"`
	ReplyCCEmails          []string                        `json:"reply-cc-emails"`
	RequesterID            int64                           `json:"requester-id"`
	ResponderID            int64                           `json:"responder-id,omitempty"`
	CompanyID              int64                           `json:"company-id,omitempty"`
	GroupID                int64                           `json:"group-id,omitempty"`
	ProductID              int64                           `json:"product-id,omitempty"`
	SupportEmail           string                          `json:"support-email,omitempty"`
	ToEmails               []string                        `json:"to-emails"`
	Spam                   bool                            `json:"spam"`
	DueBy                  string                          `json:"due-by,omitempty"`
	IsEscalated            bool                            `json:"is-escalated"`
	FirstResponseDueBy     string                          `json:"first-response-due-by,omitempty"`
	FirstResponseEscalated bool                            `json:"first-response-escalated,omitempty"`
	NextResponseDueBy      string                          `json:"next-response-due-by,omitempty"`
	NextResponseEscalated  bool                            `json:"next-response-escalated,omitempty"`
	CreatedAt              string                          `json:"created-at"`
	UpdatedAt              string                          `json:"updated-at"`
	Attachments            []taskGetTicketOutputAttachment `json:"attachments,omitempty"`
	SentimentScore         int                             `json:"sentiment-score"`
	InitialSentimentScore  int                             `json:"initial-sentiment-score"`
}

type taskGetTicketOutputAttachment struct {
	Name        string `json:"name"`
	ContentType string `json:"content-type"`
	URL         string `json:"url"`
}

func (e *execution) TaskGetTicket(in *structpb.Struct) (*structpb.Struct, error) {
	inputStruct := TaskGetTicketInput{}
	err := base.ConvertFromStructpb(in, &inputStruct)

	if err != nil {
		return nil, fmt.Errorf("failed to convert input to struct: %v", err)
	}

	resp, err := e.client.GetTicket(inputStruct.TicketID)
	if err != nil {
		return nil, err
	}

	outputStruct := TaskGetTicketOutput{
		Subject:                resp.Subject,
		DescriptionText:        resp.DescriptionText,
		Source:                 convertSourceToString(resp.Source),
		Status:                 convertStatusToString(resp.Status),
		Priority:               convertPriorityToString(resp.Priority),
		TicketType:             resp.TicketType,
		AssociationType:        convertAssociationType(resp.AssociationType),
		AssociatedTicketList:   resp.AssociatedTicketList,
		Tags:                   *checkForNil(&resp.Tags),
		CCEmails:               *checkForNil(&resp.CCEmails),
		ForwardEmails:          *checkForNil(&resp.ForwardEmails),
		ReplyCCEmails:          *checkForNil(&resp.ReplyCCEmails),
		RequesterID:            resp.RequesterID,
		ResponderID:            resp.ResponderID,
		CompanyID:              resp.CompanyID,
		GroupID:                resp.GroupID,
		ProductID:              resp.ProductID,
		SupportEmail:           resp.SupportEmail,
		ToEmails:               *checkForNil(&resp.ToEmails),
		Spam:                   resp.Spam,
		DueBy:                  convertTimestampResp(resp.DueBy),
		IsEscalated:            resp.IsEscalated,
		FirstResponseDueBy:     convertTimestampResp(resp.FirstResponseDueBy),
		FirstResponseEscalated: resp.FirstResponseEscalated,
		NextResponseDueBy:      convertTimestampResp(resp.NextResponseDueBy),
		NextResponseEscalated:  resp.NextResponseEscalated,
		CreatedAt:              convertTimestampResp(resp.CreatedAt),
		UpdatedAt:              convertTimestampResp(resp.UpdatedAt),
		Attachments:            resp.Attachments,
		SentimentScore:         resp.SentimentScore,
		InitialSentimentScore:  resp.InitialSentimentScore,
	}

	output, err := base.ConvertToStructpb(outputStruct)

	if err != nil {
		return nil, fmt.Errorf("failed to convert output to struct: %v", err)
	}

	return output, nil
}

// Create Ticket
type TaskCreateTicketInput struct {
	// Only one is needed: requester-id or email
	RequesterID      int64    `json:"requester-id"`
	Email            string   `json:"email"`
	Subject          string   `json:"subject"`
	Description      string   `json:"description-text"`
	Source           string   `json:"source"`
	Status           string   `json:"status"`
	Priority         string   `json:"priority"`
	Type             string   `json:"ticket-type"`
	CompanyID        int64    `json:"company-id"`
	ProductID        int64    `json:"product-id"`
	GroupID          int64    `json:"group-id"`
	ResponderID      int64    `json:"responder-id"`
	Tags             []string `json:"tags"`
	CCEmails         []string `json:"cc-emails"`
	ParentID         int64    `json:"parent-id"`
	RelatedTicketIDs []int64  `json:"related-ticket-ids"`
}

type TaskCreateTicketReq struct {
	RequesterID      int64    `json:"requester_id,omitempty"`
	Email            string   `json:"email,omitempty"`
	Subject          string   `json:"subject"`
	Description      string   `json:"description"`
	Source           int      `json:"source"`
	Status           int      `json:"status"`
	Priority         int      `json:"priority"`
	Type             string   `json:"type,omitempty"`
	CompanyID        int64    `json:"company_id,omitempty"`
	ProductID        int64    `json:"product_id,omitempty"`
	GroupID          int64    `json:"group_id,omitempty"`
	ResponderID      int64    `json:"responder_id,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	CCEmails         []string `json:"cc_emails,omitempty"`
	ParentID         int64    `json:"parent_id,omitempty"`
	RelatedTicketIDs []int64  `json:"related_ticket_ids,omitempty"`
}

type TaskCreateTicketResponse struct {
	ID        int64  `json:"id"`
	CreatedAt string `json:"created_at"`
}

type TaskCreateTicketOutput struct {
	ID        int64  `json:"ticket-id"`
	CreatedAt string `json:"created-at"`
}

func (e *execution) TaskCreateTicket(in *structpb.Struct) (*structpb.Struct, error) {
	inputStruct := TaskCreateTicketInput{}
	err := base.ConvertFromStructpb(in, &inputStruct)

	if err != nil {
		return nil, fmt.Errorf("failed to convert input to struct: %v", err)
	}

	if inputStruct.ParentID != 0 && len(inputStruct.RelatedTicketIDs) > 0 {
		return nil, fmt.Errorf("only one of parent-id or related-ticket-ids can be provided")
	}

	req := TaskCreateTicketReq{
		Subject:     inputStruct.Subject,
		Description: inputStruct.Description,
		Source:      convertSourceToInt(inputStruct.Source),
		Status:      convertStatusToInt(inputStruct.Status),
		Priority:    convertPriorityToInt(inputStruct.Priority),
		Type:        inputStruct.Type,
		CompanyID:   inputStruct.CompanyID,
		ProductID:   inputStruct.ProductID,
		GroupID:     inputStruct.GroupID,
		ResponderID: inputStruct.ResponderID,
		Tags:        inputStruct.Tags,
		CCEmails:    inputStruct.CCEmails,
	}

	if inputStruct.RequesterID != 0 {
		req.RequesterID = inputStruct.RequesterID
	} else if inputStruct.Email != "" {
		req.Email = inputStruct.Email
	} else {
		return nil, fmt.Errorf("either Requester ID or email is required")
	}

	if inputStruct.ParentID != 0 {
		req.ParentID = inputStruct.ParentID
	}

	if len(inputStruct.RelatedTicketIDs) > 0 {
		req.RelatedTicketIDs = inputStruct.RelatedTicketIDs
	}

	resp, err := e.client.CreateTicket(&req)
	if err != nil {
		return nil, err
	}

	outputStruct := TaskCreateTicketOutput{
		ID:        resp.ID,
		CreatedAt: convertTimestampResp(resp.CreatedAt),
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to convert output to struct: %v", err)
	}

	return output, nil
}

func convertTimestampResp(timestamp string) string {
	// freshdesk response timestamp is always in the format of YYYY-MM-DDTHH:MM:SSZ and in UTC.
	// this function will convert it to YYYY-MM-DD HH:MM:SS UTC

	if timestamp == "" {
		return timestamp
	}
	formattedTime := strings.Replace(timestamp, "T", " ", 1)
	formattedTime = strings.Replace(formattedTime, "Z", " ", 1)
	formattedTime += "UTC"

	return formattedTime
}

func checkForNil(input *[]string) *[]string {
	if *input == nil {
		return &[]string{}
	}
	return input
}

func convertSourceToString(source int) string {
	switch source {
	case 1:
		return "Email"
	case 2:
		return "Portal"
	case 3:
		return "Phone"
	case 4:
		return "Forum"
	case 5:
		return "Twitter"
	case 6:
		return "Facebook"
	case 7:
		return "Chat"
	case 8:
		return "MobiHelp"
	case 9:
		return "Feedback Widget"
	case 10:
		return "Outbound Email"
	case 11:
		return "Ecommerce"
	case 12:
		return "Bot"
	case 13:
		return "Whatsapp"
	default:
		return fmt.Sprintf("Unknown source, received: %d", source)
	}
}

func convertSourceToInt(source string) int {
	// For creating ticket, the only source that can be used is 1,2,3,5,6,7,9,11,10
	switch source {
	case "Email":
		return 1
	case "Portal":
		return 2
	case "Phone":
		return 3
	case "Twitter":
		return 5
	case "Facebook":
		return 6
	case "Chat":
		return 7
	case "Feedback Widget":
		return 9
	case "Outbound Email":
		return 10
	case "Ecommerce":
		return 11
	}
	return 0
}

func convertStatusToString(status int) string {
	switch status {
	case 2:
		return "Open"
	case 3:
		return "Pending"
	case 4:
		return "Resolved"
	case 5:
		return "Closed"
	case 6:
		return "Waiting on Customer"
	case 7:
		return "Waiting on Third Party"
	default:
		return fmt.Sprintf("Unknown status, received: %d", status)
	}
}

func convertStatusToInt(status string) int {
	switch status {
	case "Open":
		return 2
	case "Pending":
		return 3
	case "Resolved":
		return 4
	case "Closed":
		return 5
	case "Waiting on Customer":
		return 6
	case "Waiting on Third Party":
		return 7
	}
	return 0
}

func convertPriorityToString(priority int) string {
	switch priority {
	case 1:
		return "Low"
	case 2:
		return "Medium"
	case 3:
		return "High"
	case 4:
		return "Urgent"
	default:
		return fmt.Sprintf("Unknown priority, received: %d", priority)
	}
}

func convertPriorityToInt(priority string) int {
	switch priority {
	case "Low":
		return 1
	case "Medium":
		return 2
	case "High":
		return 3
	case "Urgent":
		return 4
	}
	return 0
}

func convertAssociationType(associationType int) string {
	switch associationType {
	case 1:
		return "Parent"
	case 2:
		return "Child"
	case 3:
		return "Tracker"
	case 4:
		return "Related"
	default:
		return "No association"
	}
}
