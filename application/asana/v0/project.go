package asana

import (
	"context"
	"fmt"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/tools/logger"
	"google.golang.org/protobuf/types/known/structpb"
)

type ProjectTaskOutput struct {
	Project
}

type ProjectTaskResp struct {
	Data struct {
		GID                 string            `json:"gid"`
		Name                string            `json:"name"`
		Owner               User              `json:"owner"`
		Notes               string            `json:"notes"`
		HTMLNotes           string            `json:"html_notes"`
		DueOn               string            `json:"due_on"`
		StartOn             string            `json:"start_on"`
		Completed           bool              `json:"completed"`
		Color               string            `json:"color"`
		PrivacySetting      string            `json:"privacy_setting"`
		Archived            bool              `json:"archived"`
		CompletedBy         User              `json:"completed_by"`
		CurrentStatus       map[string]string `json:"current_status"`
		CustomFields        map[string]string `json:"custom_fields"`
		CustomFieldSettings map[string]string `json:"custom_field_settings"`
		ModifiedAt          string            `json:"modified_at"`
	} `json:"data"`
}

func projectResp2Output(resp *ProjectTaskResp) ProjectTaskOutput {
	out := ProjectTaskOutput{
		Project: Project{
			GID:                 resp.Data.GID,
			Name:                resp.Data.Name,
			Owner:               resp.Data.Owner,
			Notes:               resp.Data.Notes,
			HTMLNotes:           resp.Data.HTMLNotes,
			DueOn:               resp.Data.DueOn,
			StartOn:             resp.Data.StartOn,
			Completed:           resp.Data.Completed,
			Color:               resp.Data.Color,
			PrivacySetting:      resp.Data.PrivacySetting,
			Archived:            resp.Data.Archived,
			CompletedBy:         resp.Data.CompletedBy,
			CurrentStatus:       resp.Data.CurrentStatus,
			CustomFields:        resp.Data.CustomFields,
			CustomFieldSettings: resp.Data.CustomFieldSettings,
			ModifiedAt:          resp.Data.ModifiedAt,
		},
	}
	return out
}

type GetProjectInput struct {
	ID string `json:"project-gid"`
}

func (c *Client) GetProject(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("GetProject", logger.Develop).SessionEnd()
	var input GetProjectInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}

	debug.Info("input", input)

	apiEndpoint := fmt.Sprintf("/projects/%s", input.ID)
	req := c.Client.R().SetResult(&ProjectTaskResp{})

	wantOptFields := parseWantOptionFields(Project{})
	debug.Info("wantOptFields", wantOptFields)
	if err := addQueryOptions(req, map[string]interface{}{"opt_fields": wantOptFields}); err != nil {
		return nil, err
	}
	resp, err := req.Get(apiEndpoint)
	if err != nil {
		return nil, err
	}

	debug.Info("resp", resp)
	project := resp.Result().(*ProjectTaskResp)
	debug.Info("project", project)
	out := projectResp2Output(project)
	return base.ConvertToStructpb(out)
}

type UpdateProjectInput struct {
	ID             string `json:"project-gid"`
	Name           string `json:"name"`
	Notes          string `json:"notes"`
	DueOn          string `json:"due-on"`
	StartOn        string `json:"start-on"`
	Color          string `json:"color"`
	PrivacySetting string `json:"privacy-setting"`
	Archived       bool   `json:"archived"`
}

type UpdateProjectReq struct {
	Name           string `json:"name"`
	Notes          string `json:"notes"`
	DueOn          string `json:"due_on"`
	StartOn        string `json:"start_on"`
	Color          string `json:"color"`
	PrivacySetting string `json:"privacy_setting"`
	Archived       bool   `json:"archived"`
}

func (c *Client) UpdateProject(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("UpdateProject", logger.Develop).SessionEnd()
	var input UpdateProjectInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}

	debug.Info("input", input)

	apiEndpoint := fmt.Sprintf("/projects/%s", input.ID)
	req := c.Client.R().SetResult(&ProjectTaskResp{}).SetBody(
		map[string]interface{}{
			"body": &UpdateProjectReq{
				Name:           input.Name,
				Notes:          input.Notes,
				DueOn:          input.DueOn,
				StartOn:        input.StartOn,
				Color:          input.Color,
				PrivacySetting: input.PrivacySetting,
				Archived:       input.Archived,
			},
		})

	wantOptFields := parseWantOptionFields(Project{})
	debug.Info("wantOptFields", wantOptFields)
	if err := addQueryOptions(req, map[string]interface{}{"opt_fields": wantOptFields}); err != nil {
		return nil, err
	}

	resp, err := req.Put(apiEndpoint)

	if err != nil {
		return nil, err
	}
	debug.Info("resp", resp)
	project := resp.Result().(*ProjectTaskResp)
	debug.Info("project", project)
	out := projectResp2Output(project)
	return base.ConvertToStructpb(out)
}

