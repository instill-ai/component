package asana

import (
	"context"
	"fmt"
	"strings"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/tools/logger"
	"google.golang.org/protobuf/types/known/structpb"
)

type TaskTaskOutput struct {
	Task
}

type TaskTaskResp struct {
	Data struct {
		GID       string          `json:"gid"`
		Name      string          `json:"name"`
		Notes     string          `json:"notes"`
		HTMLNotes string          `json:"html_notes"`
		Projects  []SimpleProject `json:"projects"`
		DueOn     string          `json:"due_on"`
		StartOn   string          `json:"start_on"`
		Liked     bool            `json:"liked"`
		Likes     []RawLike       `json:"likes"`
	} `json:"data"`
}

func taskResp2Output(resp *TaskTaskResp) TaskTaskOutput {
	out := TaskTaskOutput{
		Task: Task{
			GID:       resp.Data.GID,
			Name:      resp.Data.Name,
			Notes:     resp.Data.Notes,
			HTMLNotes: resp.Data.HTMLNotes,
			Projects:  resp.Data.Projects,
			DueOn:     resp.Data.DueOn,
			StartOn:   resp.Data.StartOn,
			Liked:     resp.Data.Liked,
			Likes:     []Like{},
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

type GetTaskInput struct {
	ID string `json:"task-gid"`
}

func (c *Client) GetTask(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("GetTask", logger.Develop).SessionEnd()
	var input GetTaskInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}

	debug.Info("input", input)

	apiEndpoint := fmt.Sprintf("/tasks/%s", input.ID)
	req := c.Client.R().SetResult(&TaskTaskResp{})

	wantOptFields := parseWantOptionFields(Task{})
	debug.Info("wantOptFields", wantOptFields)
	if err := addQueryOptions(req, map[string]interface{}{"opt_fields": wantOptFields}); err != nil {
		return nil, err
	}
	resp, err := req.Get(apiEndpoint)
	if err != nil {
		return nil, err
	}

	debug.Info("resp", resp)
	task := resp.Result().(*TaskTaskResp)
	debug.Info("task", task)
	out := taskResp2Output(task)
	return base.ConvertToStructpb(out)
}

type UpdateTaskInput struct {
	ID              string `json:"task-gid"`
	Name            string `json:"name"`
	ResourceSubtype string `json:"resource-subtype"`
	ApprovalStatus  string `json:"approval-status"`
	Completed       bool   `json:"completed"`
	Liked           bool   `json:"liked"`
	Notes           string `json:"notes"`
	Assignee        string `json:"assignee"`
	Parent          string `json:"parent"`
}

type UpdateTaskReq struct {
	Name            string `json:"name"`
	ResourceSubtype string `json:"resource_subtype"`
	ApprovalStatus  string `json:"approval_status"`
	Completed       bool   `json:"completed"`
	Liked           bool   `json:"liked"`
	Notes           string `json:"notes"`
	Assignee        string `json:"assignee"`
	Parent          string `json:"parent"`
}

func (c *Client) UpdateTask(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("UpdateTask", logger.Develop).SessionEnd()
	var input UpdateTaskInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}

	debug.Info("input", input)

	apiEndpoint := fmt.Sprintf("/tasks/%s", input.ID)
	req := c.Client.R().SetResult(&TaskTaskResp{}).SetBody(
		map[string]interface{}{
			"body": &UpdateTaskReq{
				Name:            input.Name,
				ResourceSubtype: input.ResourceSubtype,
				ApprovalStatus:  input.ApprovalStatus,
				Completed:       input.Completed,
				Liked:           input.Liked,
				Notes:           input.Notes,
				Assignee:        input.Assignee,
				Parent:          input.Parent,
			},
		})

	wantOptFields := parseWantOptionFields(Task{})
	debug.Info("wantOptFields", wantOptFields)
	if err := addQueryOptions(req, map[string]interface{}{"opt_fields": wantOptFields}); err != nil {
		return nil, err
	}

	resp, err := req.Put(apiEndpoint)

	if err != nil {
		return nil, err
	}
	debug.Info("resp", resp)
	task := resp.Result().(*TaskTaskResp)
	debug.Info("task", task)
	out := taskResp2Output(task)
	return base.ConvertToStructpb(out)
}

type CreateTaskInput struct {
	ID              string `json:"task-gid"`
	Name            string `json:"name"`
	Notes           string `json:"notes"`
	ResourceSubtype string `json:"resource-subtype"`
	ApprovalStatus  string `json:"approval-status"`
	Completed       bool   `json:"completed"`
	Liked           bool   `json:"liked"`
	Assignee        string `json:"assignee"`
	Parent          string `json:"parent"`
	StartAt         string `json:"start-at"`
	DueAt           string `json:"due-at"`
}

type CreateTaskReq struct {
	Name            string `json:"name"`
	Notes           string `json:"notes"`
	ResourceSubtype string `json:"resource_subtype"`
	ApprovalStatus  string `json:"approval_status"`
	Completed       bool   `json:"completed"`
	Liked           bool   `json:"liked"`
	Assignee        string `json:"assignee"`
	Parent          string `json:"parent"`
	StartAt         string `json:"start_at"`
	DueAt           string `json:"due_at"`
}

func (c *Client) CreateTask(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("CreateTask", logger.Develop).SessionEnd()
	var input CreateTaskInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}

	debug.Info("input", input)

	apiEndpoint := "/tasks"
	req := c.Client.R().SetResult(&TaskTaskResp{}).SetBody(
		map[string]interface{}{
			"body": &CreateTaskReq{
				Name:            input.Name,
				Notes:           input.Notes,
				ResourceSubtype: input.ResourceSubtype,
				ApprovalStatus:  input.ApprovalStatus,
				Completed:       input.Completed,
				Liked:           input.Liked,
				Assignee:        input.Assignee,
				Parent:          input.Parent,
				StartAt:         input.StartAt,
				DueAt:           input.DueAt,
			},
		})
	wantOptFields := parseWantOptionFields(Task{})
	debug.Info("wantOptFields", wantOptFields)
	if err := addQueryOptions(req, map[string]interface{}{"opt_fields": wantOptFields}); err != nil {
		return nil, err
	}

	resp, err := req.Post(apiEndpoint)
	if err != nil {
		return nil, err
	}
	task := resp.Result().(*TaskTaskResp)
	out := taskResp2Output(task)
	return base.ConvertToStructpb(out)
}

