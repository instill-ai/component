package github

import (
	"context"

	"github.com/google/go-github/v62/github"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

type IssuesService interface {
	ListByRepo(context.Context, string, string, *github.IssueListByRepoOptions) ([]*github.Issue, *github.Response, error)
	Get(context.Context, string, string, int) (*github.Issue, *github.Response, error)
	Create(context.Context, string, string, *github.IssueRequest) (*github.Issue, *github.Response, error)
	// Edit(context.Context, string, string, int, *github.IssueRequest) (*github.Issue, *github.Response, error)
}

type Issue struct {
	Number        int      `json:"number"`
	Title         string   `json:"title"`
	State         string   `json:"state"`
	Body          string   `json:"body"`
	Assignee      string   `json:"assignee"`
	Assignees     []string `json:"assignees"`
	Labels        []string `json:"labels"`
	IsPullRequest bool     `json:"is_pull_request"`
}

func (githubClient *Client) extractIssue(originalIssue *github.Issue) Issue {
	return Issue{
		Number:        originalIssue.GetNumber(),
		Title:         originalIssue.GetTitle(),
		State:         originalIssue.GetState(),
		Body:          originalIssue.GetBody(),
		Assignee:      originalIssue.GetAssignee().GetName(),
		Assignees:     extractAssignees(originalIssue.Assignees),
		Labels:        extractLabels(originalIssue.Labels),
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

func (githubClient *Client) getIssue(ctx context.Context, owner, repository string, issueNumber int) (*github.Issue, error) {
	issue, _, err := githubClient.Issues.Get(ctx, owner, repository, issueNumber)
	if err != nil {
		return nil, err
	}
	return issue, nil
}
func filterOutPullRequests(issues []Issue) []Issue {
	filteredIssues := make([]Issue, 0)
	for _, issue := range issues {
		if !issue.IsPullRequest {
			filteredIssues = append(filteredIssues, issue)
		}
	}
	return filteredIssues
}

type GetAllIssuesInput struct {
	RepoInfo
	State         string `json:"state"`
	Sort          string `json:"sort"`
	Direction     string `json:"direction"`
	Since         string `json:"since"`
	NoPullRequest bool   `json:"no_pull_request"`
}

type GetAllIssuesResp struct {
	Issues []Issue `json:"issues"`
}

func (githubClient *Client) getAllIssuesTask(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct GetAllIssuesInput
	err := base.ConvertFromStructpb(props, &inputStruct)
	if err != nil {
		return nil, err
	}
	owner, repository, err := parseTargetRepo(inputStruct)
	if err != nil {
		return nil, err
	}
	// from format like `2006-01-02T15:04:05Z07:00` to time.Time
	sinceTime, err := parseTime(inputStruct.Since)
	if err != nil {
		return nil, err
	}
	opts := &github.IssueListByRepoOptions{
		State:     inputStruct.State,
		Sort:      inputStruct.Sort,
		Direction: inputStruct.Direction,
		Since:     *sinceTime,
	}
	if opts.Mentioned == "none" {
		opts.Mentioned = ""
	}

	issues, _, err := githubClient.Issues.ListByRepo(ctx, owner, repository, opts)
	if err != nil {
		return nil, err
	}

	issueList := make([]Issue, len(issues))
	for idx, issue := range issues {
		issueList[idx] = githubClient.extractIssue(issue)
	}

	// filter out pull requests if no_pull_request is true
	if inputStruct.NoPullRequest {
		issueList = filterOutPullRequests(issueList)
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
	RepoInfo
	IssueNumber int `json:"issue_number"`
}

type GetIssueResp struct {
	Issue Issue `json:"issue"`
}

func (githubClient *Client) getIssueTask(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct GetIssueInput
	err := base.ConvertFromStructpb(props, &inputStruct)
	if err != nil {
		return nil, err
	}
	owner, repository, err := parseTargetRepo(inputStruct)
	if err != nil {
		return nil, err
	}

	issueNumber := inputStruct.IssueNumber
	issue, err := githubClient.getIssue(ctx, owner, repository, issueNumber)
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
	RepoInfo
	Title string `json:"title"`
	Body  string `json:"body"`
}

type CreateIssueResp struct {
	Issue Issue `json:"issue"`
}

func (githubClient *Client) createIssueTask(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct CreateIssueInput
	err := base.ConvertFromStructpb(props, &inputStruct)
	if err != nil {
		return nil, err
	}
	owner, repository, err := parseTargetRepo(inputStruct)
	if err != nil {
		return nil, err
	}
	issueRequest := &github.IssueRequest{
		Title: &inputStruct.Title,
		Body:  &inputStruct.Body,
	}
	issue, _, err := githubClient.Issues.Create(ctx, owner, repository, issueRequest)
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