type CreateProjectInput struct {
	ID             string `json:"project-gid"`
	Name           string `json:"name"`
	Notes          string `json:"notes"`
	DueOn          string `json:"due-on"`
	StartOn        string `json:"start-on"`
	Color          string `json:"color"`
	PrivacySetting string `json:"privacy-setting"`
}

type CreateProjectReq struct {
	Name           string `json:"name"`
	Notes          string `json:"notes"`
	DueOn          string `json:"due_on"`
	StartOn        string `json:"start_on"`
	Color          string `json:"color"`
	PrivacySetting string `json:"privacy_setting"`
}

func (c *Client) CreateProject(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("CreateProject", logger.Develop).SessionEnd()
	var input CreateProjectInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}

	debug.Info("input", input)

	apiEndpoint := "/projects"
	req := c.Client.R().SetResult(&ProjectTaskResp{}).SetBody(
		map[string]interface{}{
			"body": &CreateProjectReq{
				Name:           input.Name,
				Notes:          input.Notes,
				DueOn:          input.DueOn,
				StartOn:        input.StartOn,
				Color:          input.Color,
				PrivacySetting: input.PrivacySetting,
			},
		})
	wantOptFields := parseWantOptionFields(Project{})
	debug.Info("wantOptFields", wantOptFields)
	if err := addQueryOptions(req, map[string]interface{}{"opt_fields": wantOptFields}); err != nil {
		return nil, err
	}

	resp, err := req.Post(apiEndpoint)
	if err != nil {
		return nil, err
	}
	project := resp.Result().(*ProjectTaskResp)
	out := projectResp2Output(project)
	return base.ConvertToStructpb(out)
}

type DeleteProjectInput struct {
	ID string `json:"project-gid"`
}

func (c *Client) DeleteProject(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("DeleteProject", logger.Develop).SessionEnd()
	var input DeleteProjectInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}

	debug.Info("input", input)

	apiEndpoint := fmt.Sprintf("/projects/%s", input.ID)
	req := c.Client.R().SetResult(&ProjectTaskResp{})

	resp, err := req.Delete(apiEndpoint)
	if err != nil {
		return nil, err
	}
	project := resp.Result().(*ProjectTaskResp)
	out := projectResp2Output(project)
	return base.ConvertToStructpb(out)
}

type DuplicateProjectInput struct {
	ID                 string `json:"project-gid"`
	Name               string `json:"name"`
	Team               string `json:"team"`
	DueOn              string `json:"due-on,omitempty"`
	StartOn            string `json:"start-on,omitempty"`
	ShouldSkipWeekends bool   `json:"should-skip-weekends"`
}

type ScheduleDates struct {
	ShouldSkipWeekends bool   `json:"should_skip_weekends"`
	DueOn              string `json:"due_on,omitempty"`
	StartOn            string `json:"start_on,omitempty"`
}
type DuplicateProjectReq struct {
	Name          string        `json:"name"`
	Team          string        `json:"team"`
	Include       string        `json:"include"`
	ScheduleDates ScheduleDates `json:"schedule_dates"`
}

func (c *Client) DuplicateProject(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("DeleteProject", logger.Develop).SessionEnd()
	var input DuplicateProjectInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}

	debug.Info("input", input)

	apiEndpoint := fmt.Sprintf("/projects/%s/duplicate", input.ID)
	req := c.Client.R().SetResult(&ProjectTaskResp{}).SetBody(
		map[string]interface{}{
			"body": &DuplicateProjectReq{
				Name: input.Name,
				Team: input.Team,
				// include all fields, see https://developers.asana.com/reference/duplicateproject
				Include: "allocations,forms,members,notes,task_assignee,task_attachments,task_dates,task_dependencies,task_followers,task_notes,task_projects,task_subtasks,task_tags",
				ScheduleDates: ScheduleDates{
					ShouldSkipWeekends: input.ShouldSkipWeekends,
					DueOn:              input.DueOn,
					StartOn:            input.StartOn,
				},
			},
		},
	)

	wantOptFields := parseWantOptionFields(Project{})
	debug.Info("wantOptFields", wantOptFields)
	if err := addQueryOptions(req, map[string]interface{}{"opt_fields": wantOptFields}); err != nil {
		return nil, err
	}

	resp, err := req.Post(apiEndpoint)
	if err != nil {
		return nil, err
	}
	project := resp.Result().(*ProjectTaskResp)
	out := projectResp2Output(project)
	return base.ConvertToStructpb(out)
}