type DeleteTaskInput struct {
	ID string `json:"task-gid"`
}

func (c *Client) DeleteTask(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("DeleteTask", logger.Develop).SessionEnd()
	var input DeleteTaskInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}

	debug.Info("input", input)

	apiEndpoint := fmt.Sprintf("/tasks/%s", input.ID)
	req := c.Client.R().SetResult(&TaskTaskResp{})

	resp, err := req.Delete(apiEndpoint)
	if err != nil {
		return nil, err
	}
	task := resp.Result().(*TaskTaskResp)
	out := taskResp2Output(task)
	return base.ConvertToStructpb(out)

}

type DuplicateTaskInput struct {
	ID   string `json:"task-gid"`
	Name string `json:"name"`
}

type DuplicateTaskReq struct {
	Name    string `json:"name"`
	Include string `json:"include"`
}

func (c *Client) DuplicateTask(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("DeleteTask", logger.Develop).SessionEnd()
	var input DuplicateTaskInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}

	debug.Info("input", input)

	apiEndpoint := fmt.Sprintf("/tasks/%s/duplicate", input.ID)
	req := c.Client.R().SetResult(&TaskTaskResp{}).SetBody(
		map[string]interface{}{
			"body": &DuplicateTaskReq{
				Name: input.Name,
				// include all fields, see https://developers.asana.com/reference/duplicatetask
				Include: "assignee,attachments,dates,dependencies,followers,notes,parent,projects,subtasks,tags",
			},
		},
	)

	wantOptFields := parseWantOptionFields(Task{})
	debug.Info("wantOptFields", wantOptFields)
	if err := addQueryOptions(req, map[string]interface{}{"opt_fields": wantOptFields}); err != nil {
		return nil, err
	}

	resp, err := req.Post(apiEndpoint)
	if err != nil {
		return nil, err
	}
	task := resp.Result().(*TaskTaskResp)
	out := taskResp2Output(task)
	return base.ConvertToStructpb(out)
}

type TaskSetParentInput struct {
	ID     string `json:"task-gid"`
	Parent string `json:"parent"`
}
type TaskSetParentReq struct {
	Parent string `json:"parent"`
}

