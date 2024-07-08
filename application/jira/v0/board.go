package jira

import (
	"context"
	_ "embed"

	jira "github.com/andygrunwald/go-jira"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

type Board struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Self      string `json:"self"`
	BoardType string `json:"type"`
}

type GetBoardsInput struct {
	ProjectKeyOrID string `json:"project_key_or_id,omitempty"`
	BoardType      string `json:"type,omitempty"`
	Name           string `json:"name,omitempty"`
	StartAt        int    `json:"start_at,omitempty"`
	MaxResults     int    `json:"max_results,omitempty"`
}

type GetBoardsOutput struct {
	Boards     []Board `json:"boards"`
	StartAt    int     `json:"start_at"`
	MaxResults int     `json:"max_results"`
	IsLast     bool    `json:"is_last"`
}

func (jiraClient *Client) extractBoard(board *jira.Board) *Board {
	return &Board{
		ID:        board.ID,
		Name:      board.Name,
		Self:      board.Self,
		BoardType: board.Type,
	}
}
func (jiraClient *Client) listBoardsTask(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var input GetBoardsInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}

	opt := jira.BoardListOptions{
		SearchOptions: jira.SearchOptions{
			StartAt:    input.StartAt,
			MaxResults: input.MaxResults,
		},
		BoardType:      input.BoardType,
		Name:           input.Name,
		ProjectKeyOrID: input.ProjectKeyOrID,
	}

	boards, _, err := jiraClient.Client.Board.GetAllBoardsWithContext(ctx, &opt)
	if err != nil {
		return nil, err
	}
	var output GetBoardsOutput
	for _, board := range boards.Values {
		output.Boards = append(output.Boards, *jiraClient.extractBoard(&board))
	}
	output.StartAt = boards.StartAt
	output.MaxResults = boards.MaxResults
	output.IsLast = boards.IsLast
	return base.ConvertToStructpb(output)
}
