package github

import (
	"context"
	"strings"

	"github.com/google/go-github/v62/github"
	"github.com/instill-ai/component/base"
	"github.com/instill-ai/x/errmsg"
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
	RepoInfo
	PrNumber  int    `json:"pr_number"`
	Sort      string `json:"sort"`
	Direction string `json:"direction"`
	Since     string `json:"since"`
}

type GetAllReviewCommentsResp struct {
	ReviewComments []ReviewComment `json:"comments"`
}

// GetAllReviewComments retrieves all review comments for a given pull request.
// Specifying a pull request number of 0 will return all comments on all pull requests for the repository.
//
// * This only works for public repositories.
func (githubClient *Client) getAllReviewCommentsTask(props *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct GetAllReviewCommentsInput
	err := base.ConvertFromStructpb(props, &inputStruct)
	if err != nil {
		return nil, err
	}

	err = githubClient.setTargetRepo(inputStruct)
	if err != nil {
		return nil, err
	}
	// from format like `2006-01-02T15:04:05Z07:00` to time.Time
	sinceTime, err := parseTime(inputStruct.Since)
	if err != nil {
		return nil, err
	}
	opts := &github.PullRequestListCommentsOptions{
		Sort:      inputStruct.Sort,
		Direction: inputStruct.Direction,
		Since:     *sinceTime,
	}
	number := inputStruct.PrNumber
	comments, _, err := githubClient.client.PullRequests.ListComments(context.Background(), githubClient.owner, githubClient.repository, number, opts)
	if err != nil {
		errMessage := strings.Split(err.Error(), ": ")
		if len(errMessage) < 2 {
			return nil, err
		}
		errType := strings.TrimSpace(errMessage[1])
		if strings.Contains(errType, "404 Not Found") {
			return nil, errmsg.AddMessage(
				err,
				"Pull request not found. Ensure the pull request number is correct, the repository is public, and fill in the correct GitHub token.",
			)
		}
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
	RepoInfo
	PrNumber int                       `json:"pr_number"`
	Comment  github.PullRequestComment `json:"comment"`
}

type CreateReviewCommentResp struct {
	ReviewComment ReviewComment `json:"comment"`
}

// CreateReviewComment creates a review comment for a given pull request.
//
// * This only works for public repositories.
func (githubClient *Client) createReviewCommentTask(props *structpb.Struct) (*structpb.Struct, error) {
	var commentInput CreateReviewCommentInput
	err := base.ConvertFromStructpb(props, &commentInput)
	if err != nil {
		return nil, err
	}

	err = githubClient.setTargetRepo(commentInput)
	if err != nil {
		return nil, err
	}
	number := commentInput.PrNumber
	commentReqs := &commentInput.Comment
	commentReqs.Position = commentReqs.Line // Position is deprecated, use Line instead

	if *commentReqs.Line == *commentReqs.StartLine {
		commentReqs.StartLine = nil // If it's a one line comment, don't send start_line
	}
	comment, _, err := githubClient.client.PullRequests.CreateComment(context.Background(), githubClient.owner, githubClient.repository, number, commentReqs)
	if err != nil {
		errMessage := strings.Split(err.Error(), ": ")
		if len(errMessage) < 2 {
			return nil, err
		}
		errType := strings.TrimSpace(errMessage[1])
		if strings.Contains(errType, "404 Not Found") {
			return nil, errmsg.AddMessage(
				err,
				"Pull request not found. Ensure the pull request number is correct, the repository is public, and fill in the correct GitHub token.",
			)
		}
		if strings.Contains(errType, "422 Validation Failed") {
			return nil, errmsg.AddMessage(
				err,
				"Invalid comment. Ensure the comment is not empty and the line numbers and sides are correct.",
			)
		}
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
