package github

import (
	"context"

	"github.com/google/go-github/v62/github"
)

type MockIssuesService struct {}

func (m *MockIssuesService)ListByRepo(ctx context.Context, owner, repo string, opt *github.IssueListByRepoOptions) ([]*github.Issue, *github.Response, error) {
	return nil, nil, nil
}
func (m *MockIssuesService)Get(ctx context.Context, owner, repo string, number int) (*github.Issue, *github.Response, error) {
	return nil, nil, nil
}
func (m *MockIssuesService)Create(ctx context.Context, owner, repo string, issue *github.IssueRequest) (*github.Issue, *github.Response, error) {
	return nil, nil, nil
}
func (m *MockIssuesService)Edit(ctx context.Context, owner, repo string, number int, issue *github.IssueRequest) (*github.Issue, *github.Response, error) {
	return nil, nil, nil
}
