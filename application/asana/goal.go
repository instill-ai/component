package asana

import (
	"context"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/tools/logger"
	"google.golang.org/protobuf/types/known/structpb"
)

type GoalTaskOutput struct {
	Goal
}

type GetGoalInput struct {
	ID string `json:"goal-gid"`
}

func (c *Client) GetGoal(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("", logger.Develop).SessionEnd()

	var input GetGoalInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}

	debug.Info("GetGoal", input)

	apiEndpoint := "/goals/" + input.ID
	req := c.Client.R().SetResult(&Goal{})

	wantOptFields := parseWantOptionFields(Goal{})
	debug.Info("wantOptFields", wantOptFields)
	if err := addQueryOptions(req, map[string]interface{}{"opt_fields": wantOptFields}); err != nil {
		return nil, err
	}
	resp, err := req.Get(apiEndpoint)
	if err != nil {
		return nil, err
	}
	goal := resp.Result().(*Goal)
	return base.ConvertToStructpb(goal)
}

func (c *Client) CreateGoal(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	return nil, nil
}

func (c *Client) UpdateGoal(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	return nil, nil
}

func (c *Client) DeleteGoal(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	return nil, nil
}

func (c *Client) DuplicateGoal(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	return nil, nil
}
