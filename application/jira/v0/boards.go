package jira

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/x/errmsg"
	"google.golang.org/protobuf/types/known/structpb"
)

type Board struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Self      string `json:"self"`
	BoardType string `json:"type"`
}

type GetBoardsInput struct {
	ProjectKeyOrID string `json:"project_key_or_id,omitempty" struct:"projectKeyOrID"`
	BoardType      string `json:"board_type,omitempty" struct:"boardType"`
	Name           string `json:"name,omitempty" struct:"name"`
	StartAt        int    `json:"start_at,omitempty" struct:"startAt"`
	MaxResults     int    `json:"max_results,omitempty" struct:"maxResults"`
}
type GetBoardsResp struct {
	Values     []Board `json:"values"`
	StartAt    int     `json:"startAt"`
	MaxResults int     `json:"maxResults"`
	Total      int     `json:"total"`
	IsLast     bool    `json:"isLast"`
}

type GetBoardsOutput struct {
	Boards     []Board `json:"boards"`
	StartAt    int     `json:"start_at"`
	MaxResults int     `json:"max_results"`
	Total      int     `json:"total"`
	IsLast     bool    `json:"is_last"`
}

func (jiraClient *Client) listBoardsTask(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var opt GetBoardsInput
	if err := base.ConvertFromStructpb(props, &opt); err != nil {
		return nil, err
	}

	boards, err := jiraClient.listBoards(ctx, &opt)
	if err != nil {
		return nil, err
	}
	var output GetBoardsOutput
	output.Boards = append(output.Boards, boards.Values...)
	if output.Boards == nil {
		output.Boards = []Board{}
	}
	output.StartAt = boards.StartAt
	output.MaxResults = boards.MaxResults
	output.IsLast = boards.IsLast
	output.Total = boards.Total
	return base.ConvertToStructpb(output)
}

func (jiraClient *Client) listBoards(_ context.Context, opt *GetBoardsInput) (*GetBoardsResp, error) {
	apiEndpoint := "rest/agile/1.0/board"

	req := jiraClient.Client.R().SetResult(&GetBoardsResp{})
	err := addQueryOptions(req, *opt)
	if err != nil {
		return nil, err
	}
	resp, err := req.Get(apiEndpoint)
	if resp.StatusCode() == 404 {
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
	boards := resp.Result().(*GetBoardsResp)
	return boards, err
}
