package asana

import (
	"testing"

	"github.com/instill-ai/component/application/asana/v0"
)

func TestGetProject(t *testing.T) {
	testcases := []taskCase[asana.GetProjectInput, asana.ProjectTaskOutput]{
		{
			_type: "ok",
			name:  "Get project",
			input: asana.GetProjectInput{
				Action: "get",
				ID:     "1234",
			},
			wantResp: asana.ProjectTaskOutput{
				Project: asana.Project{
					GID:            "1234",
					Name:           "Test Project",
					Owner:          asana.User{GID: "123", Name: "Admin User"},
					Notes:          "Test Notes",
					HTMLNotes:      "Test HTML Notes",
					DueOn:          "2021-01-01",
					StartOn:        "2021-01-01",
					Archived:       true,
					Color:          "red",
					Completed:      true,
					ModifiedAt:     "2021-01-01",
					PrivacySetting: "public_to_workspace",
					CompletedBy:    asana.User{GID: "123", Name: "Admin User"},
					CurrentStatus: map[string]string{
						"status": "completed",
					},
					CustomFields: map[string]string{
						"field": "value",
					},
					CustomFieldSettings: map[string]string{
						"field": "value",
					},
				},
			},
		},
		{
			_type: "nok",
			name:  "Get project - 404 Not Found",
			input: asana.GetProjectInput{
				Action: "get",
				ID:     "12345",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, asana.TaskAsanaProject, t)
}
func TestUpdateProject(t *testing.T) {
	testcases := []taskCase[asana.UpdateProjectInput, asana.ProjectTaskOutput]{
		{
			_type: "ok",
			name:  "Update project",
			input: asana.UpdateProjectInput{
				Action:   "update",
				ID:       "123",
				Notes:    "Modified Notes",
				DueOn:    "2021-01-02",
				Archived: true,
			},
			wantResp: asana.ProjectTaskOutput{
				Project: asana.Project{
					GID:            "123",
					Name:           "Test Project",
					Owner:          asana.User{GID: "123", Name: "Admin User"},
					Notes:          "Modified Notes",
					HTMLNotes:      "Test HTML Notes",
					DueOn:          "2021-01-02",
					StartOn:        "2021-01-01",
					Archived:       true,
					Color:          "red",
					Completed:      true,
					ModifiedAt:     "2021-01-01",
					PrivacySetting: "public_to_workspace",
					CompletedBy:    asana.User{GID: "123", Name: "Admin User"},
					CurrentStatus: map[string]string{
						"status": "completed",
					},
					CustomFields: map[string]string{
						"field": "value",
					},
					CustomFieldSettings: map[string]string{
						"field": "value",
					},
				},
			},
		},
		{
			_type: "nok",
			name:  "Update project - 404 Not Found",
			input: asana.UpdateProjectInput{
				Action: "update",
				ID:     "12345",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, asana.TaskAsanaProject, t)
}
func TestCreateProject(t *testing.T) {
	testcases := []taskCase[asana.CreateProjectInput, asana.ProjectTaskOutput]{
		{
			_type: "ok",
			name:  "Create project",
			input: asana.CreateProjectInput{
				Action:         "create",
				Name:           "Test Project",
				Notes:          "Test Notes",
				DueOn:          "2021-01-02",
				StartOn:        "2021-01-01",
				Color:          "red",
				PrivacySetting: "public to workspace",
			},
			wantResp: asana.ProjectTaskOutput{
				Project: asana.Project{
					GID:            "123456789",
					Name:           "Test Project",
					Owner:          asana.User{GID: "123", Name: "Admin User"},
					Notes:          "Test Notes",
					HTMLNotes:      "Test HTML Notes",
					DueOn:          "2021-01-02",
					StartOn:        "2021-01-01",
					Color:          "red",
					PrivacySetting: "public_to_workspace",
					Completed:      false,
					Archived:       false,
					CompletedBy:    asana.User{GID: "123", Name: "Admin User"},
					CurrentStatus: map[string]string{
						"status": "on_track",
					},
					CustomFields: map[string]string{
						"field": "value",
					},
					CustomFieldSettings: map[string]string{
						"field": "value",
					},
				},
			},
		},
		{
			_type: "nok",
			name:  "Create project - 400 Bad Request",
			input: asana.CreateProjectInput{
				Action: "create",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, asana.TaskAsanaProject, t)
}

func TestDeleteProject(t *testing.T) {
	testcases := []taskCase[asana.DeleteProjectInput, asana.ProjectTaskOutput]{
		{
			_type: "ok",
			name:  "Delete project",
			input: asana.DeleteProjectInput{
				Action: "delete",
				ID:     "1234567890",
			},
			wantResp: asana.ProjectTaskOutput{
				Project: asana.Project{},
			},
		},
		{
			_type: "nok",
			name:  "Delete project - 404 Not Found",
			input: asana.DeleteProjectInput{
				Action: "delete",
				ID:     "12345",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, asana.TaskAsanaProject, t)
}

func TestDuplicateProject(t *testing.T) {
	testcases := []taskCase[asana.DuplicateProjectInput, asana.ProjectTaskOutput]{
		{
			_type: "ok",
			name:  "Duplicate project",
			input: asana.DuplicateProjectInput{
				Action:             "duplicate",
				ID:                 "1234",
				Name:               "New Test Project",
				Team:               "test@instill.tech",
				ShouldSkipWeekends: true,
			},
			wantResp: asana.ProjectTaskOutput{
				Project: asana.Project{
					GID:            "4321",
					Name:           "New Test Project",
					Owner:          asana.User{GID: "123", Name: "Admin User"},
					Notes:          "Test Notes",
					HTMLNotes:      "Test HTML Notes",
					DueOn:          "2021-01-01",
					StartOn:        "2021-01-01",
					Archived:       true,
					Color:          "red",
					Completed:      true,
					ModifiedAt:     "2021-01-01",
					PrivacySetting: "public_to_workspace",
					CompletedBy:    asana.User{GID: "123", Name: "Admin User"},
					CurrentStatus: map[string]string{
						"status": "completed",
					},
					CustomFields: map[string]string{
						"field": "value",
					},
					CustomFieldSettings: map[string]string{
						"field": "value",
					},
				},
			},
		},
		{
			_type: "nok",
			name:  "Duplicate project - 404 Not Found",
			input: asana.DuplicateProjectInput{
				Action: "duplicate",
				ID:     "12345",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, asana.TaskAsanaProject, t)
}
