package github

import (
	"context"

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
//
// * This only works for public repositories.
func (githubClient *Client) getAllReviewCommentsTask(props *structpb.Struct) (*structpb.Struct, error) {
	err := githubClient.setTargetRepo(props)
	if err != nil {
		return nil, err
	}

	// from format like `2006-01-02T15:04:05Z07:00` to time.Time
	sinceTime, err := parseTime(props.GetFields()["since"].GetStringValue())
	if err != nil {
		return nil, err
	}
	opts := &github.PullRequestListCommentsOptions{
		Sort: props.GetFields()["sort"].GetStringValue(),
		Direction: props.GetFields()["direction"].GetStringValue(),
		Since: *sinceTime,
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
	var reviewCommentsResp GetAllReviewCommentsResp
	reviewCommentsResp.ReviewComments = reviewComments
	out, err := base.ConvertToStructpb(reviewCommentsResp)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type CreateReviewCommentInput struct {
	Owner      	string 		`json:"owner"`
	Repository 	string 		`json:"repository"`
	PrNumber    int    		`json:"pr_number"`
	Comment 	github.PullRequestComment 		`json:"comment"`
}

type CreateReviewCommentResp struct {
	ReviewComment ReviewComment `json:"comment"`
}

// CreateReviewComment creates a review comment for a given pull request.
//
// * This only works for public repositories.
func (githubClient *Client) createReviewCommentTask(props *structpb.Struct) (*structpb.Struct, error) {
	err := githubClient.setTargetRepo(props)
	if err != nil {
		return nil, err
	}

	var commentInput CreateReviewCommentInput
	err = base.ConvertFromStructpb(props, &commentInput)
	if err != nil {
		return nil, err
	}
	number := commentInput.PrNumber
	commentReqs := &commentInput.Comment
	commentReqs.Position = commentReqs.Line // Position is deprecated, use Line instead
	// commentReqs.OriginalLine = commentReqs.Line
	// commentReqs.OriginalPosition = commentReqs.Position
	// commentReqs.OriginalStartLine = commentReqs.StartLine

	comment, _, err := githubClient.client.PullRequests.CreateComment(context.Background(), githubClient.owner, githubClient.repository, number, commentReqs)
	if err != nil {
		return nil, err
	}

	reviewComment := extractReviewCommentInformation(comment)
	var reviewCommentResp CreateReviewCommentResp
	reviewCommentResp.ReviewComment = reviewComment
	out, err := base.ConvertToStructpb(reviewCommentResp)
	if err != nil {
		return nil, err
	}
	return out, nil
}
