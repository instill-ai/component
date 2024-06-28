package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v62/github"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

type Commit struct {
	SHA          	string         		`json:"sha,omitempty"`
	Message 		string          	`json:"message,omitempty"`
	Stats        	CommitStats         `json:"stats,omitempty"`
	Files 	  		[]CommitFile        `json:"files,omitempty"`
}
type CommitStats struct {
	Additions 		int 	`json:"additions,omitempty"`
	Deletions 		int 	`json:"deletions,omitempty"`
	Changes 		int 	`json:"changes,omitempty"`
}
type CommitFile struct {
	Filename 		string 	`json:"filename,omitempty"`
	Patch 			string 	`json:"patch,omitempty"`
	CommitStats
}
func (githubClient *GitHubClient) extractCommitFile(file *github.CommitFile) CommitFile {
	return CommitFile{
		Filename: file.GetFilename(),
		Patch: file.GetPatch(),
		CommitStats: CommitStats{
			Additions: file.GetAdditions(),
			Deletions: file.GetDeletions(),
			Changes: file.GetChanges(),
		},
	}
}
func (githubClient *GitHubClient) extractCommitInformation(originalCommit *github.RepositoryCommit) Commit {
	stats := originalCommit.GetStats()
	commitFiles := originalCommit.Files

	if stats == nil || commitFiles == nil{
		// fmt.Println("Getting stats for commit: ", originalCommit.GetSHA())
		commit, err := githubClient.getCommit(githubClient.owner, githubClient.repository ,originalCommit.GetSHA())
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
		SHA: originalCommit.GetSHA(),
		Message: originalCommit.GetCommit().GetMessage(),
		Stats: CommitStats{
			Additions: stats.GetAdditions(),
			Deletions: stats.GetDeletions(),
			Changes: stats.GetTotal(),
		},
		Files: files,
	}
}


func (githubClient *GitHubClient) getCommit(owner string, repository string, sha string) (*github.RepositoryCommit, error) {
	commit, _, err := githubClient.client.Repositories.GetCommit(context.Background(), owner, repository, sha, nil)
	return commit, err
}

type GetCommitResp struct {
	Commit Commit `json:"commit"`
}
func  (githubClient *GitHubClient) getCommitTask(props *structpb.Struct) (*structpb.Struct, error) {
	err := githubClient.setTargetRepo(props)
	if err != nil {
		return nil, err
	}
	sha := props.GetFields()["sha"].GetStringValue()
	commit, err := githubClient.getCommit(githubClient.owner, githubClient.repository, sha)
	if err != nil {
		return nil, err
	}
	var resp GetCommitResp
	resp.Commit = githubClient.extractCommitInformation(commit)
	fmt.Println("=====================================")
	fmt.Println("commit: ",resp.Commit)
	fmt.Println("=====================================")
	out, err := base.ConvertToStructpb(resp)
	if err != nil {
		return nil, err
	}

	return out, nil
}
