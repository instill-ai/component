package asana

import (
	"context"
	"encoding/json"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/tools/logger"
	"google.golang.org/protobuf/types/known/structpb"
)

type GoalTaskOutput struct {
	Goal
}

type GoalTaskResp struct {
	Data struct {
		GID       string    `json:"gid"`
		Name      string    `json:"name"`
		Owner     User      `json:"owner"`
		Notes     string    `json:"notes"`
		HTMLNotes string    `json:"html_notes"`
		DueOn     string    `json:"due_on"`
		StartOn   string    `json:"start_on"`
		Liked     bool      `json:"liked"`
		Likes     []RawLike `json:"likes"`
	} `json:"data"`
}

func goalResp2Output(resp *GoalTaskResp) GoalTaskOutput {
	out := GoalTaskOutput{
		Goal: Goal{
			GID:       resp.Data.GID,
			Name:      resp.Data.Name,
			Owner:     resp.Data.Owner,
			Notes:     resp.Data.Notes,
			HTMLNotes: resp.Data.HTMLNotes,
			DueOn:     resp.Data.DueOn,
			StartOn:   resp.Data.StartOn,
			Liked:     resp.Data.Liked,
		},
	}
	for _, like := range resp.Data.Likes {
		out.Likes = append(out.Likes, Like{
			LikeGID:  like.LikeGID,
			UserGID:  like.User.GID,
			UserName: like.User.Name,
		})
	}
	return out
}

type GetGoalInput struct {
	Action string `json:"action"`
	ID     string `json:"goal-gid"`
}

func (c *Client) GetGoal(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("GetGoal", logger.Develop).SessionEnd()
	var input GetGoalInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}

	debug.Info("input", input)
	apiEndpoint := "/goals/" + input.ID
	req := c.Client.R().SetResult(&GoalTaskResp{})

	wantOptFields := parseWantOptionFields(Goal{})
	debug.Info("wantOptFields", wantOptFields)
	if err := addQueryOptions(req, map[string]interface{}{"opt_fields": wantOptFields}); err != nil {
		return nil, err
	}
	resp, err := req.Get(apiEndpoint)
	if err != nil {
		return nil, err
	}
	goal := resp.Result().(*GoalTaskResp)
	debug.Info("goal", goal)
	out := goalResp2Output(goal)
	return base.ConvertToStructpb(out)
}

type UpdateGoalInput struct {
	Action  string `json:"action"`
	ID      string `json:"goal-gid"`
	Name    string `json:"name"`
	Notes   string `json:"notes"`
	DueOn   string `json:"due-on"`
	StartOn string `json:"start-on"`
	Liked   bool   `json:"liked"`
	Status  string `json:"status"`
}

type UpdateGoalReq struct {
	Name    string `json:"name,omitempty"`
	Notes   string `json:"notes,omitempty"`
	DueOn   string `json:"due_on,omitempty"`
	StartOn string `json:"start_on,omitempty"`
	Liked   bool   `json:"liked,omitempty"`
	Status  string `json:"status,omitempty"`
}

func (c *Client) UpdateGoal(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("UpdateGoal", logger.Develop).SessionEnd()
	var input UpdateGoalInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}

	debug.Info("input", input)

	apiEndpoint := "/goals/" + input.ID

	body, _ := json.Marshal(map[string]interface{}{
		"data": &UpdateGoalReq{
			Name:    input.Name,
			Notes:   input.Notes,
			DueOn:   input.DueOn,
			StartOn: input.StartOn,
			Liked:   input.Liked,
			Status:  input.Status,
		},
	})
	req := c.Client.R().SetResult(&GoalTaskResp{}).SetBody(string(body))
	// req := c.Client.R().SetResult(&GoalTaskResp{}).SetBody(map[string]interface{}{
	// 	"data": &UpdateGoalReq{
	// 		Name:    input.Name,
	// 		Notes:   input.Notes,
	// 		DueOn:   input.DueOn,
	// 		StartOn: input.StartOn,
	// 		Liked:   input.Liked,
	// 		Status:  input.Status,
	// 	},
	// })

	wantOptFields := parseWantOptionFields(Goal{})
	debug.Info("wantOptFields", wantOptFields)
	if err := addQueryOptions(req, map[string]interface{}{"opt_fields": wantOptFields}); err != nil {
		return nil, err
	}
	debug.Info("req", req)
	resp, err := req.Put(apiEndpoint)

	if err != nil {
		return nil, err
	}
	debug.Info("resp", resp)
	goal := resp.Result().(*GoalTaskResp)
	debug.Info("goal", goal)
	out := goalResp2Output(goal)
	return base.ConvertToStructpb(out)
}

type CreateGoalInput struct {
	Action  string `json:"action"`
	Name    string `json:"name"`
	Notes   string `json:"notes"`
	DueOn   string `json:"due-on"`
	StartOn string `json:"start-on"`
	Liked   bool   `json:"liked"`
}
type CreateGoalReq struct {
	Name    string `json:"name"`
	Notes   string `json:"notes"`
	DueOn   string `json:"due_on"`
	StartOn string `json:"start_on"`
	Liked   bool   `json:"liked"`
}

func (c *Client) CreateGoal(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("CreateGoal", logger.Develop).SessionEnd()
	var input CreateGoalInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}

	debug.Info("input", input)

	apiEndpoint := "/goals"
	req := c.Client.R().SetResult(&GoalTaskResp{}).SetBody(
		map[string]interface{}{
			"data": &CreateGoalReq{
				Name:    input.Name,
				Notes:   input.Notes,
				DueOn:   input.DueOn,
				StartOn: input.StartOn,
				Liked:   input.Liked,
			},
		})
	wantOptFields := parseWantOptionFields(Goal{})
	debug.Info("wantOptFields", wantOptFields)
	if err := addQueryOptions(req, map[string]interface{}{"opt_fields": wantOptFields}); err != nil {
		return nil, err
	}
	resp, err := req.Post(apiEndpoint)
	if err != nil {
		return nil, err
	}
	goal := resp.Result().(*GoalTaskResp)
	out := goalResp2Output(goal)
	return base.ConvertToStructpb(out)
}

type DeleteGoalInput struct {
	Action string `json:"action"`
	ID     string `json:"goal-gid"`
}

func (c *Client) DeleteGoal(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("DeleteGoal", logger.Develop).SessionEnd()
	var input DeleteGoalInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}

	debug.Info("input", input)

	apiEndpoint := "/goals/" + input.ID
	req := c.Client.R()

	_, err := req.Delete(apiEndpoint)
	if err != nil {
		return nil, err
	}

	out := GoalTaskOutput{}
	return base.ConvertToStructpb(out)
}
