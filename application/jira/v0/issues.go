package jira

import (
	"context"
	_ "embed"
	"fmt"
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
	Name    string                 `json:"Name"`
	Summary string                 `json:"summary"`
	Self    string                 `json:"self"`
	Done    bool                   `json:"done"`
	Fields  map[string]interface{} `json:"fields"`
}

type GetIssueInput struct {
	IssueKeyOrID  string `json:"issue-id-or-key,omitempty" struct:"issueIdOrKey"`
	UpdateHistory bool   `json:"update-history,omitempty" struct:"updateHistory"`
	FromBacklog   bool   `json:"from-backlog,omitempty" struct:"fromBacklog"`
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
	debug.SessionStart("getIssueTask", DevelopVerboseLevel)
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

type ListIssuesInput struct {
	BoardID    int `json:"board-id,omitempty" struct:"boardId"`
	MaxResults int `json:"max-results,omitempty" struct:"maxResults"`
	StartAt    int `json:"start-at,omitempty" struct:"startAt"`
	Range      struct {
		Range     string `json:"range,omitempty"`
		EpicKey   string `json:"epic-key,omitempty"`
		SprintKey string `json:"sprint-key,omitempty"`
	} `json:"range,omitempty"`
}

type ListIssuesResp struct {
	Issues     []Issue        `json:"issues"`
	Values     []SprintOrEpic `json:"values"`
	StartAt    int            `json:"start-at"`
	MaxResults int            `json:"max-results"`
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
	debug.SessionStart("listIssuesTask", DevelopVerboseLevel)
	defer debug.SessionEnd()

	debug.AddRawMessage(props)
	var opt ListIssuesInput
	if err := base.ConvertFromStructpb(props, &opt); err != nil {
		return nil, err
	}
	debug.AddMessage(fmt.Sprintf("GetIssueInput: %+v", opt))
	board, err := jiraClient.getBoard(ctx, opt.BoardID)
	if err != nil {
		return nil, err
	}
	debug.AddMapMessage("board", *board)
	boardKey := strings.Split(board.Name, " ")[0]
	apiEndpoint := fmt.Sprintf("rest/agile/1.0/board/%d", opt.BoardID)
	switch opt.Range.Range {
	case "All":
		apiEndpoint = apiEndpoint + "/issue"
	case "Epics only":
		apiEndpoint = apiEndpoint + "/epic"
	case "Sprints only":
		apiEndpoint = apiEndpoint + "/sprint"
	case "Issues of an epic":
		return jiraClient.nextGenIssuesSearch(ctx, nextGenSearchRequest{
			JQL:        fmt.Sprintf("project=\"%s\" AND parent=\"%s\"", boardKey, opt.Range.EpicKey),
			MaxResults: opt.MaxResults,
			StartAt:    opt.StartAt,
		},
		)
	case "Issues of a sprint":
		return jiraClient.nextGenIssuesSearch(ctx, nextGenSearchRequest{
			JQL:        fmt.Sprintf("project=\"%s\" AND sprint=\"%s\"", boardKey, opt.Range.EpicKey),
			MaxResults: opt.MaxResults,
			StartAt:    opt.StartAt,
		},
		)
	case "In backlog only":
		apiEndpoint = apiEndpoint + "/backlog"
	case "Issues without epic assigned":
		apiEndpoint = apiEndpoint + "/epic/none/issue"
	default:
		return nil, errmsg.AddMessage(
			fmt.Errorf("invalid range"),
			fmt.Sprintf("%s is an invalid range", opt.Range.Range),
		)
	}

	debug.AddMapMessage("opt", opt)
	req := jiraClient.Client.R().SetResult(&ListIssuesResp{})

	err = addQueryOptions(req, map[string]interface{}{
		"maxResults": opt.MaxResults,
		"startAt":    opt.StartAt,
	})
	if err != nil {
		return nil, err
	}
	debug.AddMessage("GET", apiEndpoint)
	resp, err := req.Get(apiEndpoint)
	debug.AddMessage("GET", apiEndpoint)
	debug.AddMapMessage("QueryParam", resp.Request.QueryParam)

	if err != nil {
		debug.AddMessage("Err", err.Error())
		debug.AddMessage("Err Message", errmsg.Message(err))
		if resp != nil && resp.StatusCode() == 404 {
			return nil, fmt.Errorf(
				err.Error(),
				errmsg.Message(err)+"Please check you have the correct permissions to access this resource.",
			)
		} else {
			return nil, err
		}
	}
	debug.AddMessage("Status", resp.Status())
	// debug.AddMessage("Result", fmt.Sprintf("%v", resp.Result()))

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
	for _, issue := range output.Issues {
		extractIssue(&issue)
	}
	return base.ConvertToStructpb(output)
}

// https://support.atlassian.com/jira-software-cloud/docs/jql-fields/
type nextGenSearchRequest struct {
	JQL        string `json:"jql,omitempty"`
	MaxResults int    `json:"maxResults,omitempty"`
	StartAt    int    `json:"startAt,omitempty"`
}

// https://developer.atlassian.com/cloud/jira/platform/rest/v2/api-group-issue-search/#api-rest-api-2-search-get
// https://developer.atlassian.com/cloud/jira/platform/rest/v2/api-group-issue-search/#api-rest-api-2-search-post
func (jiraClient *Client) nextGenIssuesSearch(_ context.Context, opt nextGenSearchRequest) (*structpb.Struct, error) {
	var debug DebugSession
	debug.SessionStart("nextGenIssuesSearch", DevelopVerboseLevel)
	defer debug.SessionEnd()

	debug.AddRawMessage(opt)
	var err error
	apiEndpoint := "/rest/api/2/search"

	req := jiraClient.Client.R().SetResult(&ListIssuesResp{})

	var resp *resty.Response
	if len(opt.JQL) < 50 {
		// filter seems not working
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
	for _, issue := range output.Issues {
		extractIssue(&issue)
	}
	return base.ConvertToStructpb(output)
}

var JQLReserveWords = []string{"a", "an", "abort", "access", "add", "after", "alias", "all", "alter", "and", "any", "are", "as", "asc", "at", "audit", "avg", "be", "before", "begin", "between", "boolean", "break", "but", "by", "byte", "catch", "cf", "char", "character", "check", "checkpoint", "collate", "collation", "column", "commit", "connect", "continue", "count", "create", "current", "date", "decimal", "declare", "decrement", "default", "defaults", "define", "delete", "delimiter", "desc", "difference", "distinct", "divide", "do", "double", "drop", "else", "empty", "encoding", "end", "equals", "escape", "exclusive", "exec", "execute", "exists", "explain", "false", "fetch", "file", "field", "first", "float", "for", "from", "function", "go", "goto", "grant", "greater", "group", "having", "identified", "if", "immediate", "in", "increment", "index", "initial", "inner", "inout", "input", "insert", "int", "integer", "intersect", "intersection", "into", "is", "isempty", "isnull", "it", "join", "last", "left", "less", "like", "limit", "lock", "long", "max", "min", "minus", "mode", "modify", "modulo", "more", "multiply", "next", "no", "noaudit", "not", "notin", "nowait", "null", "number", "object", "of", "on", "option", "or", "order", "outer", "output", "power", "previous", "prior", "privileges", "public", "raise", "raw", "remainder", "rename", "resource", "return", "returns", "revoke", "right", "row", "rowid", "rownum", "rows", "select", "session", "set", "share", "size", "sqrt", "start", "strict", "string", "subtract", "such", "sum", "synonym", "table", "that", "the", "their", "then", "there", "these", "they", "this", "to", "trans", "transaction", "trigger", "true", "uid", "union", "unique", "update", "user", "validate", "values", "view", "was", "when", "whenever", "where", "while", "will", "with"}
