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

type ListBoardsInput struct {
	ProjectKeyOrID string `json:"project-key-or-id,omitempty" api:"projectKeyOrID"`
	BoardType      string `json:"board-type,omitempty" api:"type"`
	Name           string `json:"name,omitempty" api:"name"`
	StartAt        int    `json:"start-at,omitempty" api:"startAt"`
	MaxResults     int    `json:"max-results,omitempty" api:"maxResults"`
}
type ListBoardsResp struct {
	Values     []Board `json:"values"`
	StartAt    int     `json:"startAt"`
	MaxResults int     `json:"maxResults"`
	Total      int     `json:"total"`
	IsLast     bool    `json:"isLast"`
}

type ListBoardsOutput struct {
	Boards     []Board `json:"boards"`
	StartAt    int     `json:"start-at"`
	MaxResults int     `json:"max-results"`
	Total      int     `json:"total"`
	IsLast     bool    `json:"is-last"`
}

func (jiraClient *Client) listBoardsTask(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var opt ListBoardsInput
	if err := base.ConvertFromStructpb(props, &opt); err != nil {
		return nil, err
	}

	boards, err := jiraClient.listBoards(ctx, &opt)
	if err != nil {
		return nil, err
	}
	var output ListBoardsOutput
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

func (jiraClient *Client) listBoards(_ context.Context, opt *ListBoardsInput) (*ListBoardsResp, error) {
	var debug DebugSession
	debug.SessionStart("listBoards", StaticVerboseLevel)
	defer debug.SessionEnd()

	apiEndpoint := "rest/agile/1.0/board"

	req := jiraClient.Client.R().SetResult(&ListBoardsResp{})
	debug.AddMapMessage("opt", *opt)
	err := addQueryOptions(req, *opt)
	if err != nil {
		return nil, err
	}
	resp, err := req.Get(apiEndpoint)

	if resp != nil && resp.StatusCode() == 404 {
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
	debug.AddMessage("GET", apiEndpoint)
	debug.AddMapMessage("QueryParam", resp.Request.QueryParam)
	debug.AddMessage("Status", resp.Status())
	boards := resp.Result().(*ListBoardsResp)
	return boards, err
}

func (jiraClient *Client) getBoard(_ context.Context, boardID int) (*Board, error) {
	var debug DebugSession
	debug.SessionStart("getBoard", DevelopVerboseLevel)
	defer debug.SessionEnd()

	apiEndpoint := fmt.Sprintf("rest/agile/1.0/board/%v", boardID)
	req := jiraClient.Client.R().SetResult(&Board{})
	resp, err := req.Get(apiEndpoint)

	if err != nil {
		return nil, fmt.Errorf(
			err.Error(), errmsg.Message(err),
		)
	}
	debug.AddMessage("GET", apiEndpoint)
	debug.AddMapMessage("QueryParam", resp.Request.QueryParam)
	debug.AddMessage("Status", resp.Status())
	board := resp.Result().(*Board)
	return board, err
}
