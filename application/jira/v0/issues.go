package jira

import (
	"context"
	_ "embed"
	"fmt"
	"slices"
	"strings"

	"github.com/go-resty/resty/v2"
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
type SprintOrEpic struct {
	ID      int                    `json:"id"`
	Key     string                 `json:"key"`
	Name    string                 `json:"name"`
	Summary string                 `json:"summary"`
	Self    string                 `json:"self"`
	Done    bool                   `json:"done"`
	Fields  map[string]interface{} `json:"fields"`
}

type GetIssueInput struct {
	IssueKeyOrID  string `json:"issue-id-or-key,omitempty" api:"issueIdOrKey"`
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

func transformToIssue(val *SprintOrEpic) *Issue {
	fields := make(map[string]interface{})
	if val.Fields != nil {
		for key, value := range val.Fields {
			fields[key] = value
		}
	}

	return &Issue{
		ID:          fmt.Sprintf("%d", val.ID),
		Key:         val.Key,
		Description: val.Name,
		Summary:     val.Summary,
		Self:        val.Self,
		Fields:      fields,
	}
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
	req := jiraClient.Client.R().SetResult(&Issue{})

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
	debug.AddMapMessage("resp.Result()", resp.Result())
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
	Range     string `json:"range,omitempty"`
	EpicKey   string `json:"epic-key,omitempty"`
	SprintKey string `json:"sprint-key,omitempty"`
}
type ListIssuesInput struct {
	BoardID    int   `json:"board-id,omitempty" api:"boardId"`
	MaxResults int   `json:"max-results,omitempty" api:"maxResults"`
	StartAt    int   `json:"start-at,omitempty" api:"startAt"`
	Range      Range `json:"range,omitempty"`
}

type ListIssuesResp struct {
	Issues     []Issue        `json:"issues"`
	Values     []SprintOrEpic `json:"values"`
	StartAt    int            `json:"startAt"`
	MaxResults int            `json:"maxResults"`
	Total      int            `json:"total"`
}
type ListIssuesOutput struct {
	Issues     []Issue `json:"issues"`
	StartAt    int     `json:"start-at"`
	MaxResults int     `json:"max-results"`
	Total      int     `json:"total"`
}

func (jiraClient *Client) listIssuesTask(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug DebugSession
	debug.SessionStart("listIssuesTask", StaticVerboseLevel)
	defer debug.SessionEnd()

	debug.AddMapMessage("props", props)
	var (
		opt ListIssuesInput
		jql string
	)

	if err := base.ConvertFromStructpb(props, &opt); err != nil {
		return nil, err
	}
	debug.AddMapMessage("ListIssuesInput", opt)
	board, err := jiraClient.getBoard(ctx, opt.BoardID)
	if err != nil {
		return nil, err
	}
	debug.AddMapMessage("board", *board)
	boardKey := strings.Split(board.Name, " ")[0]
	apiEndpoint := fmt.Sprintf("rest/agile/1.0/board/%d", opt.BoardID)
	switch opt.Range.Range {
	case "All":
		// https://developer.atlassian.com/cloud/jira/software/rest/api-group-board/#api-rest-agile-1-0-board-boardid-issue-get
		apiEndpoint = apiEndpoint + "/issue"
	case "Epics only":
		// https://developer.atlassian.com/cloud/jira/software/rest/api-group-board/#api-rest-agile-1-0-board-boardid-epic-get
		apiEndpoint = apiEndpoint + "/epic"
	case "Sprints only":
		// https://developer.atlassian.com/cloud/jira/software/rest/api-group-board/#api-rest-agile-1-0-board-boardid-sprint-get
		apiEndpoint = apiEndpoint + "/sprint"
	case "Issues of an epic":
		// API not working: https://developer.atlassian.com/cloud/jira/software/rest/api-group-board/#api-rest-agile-1-0-board-boardid-epic-epicid-issue-get
		// use JQL instead
		jql = fmt.Sprintf("project=\"%s\" AND parent=\"%s\"", boardKey, opt.Range.EpicKey)
	case "Issues of a sprint":
		// API not working: https://developer.atlassian.com/cloud/jira/software/rest/api-group-board/#api-rest-agile-1-0-board-boardid-sprint-sprintid-issue-get
		// use JQL instead
		jql = fmt.Sprintf("project=\"%s\" AND sprint=\"%s\"", boardKey, opt.Range.SprintKey)
	case "In backlog only":
		// https://developer.atlassian.com/cloud/jira/software/rest/api-group-board/#api-rest-agile-1-0-board-boardid-backlog-get
		apiEndpoint = apiEndpoint + "/backlog"
	case "Issues without epic assigned":
		// https://developer.atlassian.com/cloud/jira/software/rest/api-group-board/#api-rest-agile-1-0-board-boardid-epic-none-issue-get
		apiEndpoint = apiEndpoint + "/epic/none/issue"
	default:
		return nil, errmsg.AddMessage(
			fmt.Errorf("invalid range"),
			fmt.Sprintf("%s is an invalid range", opt.Range.Range),
		)
	}

	var resp *resty.Response
	if slices.Contains([]string{"Issues of an epic", "Issues of a sprint"}, opt.Range.Range) {
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
	debug.AddMessage("Status", resp.Status())

	issues, ok := resp.Result().(*ListIssuesResp)
	if !ok {
		return nil, errmsg.AddMessage(
			fmt.Errorf("failed to convert response to `List Issue` Output"),
			fmt.Sprintf("failed to convert %v to `List Issue` Output", resp.Result()),
		)
	}

	if issues.Issues == nil && issues.Values == nil {
		issues.Issues = []Issue{}
	} else if issues.Issues == nil {
		issues.Issues = make([]Issue, len(issues.Values))
		for i, val := range issues.Values {
			issues.Issues[i] = *transformToIssue(&val)
		}
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
		} else if opt.Range.Range == "Sprints only" {
			output.Issues[idx].IssueType = "Sprint"
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
	var debug DebugSession
	debug.SessionStart("nextGenIssuesSearch", StaticVerboseLevel)
	defer debug.SessionEnd()

	debug.AddMessage("opt:")
	debug.AddRawMessage(opt)
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
	debug.AddMapMessage("Query", req.QueryParam)
	debug.AddMessage("Status", resp.Status())
	return resp, nil
}

var JQLReservedWords = []string{"a", "an", "abort", "access", "add", "after", "alias", "all", "alter", "and", "any", "are", "as", "asc", "at", "audit", "avg", "be", "before", "begin", "between", "boolean", "break", "but", "by", "byte", "catch", "cf", "char", "character", "check", "checkpoint", "collate", "collation", "column", "commit", "connect", "continue", "count", "create", "current", "date", "decimal", "declare", "decrement", "default", "defaults", "define", "delete", "delimiter", "desc", "difference", "distinct", "divide", "do", "double", "drop", "else", "empty", "encoding", "end", "equals", "escape", "exclusive", "exec", "execute", "exists", "explain", "false", "fetch", "file", "field", "first", "float", "for", "from", "function", "go", "goto", "grant", "greater", "group", "having", "identified", "if", "immediate", "in", "increment", "index", "initial", "inner", "inout", "input", "insert", "int", "integer", "intersect", "intersection", "into", "is", "isempty", "isnull", "it", "join", "last", "left", "less", "like", "limit", "lock", "long", "max", "min", "minus", "mode", "modify", "modulo", "more", "multiply", "next", "no", "noaudit", "not", "notin", "nowait", "null", "number", "object", "of", "on", "option", "or", "order", "outer", "output", "power", "previous", "prior", "privileges", "public", "raise", "raw", "remainder", "rename", "resource", "return", "returns", "revoke", "right", "row", "rowid", "rownum", "rows", "select", "session", "set", "share", "size", "sqrt", "start", "strict", "string", "subtract", "such", "sum", "synonym", "table", "that", "the", "their", "then", "there", "these", "they", "this", "to", "trans", "transaction", "trigger", "true", "uid", "union", "unique", "update", "user", "validate", "values", "view", "was", "when", "whenever", "where", "while", "will", "with"}
