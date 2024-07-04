package github

import (
	"context"

	"github.com/google/go-github/v62/github"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

type RepositoriesService interface {
	GetCommit(context.Context, string, string, string, *github.ListOptions) (*github.RepositoryCommit, *github.Response, error)
	CreateHook(context.Context, string, string, *github.Hook) (*github.Hook, *github.Response, error)
}

type Commit struct {
	SHA     string       `json:"sha"`
	Message string       `json:"message"`
	Stats   CommitStats  `json:"stats"`
	Files   []CommitFile `json:"files"`
}
type CommitStats struct {
	Additions int `json:"additions"`
	Deletions int `json:"deletions"`
	Changes   int `json:"changes"`
}
type CommitFile struct {
	Filename string `json:"filename"`
	Patch    string `json:"patch"`
	CommitStats
}

func (githubClient *Client) extractCommitFile(file *github.CommitFile) CommitFile {
	return CommitFile{
		Filename: file.GetFilename(),
		Patch:    file.GetPatch(),
		CommitStats: CommitStats{
			Additions: file.GetAdditions(),
			Deletions: file.GetDeletions(),
			Changes:   file.GetChanges(),
		},
	}
}
func (githubClient *Client) extractCommitInformation(originalCommit *github.RepositoryCommit) Commit {
	stats := originalCommit.GetStats()
	commitFiles := originalCommit.Files

	if stats == nil || commitFiles == nil {
		commit, err := githubClient.getCommit(githubClient.owner, githubClient.repository, originalCommit.GetSHA())
		if err == nil {
			// only update stats and files if there is no error
			// otherwise, we will maintain the original commit information
			stats = commit.GetStats()
			commitFiles = commit.Files
		}
	}
	files := make([]CommitFile, len(commitFiles))
	for idx, file := range commitFiles {
		files[idx] = githubClient.extractCommitFile(file)
	}

	return Commit{
		SHA:     originalCommit.GetSHA(),
		Message: originalCommit.GetCommit().GetMessage(),
		Stats: CommitStats{
			Additions: stats.GetAdditions(),
			Deletions: stats.GetDeletions(),
			Changes:   stats.GetTotal(),
		},
		Files: files,
	}
}

func (githubClient *Client) getCommit(owner string, repository string, sha string) (*github.RepositoryCommit, error) {
	commit, _, err := githubClient.client.Repositories.GetCommit(context.Background(), owner, repository, sha, nil)
	return commit, err
}

type GetCommitInput struct {
	Owner      string `json:"owner"`
	Repository string `json:"repository"`
	SHA        string `json:"sha"`
}

type GetCommitResp struct {
	Commit Commit `json:"commit"`
}

func (githubClient *Client) getCommitTask(props *structpb.Struct) (*structpb.Struct, error) {
	err := githubClient.setTargetRepo(props)
	if err != nil {
		return nil, err
	}
	var inputStruct GetCommitInput
	err = base.ConvertFromStructpb(props, &inputStruct)
	if err != nil {
		return nil, err
	}

	sha := inputStruct.SHA
	commit, err := githubClient.getCommit(githubClient.owner, githubClient.repository, sha)
	if err != nil {
		return nil, err
	}
	var resp GetCommitResp
	resp.Commit = githubClient.extractCommitInformation(commit)
	out, err := base.ConvertToStructpb(resp)
	if err != nil {
		return nil, err
	}

	return out, nil
}
