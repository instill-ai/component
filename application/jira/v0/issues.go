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
	Summary     string                 `json:"summary"`
	Fields      map[string]interface{} `json:"fields"`
	Self        string                 `json:"self"`
	IssueType   string                 `json:"issue-type"`
	Status      string                 `json:"status"`
}

type GetIssueInput struct {
	IssueKeyOrID  string `json:"issue-id-or-key,omitempty" struct:"issueIdOrKey"`
	UpdateHistory bool   `json:"update-history,omitempty" struct:"updateHistory"`
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

	apiEndpoint := fmt.Sprintf("rest/agile/1.0/issue/%s", opt.IssueKeyOrID)
	req := jiraClient.Client.R().SetResult(&GetIssueOutput{})

	opt.IssueKeyOrID = "" // Remove from query params
	err := addQueryOptions(req, opt)
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

	issue, ok := resp.Result().(*GetIssueOutput)
	if !ok {
		return nil, errmsg.AddMessage(
			fmt.Errorf("failed to convert response to `Get Issue` Output"),
			fmt.Sprintf("failed to convert %v to `Get Issue` Output", resp.Result()),
		)
	}

	if issue.Description == "" && issue.Fields["description"] != nil {
		if issue.Description, ok = issue.Fields["description"].(string); !ok {
			issue.Description = ""
		}
	}
	if issue.Summary == "" && issue.Fields["summary"] != nil {
		if issue.Summary, ok = issue.Fields["summary"].(string); !ok {
			issue.Summary = ""
		}
	}
	if issue.IssueType == "" && issue.Fields["issuetype"] != nil {
		if issueType, ok := issue.Fields["issuetype"]; ok {
			if issue.IssueType, ok = issueType.(map[string]interface{})["name"].(string); !ok {
				issue.IssueType = ""
			}
		}
	}
	if issue.Status == "" && issue.Fields["status"] != nil {
		if status, ok := issue.Fields["status"]; ok {
			if issue.Status, ok = status.(map[string]interface{})["name"].(string); !ok {
				issue.Status = ""
			}
		}
	}
	return base.ConvertToStructpb(issue)
}
