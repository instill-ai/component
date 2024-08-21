package freshdesk

import (
	"fmt"

	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	AgentPath = "agents"
)

// API function for Agent

func (c *FreshdeskClient) GetAgent(agentID int64) (*TaskGetAgentResponse, error) {
	resp := &TaskGetAgentResponse{}

	httpReq := c.httpclient.R().SetResult(resp)
	if _, err := httpReq.Get(fmt.Sprintf("/%s/%d", AgentPath, agentID)); err != nil {
		return nil, err
	}
	return resp, nil
}

// Task 1: Get Agent

type TaskGetAgentInput struct {
	AgentID int64 `json:"agent-id"`
}

type TaskGetAgentResponse struct {
	Contact     taskGetAgentResponseContact `json:"contact"`
	Type        string                      `json:"type"`
	TicketScope int                         `json:"ticket_scope"`
	Available   bool                        `json:"available"`
	GroupIDs    []int64                     `json:"group_ids"`
	RoleIDs     []int64                     `json:"role_ids"`
	SkillIDs    []int64                     `json:"skill_ids"`
	Occasional  bool                        `json:"occasional"`
	Signature   string                      `json:"signature"`
	FocusMode   bool                        `json:"focus_mode"`
	Deactivated bool                        `json:"deactivated"`
	CreatedAt   string                      `json:"created_at"`
	UpdatedAt   string                      `json:"updated_at"`
}

type taskGetAgentResponseContact struct {
	Name        string `json:"name"`
	Active      bool   `json:"active"`
	Email       string `json:"email"`
	JobTitle    string `json:"job_title"`
	Language    string `json:"language"`
	LastLoginAt string `json:"last_login_at"`
	Mobile      string `json:"mobile"`
	Phone       string `json:"phone"`
	TimeZone    string `json:"time_zone"`
}

type TaskGetAgentOutput struct {
	Name        string  `json:"name,omitempty"`
	Active      bool    `json:"active"`
	Email       string  `json:"email"`
	JobTitle    string  `json:"job-title,omitempty"`
	Language    string  `json:"language,omitempty"`
	LastLoginAt string  `json:"last-login-at"`
	Mobile      string  `json:"mobile,omitempty"`
	Phone       string  `json:"phone,omitempty"`
	TimeZone    string  `json:"time-zone,omitempty"`
	Type        string  `json:"type"`
	TicketScope string  `json:"ticket-scope"`
	Available   bool    `json:"available"`
	GroupIDs    []int64 `json:"group-ids"`
	RoleIDs     []int64 `json:"role-ids"`
	SkillIDs    []int64 `json:"skill-ids"`
	Occasional  bool    `json:"occasional"`
	Signature   string  `json:"signature,omitempty"`
	FocusMode   bool    `json:"focus-mode"`
	Deactivated bool    `json:"deactivated"`
	CreatedAt   string  `json:"created-at"`
	UpdatedAt   string  `json:"updated-at"`
}

func (e *execution) TaskGetAgent(in *structpb.Struct) (*structpb.Struct, error) {
	inputStruct := TaskGetAgentInput{}
	err := base.ConvertFromStructpb(in, &inputStruct)

	if err != nil {
		return nil, fmt.Errorf("failed to convert input to struct: %v", err)
	}

	resp, err := e.client.GetAgent(inputStruct.AgentID)

	if err != nil {
		return nil, err
	}

	outputStruct := TaskGetAgentOutput{
		Name:        resp.Contact.Name,
		Active:      resp.Contact.Active,
		Email:       resp.Contact.Email,
		JobTitle:    resp.Contact.JobTitle,
		Language:    convertCodeToLanguage(resp.Contact.Language),
		LastLoginAt: convertTimestampResp(resp.Contact.LastLoginAt),
		Mobile:      resp.Contact.Mobile,
		Phone:       resp.Contact.Phone,
		TimeZone:    resp.Contact.TimeZone,
		Type:        convertTypeResponse(resp.Type),
		TicketScope: convertTicketScopeResponse(resp.TicketScope),
		Available:   resp.Available,
		GroupIDs:    resp.GroupIDs,
		RoleIDs:     resp.RoleIDs,
		SkillIDs:    resp.SkillIDs,
		Occasional:  resp.Occasional,
		Signature:   resp.Signature,
		FocusMode:   resp.FocusMode,
		Deactivated: resp.Deactivated,
		CreatedAt:   convertTimestampResp(resp.CreatedAt),
		UpdatedAt:   convertTimestampResp(resp.UpdatedAt),
	}

	output, err := base.ConvertToStructpb(outputStruct)

	if err != nil {
		return nil, fmt.Errorf("failed to convert output to struct: %v", err)
	}

	return output, nil
}

func convertTypeResponse(in string) string {
	switch in {
	case "support_agent":
		return "Support Agent"
	case "field_agent":
		return "Field Agent"
	case "collaborator":
		return "Collaborator"
	default:
		return in
	}
}

func convertTicketScopeResponse(in int) string {
	switch in {
	case 1:
		return "Global Access"
	case 2:
		return "Group Access"
	case 3:
		return "Restricted Access"
	default:
		return "Unknown"
	}
}