func (c *Client) TaskSetParent(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("TaskSetParent", logger.Develop).SessionEnd()
	var input TaskSetParentInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}

	debug.Info("input", input)

	apiEndpoint := fmt.Sprintf("/tasks/%s/setParent", input.ID)
	req := c.Client.R().SetResult(&TaskTaskResp{}).SetBody(
		map[string]interface{}{
			"body": &TaskSetParentReq{
				Parent: input.Parent,
			},
		},
	)

	wantOptFields := parseWantOptionFields(Task{})
	debug.Info("wantOptFields", wantOptFields)
	if err := addQueryOptions(req, map[string]interface{}{"opt_fields": wantOptFields}); err != nil {
		return nil, err
	}

	resp, err := req.Post(apiEndpoint)
	if err != nil {
		return nil, err
	}
	task := resp.Result().(*TaskTaskResp)
	out := taskResp2Output(task)
	return base.ConvertToStructpb(out)
}

type TaskEditTagInput struct {
	ID         string `json:"task-gid"`
	TagID      string `json:"tag-gid"`
	EditOption string `json:"edit-option"`
}
type TaskEditTagReq struct {
	Tag string `json:"tag"`
}

func (c *Client) TaskEditTag(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("TaskEditTag", logger.Develop).SessionEnd()
	var input TaskEditTagInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}
	debug.Info("input", input)

	apiEndpoint := fmt.Sprintf("/tasks/%s/", input.ID)
	if input.EditOption == "add" {
		apiEndpoint += "addTag"
	} else if input.EditOption == "remove" {
		apiEndpoint += "removeTag"
	}

	req := c.Client.R().SetResult(&TaskTaskResp{}).SetBody(
		map[string]interface{}{
			"body": &TaskEditTagReq{
				Tag: input.TagID,
			},
		},
	)
	_, err := req.Post(apiEndpoint)
	if err != nil {
		return nil, err
	}
	return c.GetTask(ctx, props)
}

type TaskEditFollowerInput struct {
	ID         string `json:"task-gid"`
	Followers  string `json:"followers"`
	EditOption string `json:"edit-option"`
}
type TaskEditFollowerReq struct {
	Followers []string `json:"followers"`
}

func (c *Client) TaskEditFollower(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("TaskEditFollower", logger.Develop).SessionEnd()
	var input TaskEditFollowerInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}
	debug.Info("input", input)

	apiEndpoint := fmt.Sprintf("/tasks/%s/", input.ID)
	if input.EditOption == "add" {
		apiEndpoint += "addFollowers"
	} else if input.EditOption == "remove" {
		apiEndpoint += "removeFollowers"
	}
	followers := strings.Split(input.Followers, ",")
	req := c.Client.R().SetResult(&TaskTaskResp{}).SetBody(
		map[string]interface{}{
			"body": &TaskEditFollowerReq{
				Followers: followers,
			},
		},
	)
	wantOptFields := parseWantOptionFields(Task{})
	debug.Info("wantOptFields", wantOptFields)
	if err := addQueryOptions(req, map[string]interface{}{"opt_fields": wantOptFields}); err != nil {
		return nil, err
	}

	resp, err := req.Post(apiEndpoint)
	if err != nil {
		return nil, err
	}
	task := resp.Result().(*TaskTaskResp)
	out := taskResp2Output(task)
	return base.ConvertToStructpb(out)
}

type TaskEditProjectInput struct {
	ID         string `json:"task-gid"`
	ProjectID  string `json:"project-gid"`
	EditOption string `json:"edit-option"`
}
type TaskEditProjectReq struct {
	ProjectID string `json:"project"`
}

func (c *Client) TaskEditProject(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("TaskEditProject", logger.Develop).SessionEnd()
	var input TaskEditProjectInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}
	debug.Info("input", input)

	apiEndpoint := fmt.Sprintf("/tasks/%s/", input.ID)
	if input.EditOption == "add" {
		apiEndpoint += "addProject"
	} else if input.EditOption == "remove" {
		apiEndpoint += "removeProject"
	}

	req := c.Client.R().SetResult(&TaskTaskResp{}).SetBody(
		map[string]interface{}{
			"body": &TaskEditProjectReq{
				ProjectID: input.ProjectID,
			},
		},
	)
	_, err := req.Post(apiEndpoint)
	if err != nil {
		return nil, err
	}
	return c.GetTask(ctx, props)
}
