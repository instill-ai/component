package jira

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/x/errmsg"
	"google.golang.org/protobuf/types/known/structpb"
)

type Issue struct {
	ID          string                 `json:"id"`
	Key         string                 `json:"key"`
	Description string                 `json:"description"`
	Fields      map[string]interface{} `json:"fields"`
	Self        string                 `json:"self"`
}

type GetIssueInput struct {
	IssueKeyOrID  string `json:"issue_id_or_key,omitempty" struct:"issueIdOrKey"`
	UpdateHistory bool   `json:"update_history,omitempty" struct:"updateHistory"`
}
type GetIssueOutput struct {
	Issue
}

func (jiraClient *Client) getIssueTask(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug DebugSession
	debug.SessionStart("getIssueTask", StaticVerboseLevel)
	defer debug.SessionEnd()

	var opt GetIssueInput
	if err := base.ConvertFromStructpb(props, &opt); err != nil {
		return nil, err
	}
	debug.AddMessage(fmt.Sprintf("GetIssueInput: %+v", opt))
	debug.AddMapMessage("opt", opt)
	issue, err := jiraClient.getIssue(ctx, &opt)
	if err != nil {
		return nil, err
	}
	return base.ConvertToStructpb(issue)
}

func (jiraClient *Client) getIssue(_ context.Context, opt *GetIssueInput) (*GetIssueOutput, error) {
	var debug DebugSession
	debug.SessionStart("getIssue", StaticVerboseLevel)
	defer debug.SessionEnd()

	apiEndpoint := fmt.Sprintf("rest/agile/1.0/issue/%s", opt.IssueKeyOrID)

	req := jiraClient.Client.R().SetResult(&GetIssueOutput{})

	debug.AddMapMessage("opt", *opt)

	opt.IssueKeyOrID = "" // Remove from query params
	err := addQueryOptions(req, *opt)
	if err != nil {
		return nil, err
	}
	resp, err := req.Get(apiEndpoint)

	if resp != nil && resp.StatusCode() == 404 {
		return nil, fmt.Errorf(
			err.Error(),
			errmsg.Message(err)+"Please check you have the correct permissions to access this resource.",
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

	boards := resp.Result().(*GetIssueOutput)
	return boards, err
}
