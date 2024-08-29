package asana

import (
	"testing"
)

func TestGetProject(t *testing.T) {
	testcases := []taskCase[GetProjectInput, ProjectTaskOutput]{
		{
			_type: "ok",
			name:  "Get project",
			input: GetProjectInput{
				Action: "get",
				ID:     "1234",
			},
			wantResp: ProjectTaskOutput{
				Project: Project{
					GID:            "1234",
					Name:           "Test Project",
					Owner:          User{GID: "123", Name: "Admin User"},
					Notes:          "Test Notes",
					HTMLNotes:      "Test HTML Notes",
					DueOn:          "2021-01-01",
					StartOn:        "2021-01-01",
					Archived:       true,
					Color:          "red",
					Completed:      true,
					ModifiedAt:     "2021-01-01",
					PrivacySetting: "public_to_workspace",
					CompletedBy:    User{GID: "123", Name: "Admin User"},
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
			input: GetProjectInput{
				Action: "get",
				ID:     "12345",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, TaskAsanaProject, t)
}
func TestUpdateProject(t *testing.T) {
	testcases := []taskCase[UpdateProjectInput, ProjectTaskOutput]{
		{
			_type: "ok",
			name:  "Update project",
			input: UpdateProjectInput{
				Action:   "update",
				ID:       "123",
				Notes:    "Modified Notes",
				DueOn:    "2021-01-02",
				Archived: true,
			},
			wantResp: ProjectTaskOutput{
				Project: Project{
					GID:            "123",
					Name:           "Test Project",
					Owner:          User{GID: "123", Name: "Admin User"},
					Notes:          "Modified Notes",
					HTMLNotes:      "Test HTML Notes",
					DueOn:          "2021-01-02",
					StartOn:        "2021-01-01",
					Archived:       true,
					Color:          "red",
					Completed:      true,
					ModifiedAt:     "2021-01-01",
					PrivacySetting: "public_to_workspace",
					CompletedBy:    User{GID: "123", Name: "Admin User"},
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
			input: UpdateProjectInput{
				Action: "update",
				ID:     "12345",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, TaskAsanaProject, t)
}
func TestCreateProject(t *testing.T) {
	testcases := []taskCase[CreateProjectInput, ProjectTaskOutput]{
		{
			_type: "ok",
			name:  "Create project",
			input: CreateProjectInput{
				Action:         "create",
				Name:           "Test Project",
				Notes:          "Test Notes",
				DueOn:          "2021-01-02",
				StartOn:        "2021-01-01",
				Color:          "red",
				PrivacySetting: "public to workspace",
			},
			wantResp: ProjectTaskOutput{
				Project: Project{
					GID:            "123456789",
					Name:           "Test Project",
					Owner:          User{GID: "123", Name: "Admin User"},
					Notes:          "Test Notes",
					HTMLNotes:      "Test HTML Notes",
					DueOn:          "2021-01-02",
					StartOn:        "2021-01-01",
					Color:          "red",
					PrivacySetting: "public_to_workspace",
					Completed:      false,
					Archived:       false,
					CompletedBy:    User{GID: "123", Name: "Admin User"},
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
			input: CreateProjectInput{
				Action: "create",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, TaskAsanaProject, t)
}

func TestDeleteProject(t *testing.T) {
	testcases := []taskCase[DeleteProjectInput, ProjectTaskOutput]{
		{
			_type: "ok",
			name:  "Delete project",
			input: DeleteProjectInput{
				Action: "delete",
				ID:     "1234567890",
			},
			wantResp: ProjectTaskOutput{
				Project: Project{},
			},
		},
		{
			_type: "nok",
			name:  "Delete project - 404 Not Found",
			input: DeleteProjectInput{
				Action: "delete",
				ID:     "12345",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, TaskAsanaProject, t)
}

func TestDuplicateProject(t *testing.T) {
	testcases := []taskCase[DuplicateProjectInput, ProjectTaskOutput]{
		{
			_type: "ok",
			name:  "Duplicate project",
			input: DuplicateProjectInput{
				Action:             "duplicate",
				ID:                 "1234",
				Name:               "New Test Project",
				Team:               "test@instill.tech",
				ShouldSkipWeekends: true,
			},
			wantResp: ProjectTaskOutput{
				Project: Project{
					GID:            "4321",
					Name:           "New Test Project",
					Owner:          User{GID: "123", Name: "Admin User"},
					Notes:          "Test Notes",
					HTMLNotes:      "Test HTML Notes",
					DueOn:          "2021-01-01",
					StartOn:        "2021-01-01",
					Archived:       true,
					Color:          "red",
					Completed:      true,
					ModifiedAt:     "2021-01-01",
					PrivacySetting: "public_to_workspace",
					CompletedBy:    User{GID: "123", Name: "Admin User"},
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
			input: DuplicateProjectInput{
				Action: "duplicate",
				ID:     "12345",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, TaskAsanaProject, t)
}
