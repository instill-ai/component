package jira

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/x/errmsg"
	"google.golang.org/protobuf/types/known/structpb"
)

type Sprint struct {
	ID            int    `json:"id"`
	Self          string `json:"self"`
	State         string `json:"state"`
	Name          string `json:"name"`
	StartDate     string `json:"startDate"`
	EndDate       string `json:"endDate"`
	CompleteDate  string `json:"completeDate"`
	OriginBoardID int    `json:"originBoardId"`
	Goal          string `json:"goal"`
}

type GetSprintInput struct {
	SprintID int `json:"sprint-id,omitempty" struct:"sprintId"`
}
type GetSprintOutput struct {
	ID            int    `json:"id"`
	Self          string `json:"self"`
	State         string `json:"state"`
	Name          string `json:"name"`
	StartDate     string `json:"start-date"`
	EndDate       string `json:"end-date"`
	CompleteDate  string `json:"complete-date"`
	OriginBoardID int    `json:"origin-board-id"`
	Goal          string `json:"goal"`
}

func (jiraClient *Client) extractSprintOutput(sprint *Sprint) *GetSprintOutput {
	return &GetSprintOutput{
		ID:            sprint.ID,
		Self:          sprint.Self,
		State:         sprint.State,
		Name:          sprint.Name,
		StartDate:     sprint.StartDate,
		EndDate:       sprint.EndDate,
		CompleteDate:  sprint.CompleteDate,
		OriginBoardID: sprint.OriginBoardID,
		Goal:          sprint.Goal,
	}
}
func (jiraClient *Client) getSprintTask(_ context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug DebugSession
	debug.SessionStart("getSprintTask", StaticVerboseLevel)
	defer debug.SessionEnd()

	var opt GetSprintInput
	if err := base.ConvertFromStructpb(props, &opt); err != nil {
		return nil, err
	}
	debug.AddMessage(fmt.Sprintf("GetSprintInput: %+v", opt))

	apiEndpoint := fmt.Sprintf("rest/agile/1.0/sprint/%v", opt.SprintID)
	req := jiraClient.Client.R().SetResult(&Sprint{})
	resp, err := req.Get(apiEndpoint)

	if resp != nil && resp.StatusCode() == 404 {
		return nil, fmt.Errorf(
			err.Error(),
			errmsg.Message(err)+"Please check you have the correct permissions to access this resource.",
		)
	} else if resp != nil && resp.StatusCode() == 401 {
		return nil, fmt.Errorf(
			err.Error(),
			errmsg.Message(err)+"You are not logged in. Please provide a valid token and an email account.",
		)
	}
	if err != nil {
		return nil, fmt.Errorf(
			err.Error(), errmsg.Message(err),
		)
	}
	debug.AddMessage("GET", apiEndpoint)
	debug.AddMapMessage("QueryParam", resp.Request.QueryParam)
	debug.AddMessage("Status", resp.Status())

	issue, ok := resp.Result().(*Sprint)
	if !ok {
		return nil, errmsg.AddMessage(
			fmt.Errorf("failed to convert response to `Get Sprint` Output"),
			fmt.Sprintf("failed to convert %v to `Get Sprint` Output", resp.Result()),
		)
	}
	out := jiraClient.extractSprintOutput(issue)
	return base.ConvertToStructpb(out)
}
