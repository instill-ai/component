package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v62/github"
	"github.com/instill-ai/component/base"
	"github.com/instill-ai/x/errmsg"
	"google.golang.org/protobuf/types/known/structpb"
)
type IssuesService interface {
	ListByRepo(context.Context, string, string, *github.IssueListByRepoOptions) ([]*github.Issue, *github.Response, error)
	Get(context.Context, string, string, int) (*github.Issue, *github.Response, error)
	Create(context.Context, string, string, *github.IssueRequest) (*github.Issue, *github.Response, error)
	Edit(context.Context, string, string, int, *github.IssueRequest) (*github.Issue, *github.Response, error)
}

type Issue struct {
	Number        int    `json:"number"`
	Title         string `json:"title"`
	State         string `json:"state"`
	Body          string `json:"body"`
	Assignee      string `json:"assignee"`
	Assignees     []string `json:"assignees"`
	Labels        []string `json:"labels"`
	IsPullRequest bool   `json:"is_pull_request"`
}


func (githubClient *Client) extractIssue(originalIssue *github.Issue) Issue {
	return Issue{
		Number: originalIssue.GetNumber(),
		Title: originalIssue.GetTitle(),
		State: originalIssue.GetState(),
		Body: originalIssue.GetBody(),
		Assignee: originalIssue.GetAssignee().GetName(),
		Assignees: extractAssignees(originalIssue.Assignees),
		Labels: extractLabels(originalIssue.Labels),
		IsPullRequest: originalIssue.IsPullRequest(),
	}
}

func extractAssignees(assignees []*github.User) []string {
	assigneeList := make([]string, len(assignees))
	for idx, assignee := range assignees {
		assigneeList[idx] = assignee.GetName()
	}
	return assigneeList
}

func extractLabels(labels []*github.Label) []string {
	labelList := make([]string, len(labels))
	for idx, label := range labels {
		labelList[idx] = label.GetName()
	}
	return labelList
}

func (githubClient *Client) getIssue(owner, repository string, issueNumber int) (*github.Issue, error) {
	issue, _, err := githubClient.client.Issues.Get(context.Background(), owner, repository, issueNumber)
	if err != nil {
		return nil, err
	}
	return issue, nil
}

type GetAllIssuesInput struct {
	Owner		 	string 			`json:"owner"`
	Repository		string 			`json:"repository"`
	State		 	string 			`json:"state"`
	Sort		 	string 			`json:"sort"`
	Direction		string 			`json:"direction"`
	Since 		 	string 			`json:"since"`
	Mentioned 		string 			`json:"mentioned"` // A user that's mentioned in the issue.
}

type GetAllIssuesResp struct {
	Issues []Issue `json:"issues"`
}

func (githubClient *Client) getAllIssuesTask(props *structpb.Struct) (*structpb.Struct, error) {
	err := githubClient.setTargetRepo(props)
	if err != nil {
		return nil, err
	}
	since:= props.GetFields()["since"].GetStringValue()
	sinceTime, err := time.Parse(time.RFC3339, since)
	if err != nil {
		return nil, errmsg.AddMessage(
			fmt.Errorf("invalid time format"),
			fmt.Sprintf("Cannot parse time: %s, Please provide format like %s(see RFC3339)", since, time.RFC3339),
		)
	}
	opts := &github.IssueListByRepoOptions{
		State: props.GetFields()["state"].GetStringValue(),
		Sort: props.GetFields()["sort"].GetStringValue(),
		Direction: props.GetFields()["direction"].GetStringValue(),
		Since: sinceTime,
		Mentioned: props.GetFields()["mentioned"].GetStringValue(),
	}
	if opts.Mentioned == "none"{
		opts.Mentioned = ""
	}

	issues, _, err := githubClient.client.Issues.ListByRepo(context.Background(), githubClient.owner, githubClient.repository, opts)
	if err != nil {
		return nil, err
	}

	issueList := make([]Issue, len(issues))
	for idx, issue := range issues {
		issueList[idx] = githubClient.extractIssue(issue)
	}
	var resp GetAllIssuesResp
	resp.Issues = issueList
	out, err := base.ConvertToStructpb(resp)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type GetIssueInput struct {
	Owner		 	string 			`json:"owner"`
	Repository		string 			`json:"repository"`
	IssueNumber 	int 			`json:"issue_number"`
}

type GetIssueResp struct {
	Issue Issue `json:"issue"`
}

func (githubClient *Client) getIssueTask(props *structpb.Struct) (*structpb.Struct, error) {
	err := githubClient.setTargetRepo(props)
	if err != nil {
		return nil, err
	}

	issueNumber := int(props.GetFields()["issue_number"].GetNumberValue())
	issue, err := githubClient.getIssue(githubClient.owner, githubClient.repository, issueNumber)
	if err != nil {
		return nil, err
	}

	issueResp := githubClient.extractIssue(issue)
	var resp GetIssueResp
	resp.Issue = issueResp
	out, err := base.ConvertToStructpb(resp)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type CreateIssueInput struct {
	Owner      	string 		`json:"owner"`
	Repository 	string 		`json:"repository"`
	Title      	string 		`json:"title"`
	Body       	string 		`json:"body"`
}

type CreateIssueResp struct {
	Issue Issue `json:"issue"`
}

func (githubClient *Client) createIssueTask(props *structpb.Struct) (*structpb.Struct, error) {
	err := githubClient.setTargetRepo(props)
	if err != nil {
		return nil, err
	}

	title := props.GetFields()["title"].GetStringValue()
	body := props.GetFields()["body"].GetStringValue()
	issueRequest := &github.IssueRequest{
		Title: &title,
		Body: &body,
	}
	issue, _, err := githubClient.client.Issues.Create(context.Background(), githubClient.owner, githubClient.repository, issueRequest)
	if err != nil {
		return nil, err
	}

	issueResp := githubClient.extractIssue(issue)
	var resp CreateIssueResp
	resp.Issue = issueResp
	out, err := base.ConvertToStructpb(resp)
	if err != nil {
		return nil, err
	}
	return out, nil
}
