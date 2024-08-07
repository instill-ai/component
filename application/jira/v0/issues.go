package jira

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/tools/logger"
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
	IssueKey      string `json:"issue-key,omitempty" api:"issueIdOrKey"`
	UpdateHistory bool   `json:"update-history,omitempty" api:"updateHistory"`
}
type GetIssueOutput struct {
	Issue
}

func extractIssue(issue *Issue) *Issue {
	if issue.Description == "" && issue.Fields["description"] != nil {
		description, ok := issue.Fields["description"].(string)
		if ok {
			issue.Description = description
		}
	}
	if issue.Summary == "" && issue.Fields["summary"] != nil {
		summary, ok := issue.Fields["summary"].(string)
		if ok {
			issue.Summary = summary
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
	return issue
}

func (jiraClient *Client) getIssueTask(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {

	var opt GetIssueInput
	if err := base.ConvertFromStructpb(props, &opt); err != nil {
		return nil, err
	}

	apiEndpoint := fmt.Sprintf("rest/agile/1.0/issue/%s", opt.IssueKey)
	req := jiraClient.Client.R().SetResult(&Issue{})

	opt.IssueKey = "" // Remove from query params
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
	issue, ok := resp.Result().(*Issue)
	if !ok {
		return nil, errmsg.AddMessage(
			fmt.Errorf("failed to convert response to `Get Issue` Output"),
			fmt.Sprintf("failed to convert %v to `Get Issue` Output", resp.Result()),
		)
	}
	issue = extractIssue(issue)
	issueOutput := GetIssueOutput{Issue: *issue}
	return base.ConvertToStructpb(issueOutput)
}

type Range struct {
	Range      string `json:"range,omitempty"`
	EpicKey    string `json:"epic-key,omitempty"`
	SprintName string `json:"sprint-name,omitempty"`
	JQL        string `json:"jql,omitempty"`
}

type ListIssuesInput struct {
	BoardName  string `json:"board-name,omitempty" api:"boardName"`
	MaxResults int    `json:"max-results,omitempty" api:"maxResults"`
	StartAt    int    `json:"start-at,omitempty" api:"startAt"`
	Range      Range  `json:"range,omitempty"`
}

type ListIssuesResp struct {
	Issues     []Issue `json:"issues"`
	StartAt    int     `json:"startAt"`
	MaxResults int     `json:"maxResults"`
	Total      int     `json:"total"`
}
type ListIssuesOutput struct {
	Issues     []Issue `json:"issues"`
	StartAt    int     `json:"start-at"`
	MaxResults int     `json:"max-results"`
	Total      int     `json:"total"`
}

func (jiraClient *Client) listIssuesTask(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {

	var (
		opt ListIssuesInput
		jql string
	)

	if err := base.ConvertFromStructpb(props, &opt); err != nil {
		return nil, err
	}

	boards, err := jiraClient.listBoards(ctx, &ListBoardsInput{Name: opt.BoardName})
	if err != nil {
		return nil, err
	}
	if len(boards.Values) == 0 {
		return nil, errmsg.AddMessage(
			fmt.Errorf("board not found"),
			fmt.Sprintf("board with name %s not found", opt.BoardName),
		)
	} else if len(boards.Values) > 1 {
		return nil, errmsg.AddMessage(
			fmt.Errorf("multiple boards found"),
			fmt.Sprintf("multiple boards are found with the partial name \"%s\". Please provide a more specific name", opt.BoardName),
		)
	}
	board := boards.Values[0]

	boardDetails, err := jiraClient.getBoard(ctx, board.ID)
	if err != nil {
		return nil, err
	}
	projectKey := boardDetails.Location.ProjectKey
	if projectKey == "" {
		projectKey = strings.Split(board.Name, "-")[0]
	}
	apiEndpoint := fmt.Sprintf("rest/agile/1.0/board/%d", board.ID)
	switch opt.Range.Range {
	case "All":
		// https://developer.atlassian.com/cloud/jira/software/rest/api-group-board/#api-rest-agile-1-0-board-boardid-issue-get
		apiEndpoint = apiEndpoint + "/issue"
	case "Epics only":
		// https://developer.atlassian.com/cloud/jira/software/rest/api-group-board/#api-rest-agile-1-0-board-boardid-epic-get
		apiEndpoint = apiEndpoint + "/epic"
	case "Issues of an epic":
		// API not working: https://developer.atlassian.com/cloud/jira/software/rest/api-group-board/#api-rest-agile-1-0-board-boardid-epic-epicid-issue-get
		// use JQL instead
		jql = fmt.Sprintf("project=\"%s\" AND parent=\"%s\"", projectKey, opt.Range.EpicKey)
	case "Issues of a sprint":
		// API not working: https://developer.atlassian.com/cloud/jira/software/rest/api-group-board/#api-rest-agile-1-0-board-boardid-sprint-sprintid-issue-get
		// use JQL instead
		jql = fmt.Sprintf("project=\"%s\" AND sprint=\"%s\"", projectKey, opt.Range.SprintName)
	case "In backlog only":
		// https://developer.atlassian.com/cloud/jira/software/rest/api-group-board/#api-rest-agile-1-0-board-boardid-backlog-get
		apiEndpoint = apiEndpoint + "/backlog"
	case "Issues without epic assigned":
		// https://developer.atlassian.com/cloud/jira/software/rest/api-group-board/#api-rest-agile-1-0-board-boardid-epic-none-issue-get
		apiEndpoint = apiEndpoint + "/epic/none/issue"
	case "Standard Issues":
		// https://support.atlassian.com/jira-cloud-administration/docs/what-are-issue-types/
		jql = fmt.Sprintf("project=\"%s\" AND issuetype not in (Epic, subtask)", projectKey)
	case "JQL query":
		jql = opt.Range.JQL
	default:
		return nil, errmsg.AddMessage(
			fmt.Errorf("invalid range"),
			fmt.Sprintf("%s is an invalid range", opt.Range.Range),
		)
	}

	var resp *resty.Response
	if jql != "" {
		resp, err = jiraClient.nextGenIssuesSearch(ctx, nextGenSearchRequest{
			JQL:        jql,
			MaxResults: opt.MaxResults,
			StartAt:    opt.StartAt,
		},
		)
	} else {
		req := jiraClient.Client.R().SetResult(&ListIssuesResp{})
		err = addQueryOptions(req, map[string]interface{}{
			"maxResults": opt.MaxResults,
			"startAt":    opt.StartAt,
		})
		if err != nil {
			return nil, err
		}
		resp, err = req.Get(apiEndpoint)
	}

	if err != nil {
		return nil, err
	}

	issues, ok := resp.Result().(*ListIssuesResp)
	if !ok {
		return nil, errmsg.AddMessage(
			fmt.Errorf("failed to convert response to `List Issue` Output"),
			fmt.Sprintf("failed to convert %v to `List Issue` Output", resp.Result()),
		)
	}

	if issues.Issues == nil {
		issues.Issues = []Issue{}
	}

	output := ListIssuesOutput{
		Issues:     issues.Issues,
		StartAt:    issues.StartAt,
		MaxResults: issues.MaxResults,
		Total:      issues.Total,
	}
	for idx, issue := range output.Issues {
		output.Issues[idx] = *extractIssue(&issue)
		if opt.Range.Range == "Epics only" {
			output.Issues[idx].IssueType = "Epic"
		}
	}
	return base.ConvertToStructpb(output)
}

// https://support.atlassian.com/jira-software-cloud/docs/jql-fields/
type nextGenSearchRequest struct {
	JQL        string `json:"jql,omitempty" api:"jql"`
	MaxResults int    `json:"maxResults,omitempty" api:"maxResults"`
	StartAt    int    `json:"startAt,omitempty" api:"startAt"`
}

// https://developer.atlassian.com/cloud/jira/platform/rest/v2/api-group-issue-search/#api-rest-api-2-search-get
// https://developer.atlassian.com/cloud/jira/platform/rest/v2/api-group-issue-search/#api-rest-api-2-search-post
func (jiraClient *Client) nextGenIssuesSearch(_ context.Context, opt nextGenSearchRequest) (*resty.Response, error) {

	var err error
	apiEndpoint := "/rest/api/2/search"

	req := jiraClient.Client.R().SetResult(&ListIssuesResp{})
	var resp *resty.Response
	if len(opt.JQL) < 50 {
		// 50 is an arbitrary number to determine if the JQL is too long to be a query param
		if err := addQueryOptions(req, opt); err != nil {
			return nil, err
		}
		resp, err = req.Get(apiEndpoint)
	} else {
		req.SetBody(opt)
		resp, err = req.Post(apiEndpoint)
	}

	if err != nil {
		return nil, err
	}
	return resp, nil
}

type AdditionalFields struct {
	Add    string `json:"add,omitempty"`
	Copy   string `json:"copy,omitempty"`
	Set    string `json:"set,omitempty"`
	Edit   string `json:"edit,omitempty"`
	Remove string `json:"remove,omitempty"`
}
type CreateIssueInput struct {
	UpdateHistory bool   `json:"update-history"`
	ProjectKey    string `json:"project-key"`
	IssueType     string `json:"issue-type"`
	Status        string `json:"status"`
	Summary       string `json:"summary"`
	Description   string `json:"description"`
	Assignee      string `json:"assignee"`
}
type CreateIssueRequset struct {
	Fields map[string]interface{}        `json:"fields"`
	Update map[string][]AdditionalFields `json:"update"`
}
type CreateIssueResp struct {
	ID         string `json:"id"`
	Key        string `json:"key"`
	Self       string `json:"self"`
	Transition struct {
		Status          string `json:"status"`
		ErrorCollection struct {
			ErrorMessages []string               `json:"errorMessages"`
			Errors        map[string]interface{} `json:"errors"`
		} `json:"errorCollection"`
	} `json:"transition"`
}

type CreateIssueOutput struct {
	Issue
}

func convertCreateIssueRequest(issue *CreateIssueInput) *CreateIssueRequset {
	return &CreateIssueRequset{
		Fields: map[string]interface{}{
			"project": map[string]interface{}{
				"key": issue.ProjectKey,
			},
			"issuetype": map[string]interface{}{
				"name": issue.IssueType,
			},
			"status": map[string]interface{}{
				"name": issue.Status,
			},
			"summary":     issue.Summary,
			"description": issue.Description,
			"assignee": map[string]interface{}{
				"name": issue.Assignee,
			},
		},
	}
}

func (jiraClient *Client) createIssueTask(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("CreateIssueTask", logger.Develop).SessionEnd()

	var issue CreateIssueInput
	if err := base.ConvertFromStructpb(props, &issue); err != nil {
		return nil, err
	}

	apiEndpoint := "rest/api/2/issue"
	req := jiraClient.Client.R().SetResult(&CreateIssueResp{}).SetBody(convertCreateIssueRequest(&issue))
	err := addQueryOptions(req, map[string]interface{}{"updateHistory": issue.UpdateHistory})
	if err != nil {
		return nil, err
	}
	resp, err := req.Post(apiEndpoint)
	if err != nil {
		return nil, err
	}

	createdResult, ok := resp.Result().(*CreateIssueResp)

	if !ok {
		return nil, errmsg.AddMessage(
			fmt.Errorf("failed to convert response to `Create Issue` Output"),
			fmt.Sprintf("failed to convert %v to `Create Issue` Output", resp.Result()),
		)
	}
	debug.Info("Created Issue: ", createdResult)

	getIssueInput, err := base.ConvertToStructpb(GetIssueInput{IssueKey: createdResult.Key, UpdateHistory: issue.UpdateHistory})
	if err != nil {
		return nil, err
	}
	getIssueOutput, err := jiraClient.getIssueTask(ctx, getIssueInput)
	if err != nil {
		return nil, err
	}
	var issueOutput CreateIssueOutput
	if err := base.ConvertFromStructpb(getIssueOutput, &issueOutput); err != nil {
		return nil, err
	}
	return base.ConvertToStructpb(issueOutput)
}

type UpdateIssueInput struct {
	IssueKey    string                        `json:"issue-key"`
	Update      map[string][]AdditionalFields `json:"update"`
	NotifyUsers bool                          `json:"notify-users" api:"notifyUsers"`
}
type UpdateIssueRequset struct {
	Body struct {
		Update map[string][]AdditionalFields `json:"update"`
	}
	Query struct {
		NotifyUsers bool `json:"notify-users" api:"notifyUsers"`
		ReturnIssue bool `json:"return-issue" api:"returnIssue"`
	}
}
type UpdateIssueResp struct {
	Issue
}

type UpdateIssueOutput struct {
	Issue
}

func (jiraClient *Client) updateIssueTask(_ context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("UpdateIssueTask", logger.Develop).SessionEnd()

	var input UpdateIssueInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}

	apiEndpoint := "rest/api/2/issue/" + input.IssueKey
	request := UpdateIssueRequset{
		Body: struct {
			Update map[string][]AdditionalFields `json:"update"`
		}{
			Update: input.Update,
		},
		Query: struct {
			NotifyUsers bool `json:"notify-users" api:"notifyUsers"`
			ReturnIssue bool `json:"return-issue" api:"returnIssue"`
		}{
			NotifyUsers: input.NotifyUsers,
			ReturnIssue: true,
		},
	}
	req := jiraClient.Client.R().SetResult(&UpdateIssueResp{}).SetBody(request.Body)

	err := addQueryOptions(req, request.Query)
	if err != nil {
		return nil, err
	}
	resp, err := req.Post(apiEndpoint)
	if err != nil {
		return nil, err
	}

	updatedIssue, ok := resp.Result().(*UpdateIssueResp)

	if !ok {
		return nil, errmsg.AddMessage(
			fmt.Errorf("failed to convert response to `Update Issue` Output"),
			fmt.Sprintf("failed to convert %v to `Update Issue` Output", resp.Result()),
		)
	}
	debug.Info("Updated Issue: ", updatedIssue)
	out, err := base.ConvertToStructpb(UpdateIssueOutput{Issue: updatedIssue.Issue})
	if err != nil {
		return nil, err
	}
	return out, nil
}
