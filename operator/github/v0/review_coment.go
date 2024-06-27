package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v62/github"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

type ReviewComment struct {
	github.PullRequestComment
}

func extractReviewCommentInformation(originalComment *github.PullRequestComment) ReviewComment {
	return ReviewComment{
		PullRequestComment: *originalComment,
	}
}

type GetAllReviewCommentsInput struct {
	Owner      	string 		`json:"owner"`
	Repository 	string 		`json:"repository"`
	PrNumber    int    		`json:"pr_number"`
	Sort 		string 		`json:"sort"`
	Direction 	string 		`json:"direction"`
	Since 		string 		`json:"since"`
}

type GetAllReviewCommentsResp struct {
	ReviewComments []ReviewComment `json:"comments"`
}

// GetAllReviewComments retrieves all review comments for a given pull request.
// Specifying a pull request number of 0 will return all comments on all pull requests for the repository.
func (githubClient *GitHubClient) getAllReviewComments(props *structpb.Struct) (*structpb.Struct, error) {
	err := githubClient.setTargetRepo(props)
	if err != nil {
		return nil, err
	}

	// from format like `2006-01-02T15:04:05Z07:00` to time.Time
	since:= props.GetFields()["since"].GetStringValue()
	sinceTime, err := time.Parse(time.RFC3339, since)
	if err != nil {
		return nil, err
	}
	opts := &github.PullRequestListCommentsOptions{
		Sort: props.GetFields()["sort"].GetStringValue(),
		Direction: props.GetFields()["direction"].GetStringValue(),
		Since: sinceTime,
	}
	number := int(props.GetFields()["pr_number"].GetNumberValue())
	comments, _, err := githubClient.client.PullRequests.ListComments(context.Background(), githubClient.owner, githubClient.repository, number, opts)
	if err != nil {
		return nil, err
	}

	reviewComments := make([]ReviewComment, len(comments))
	for i, comment := range comments {
		reviewComments[i] = extractReviewCommentInformation(comment)
	}
	fmt.Println("===========================")
	fmt.Println(reviewComments)
	fmt.Println("===========================")
	var reviewCommentsResp GetAllReviewCommentsResp
	reviewCommentsResp.ReviewComments = reviewComments
	out, err := base.ConvertToStructpb(reviewCommentsResp)
	if err != nil {
		return nil, err
	}
	return out, nil
}
