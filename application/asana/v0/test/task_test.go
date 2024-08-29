package asana

import (
	"testing"

	"github.com/instill-ai/component/application/asana/v0"
)

func TestGetTask(t *testing.T) {
	testcases := []taskCase[asana.GetTaskInput, asana.TaskTaskOutput]{
		{
			_type: "ok",
			name:  "Get task",
			input: asana.GetTaskInput{
				Action: "get",
				ID:     "1234",
			},
			wantResp: asana.TaskTaskOutput{
				Task: asana.Task{
					GID:       "1234",
					Name:      "Test Task",
					Notes:     "Test Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-01",
					StartOn:   "2021-01-01",
					Projects: []asana.SimpleProject{
						{
							GID:  "1234",
							Name: "Test Project",
						},
					},
					Liked: true,
					Likes: []asana.Like{
						{
							LikeGID:  "123",
							UserGID:  "123",
							UserName: "Admin User",
						},
					},
					ApprovalStatus:  "approved",
					ResourceSubtype: "default_task",
					Assignee:        "123",
					Parent:          "1234",
					Completed:       true,
				},
			},
		},
		{
			_type: "nok",
			name:  "Get task - 404 Not Found",
			input: asana.GetTaskInput{
				Action: "get",
				ID:     "12345",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, asana.TaskAsanaTask, t)
}
func TestUpdateTask(t *testing.T) {
	testcases := []taskCase[asana.UpdateTaskInput, asana.TaskTaskOutput]{
		{
			_type: "ok",
			name:  "Update task",
			input: asana.UpdateTaskInput{
				Action:         "update",
				ID:             "43210",
				Notes:          "Modified Notes",
				ApprovalStatus: "approved",
				Liked:          true,
			},
			wantResp: asana.TaskTaskOutput{
				Task: asana.Task{
					GID:       "43210",
					Name:      "Test Task",
					Notes:     "Modified Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-01",
					StartOn:   "2021-01-01",
					Liked:     true,
					Likes: []asana.Like{
						{
							LikeGID:  "123",
							UserGID:  "123",
							UserName: "Admin User",
						},
					},
					Assignee: "123",
					Parent:   "1234",
					Projects: []asana.SimpleProject{
						{
							GID:  "1234",
							Name: "Test Project",
						},
					},
					ApprovalStatus:  "approved",
					ResourceSubtype: "default_task",
					Completed:       false,
				},
			},
		},
		{
			_type: "nok",
			name:  "Update task - 404 Not Found",
			input: asana.UpdateTaskInput{
				Action: "update",
				ID:     "12345",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, asana.TaskAsanaTask, t)
}
func TestCreateTask(t *testing.T) {
	testcases := []taskCase[asana.CreateTaskInput, asana.TaskTaskOutput]{
		{
			_type: "ok",
			name:  "Create task",
			input: asana.CreateTaskInput{
				Action:          "create",
				Name:            "Test Task",
				Notes:           "Test Notes",
				DueAt:           "2021-01-01",
				StartAt:         "2021-01-01",
				Liked:           true,
				ResourceSubtype: "default_task",
				ApprovalStatus:  "approved",
				Completed:       true,
				Assignee:        "123",
				Parent:          "1234",
			},
			wantResp: asana.TaskTaskOutput{
				Task: asana.Task{
					GID:       "123456789",
					Name:      "Test Task",
					Notes:     "Test Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-01",
					StartOn:   "2021-01-01",
					Completed: true,
					Liked:     true,
					Likes: []asana.Like{
						{
							LikeGID:  "123",
							UserGID:  "123",
							UserName: "Admin User",
						},
					},
					ApprovalStatus:  "approved",
					ResourceSubtype: "default_task",
					Assignee:        "123",
					Parent:          "1234",
					Projects:        []asana.SimpleProject{},
				},
			},
		},
		{
			_type: "nok",
			name:  "Create task - 400 Bad Request",
			input: asana.CreateTaskInput{
				Action: "create",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, asana.TaskAsanaTask, t)
}

func TestDeleteTask(t *testing.T) {
	testcases := []taskCase[asana.DeleteTaskInput, asana.TaskTaskOutput]{
		{
			_type: "ok",
			name:  "Delete task",
			input: asana.DeleteTaskInput{
				Action: "delete",
				ID:     "1234567890",
			},
			wantResp: asana.TaskTaskOutput{
				Task: asana.Task{},
			},
		},
		{
			_type: "nok",
			name:  "Delete task - 404 Not Found",
			input: asana.DeleteTaskInput{
				Action: "delete",
				ID:     "12345",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, asana.TaskAsanaTask, t)
}
func TestDuplicateTask(t *testing.T) {
	testcases := []taskCase[asana.DuplicateTaskInput, asana.TaskTaskOutput]{
		{
			_type: "ok",
			name:  "Duplicate task",
			input: asana.DuplicateTaskInput{
				Action: "duplicate",
				ID:     "1234",
				Name:   "Test Task",
			},
			wantResp: asana.TaskTaskOutput{
				Task: asana.Task{
					GID:       "4321",
					Name:      "Test Task",
					Notes:     "Test Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-01",
					StartOn:   "2021-01-01",
					Completed: true,
					Liked:     true,
					Likes: []asana.Like{
						{
							LikeGID:  "123",
							UserGID:  "123",
							UserName: "Admin User",
						},
					},
					ApprovalStatus:  "approved",
					ResourceSubtype: "default_task",
					Assignee:        "123",
					Parent:          "1234",
					Projects: []asana.SimpleProject{
						{
							GID:  "1234",
							Name: "Test Project",
						},
					},
				},
			},
		},
		{
			_type: "nok",
			name:  "Duplicate task - 400 Bad Request",
			input: asana.DuplicateTaskInput{
				Action: "duplicate",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, asana.TaskAsanaTask, t)
}

func TestTaskSetParent(t *testing.T) {
	testcases := []taskCase[asana.TaskSetParentInput, asana.TaskTaskOutput]{
		{
			_type: "ok",
			name:  "Set Parent task",
			input: asana.TaskSetParentInput{
				Action: "set parent",
				ID:     "1234",
				Parent: "1234",
			},
			wantResp: asana.TaskTaskOutput{
				Task: asana.Task{
					GID:       "1234",
					Name:      "Test Task",
					Notes:     "Test Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-01",
					StartOn:   "2021-01-01",
					Completed: true,
					Liked:     true,
					Likes: []asana.Like{
						{
							LikeGID:  "123",
							UserGID:  "123",
							UserName: "Admin User",
						},
					},
					ApprovalStatus:  "approved",
					ResourceSubtype: "default_task",
					Assignee:        "123",
					Parent:          "1234",
					Projects: []asana.SimpleProject{
						{
							GID:  "1234",
							Name: "Test Project",
						},
					},
				},
			},
		},
		{
			_type: "nok",
			name:  "SetParent task - 400 Bad Request",
			input: asana.TaskSetParentInput{
				Action: "set parent",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, asana.TaskAsanaTask, t)
}

func TestTaskEditTag(t *testing.T) {
	testcases := []taskCase[asana.TaskEditTagInput, asana.TaskTaskOutput]{
		{
			_type: "ok",
			name:  "Edit Tag task - add",
			input: asana.TaskEditTagInput{
				Action:     "edit tag",
				ID:         "1234",
				TagID:      "1234",
				EditOption: "add",
			},
			wantResp: asana.TaskTaskOutput{
				Task: asana.Task{
					GID:       "1234",
					Name:      "Test Task",
					Notes:     "Test Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-01",
					StartOn:   "2021-01-01",
					Completed: true,
					Liked:     true,
					Likes: []asana.Like{
						{
							LikeGID:  "123",
							UserGID:  "123",
							UserName: "Admin User",
						},
					},
					ApprovalStatus:  "approved",
					ResourceSubtype: "default_task",
					Assignee:        "123",
					Parent:          "1234",
					Projects: []asana.SimpleProject{
						{
							GID:  "1234",
							Name: "Test Project",
						},
					},
				},
			},
		},
		{
			_type: "ok",
			name:  "Edit Tag task - remove",
			input: asana.TaskEditTagInput{
				Action:     "edit tag",
				ID:         "1234",
				TagID:      "1234",
				EditOption: "remove",
			},
			wantResp: asana.TaskTaskOutput{
				Task: asana.Task{
					GID:       "1234",
					Name:      "Test Task",
					Notes:     "Test Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-01",
					StartOn:   "2021-01-01",
					Completed: true,
					Liked:     true,
					Likes: []asana.Like{
						{
							LikeGID:  "123",
							UserGID:  "123",
							UserName: "Admin User",
						},
					},
					ApprovalStatus:  "approved",
					ResourceSubtype: "default_task",
					Assignee:        "123",
					Parent:          "1234",
					Projects: []asana.SimpleProject{
						{
							GID:  "1234",
							Name: "Test Project",
						},
					},
				},
			},
		},
		{
			_type: "nok",
			name:  "EditTag task - 400 Bad Request",
			input: asana.TaskEditTagInput{
				Action: "edit tag",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, asana.TaskAsanaTask, t)
}

func TestTaskEditFollowers(t *testing.T) {
	testcases := []taskCase[asana.TaskEditFollowerInput, asana.TaskTaskOutput]{
		{
			_type: "ok",
			name:  "Edit Follower task - add",
			input: asana.TaskEditFollowerInput{
				Action:     "edit follower",
				ID:         "1234",
				Followers:  "1234,test@instill.tech",
				EditOption: "add",
			},
			wantResp: asana.TaskTaskOutput{
				Task: asana.Task{
					GID:       "1234",
					Name:      "Test Task",
					Notes:     "Test Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-01",
					StartOn:   "2021-01-01",
					Completed: true,
					Liked:     true,
					Likes: []asana.Like{
						{
							LikeGID:  "123",
							UserGID:  "123",
							UserName: "Admin User",
						},
					},
					ApprovalStatus:  "approved",
					ResourceSubtype: "default_task",
					Assignee:        "123",
					Parent:          "1234",
					Projects: []asana.SimpleProject{
						{
							GID:  "1234",
							Name: "Test Project",
						},
					},
				},
			},
		},
		{
			_type: "ok",
			name:  "Edit Follower task - remove",
			input: asana.TaskEditFollowerInput{
				Action:     "edit follower",
				ID:         "1234",
				Followers:  "1234",
				EditOption: "remove",
			},
			wantResp: asana.TaskTaskOutput{
				Task: asana.Task{
					GID:       "1234",
					Name:      "Test Task",
					Notes:     "Test Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-01",
					StartOn:   "2021-01-01",
					Completed: true,
					Liked:     true,
					Likes: []asana.Like{
						{
							LikeGID:  "123",
							UserGID:  "123",
							UserName: "Admin User",
						},
					},
					ApprovalStatus:  "approved",
					ResourceSubtype: "default_task",
					Assignee:        "123",
					Parent:          "1234",
					Projects: []asana.SimpleProject{
						{
							GID:  "1234",
							Name: "Test Project",
						},
					},
				},
			},
		},
		{
			_type: "nok",
			name:  "EditFollower task - 400 Bad Request",
			input: asana.TaskEditFollowerInput{
				Action: "edit follower",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, asana.TaskAsanaTask, t)
}

func TestTaskEditProject(t *testing.T) {
	testcases := []taskCase[asana.TaskEditProjectInput, asana.TaskTaskOutput]{
		{
			_type: "ok",
			name:  "Edit Project task - add",
			input: asana.TaskEditProjectInput{
				Action:     "edit project",
				ID:         "1234",
				ProjectID:  "1234",
				EditOption: "add",
			},
			wantResp: asana.TaskTaskOutput{
				Task: asana.Task{
					GID:       "1234",
					Name:      "Test Task",
					Notes:     "Test Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-01",
					StartOn:   "2021-01-01",
					Completed: true,
					Liked:     true,
					Likes: []asana.Like{
						{
							LikeGID:  "123",
							UserGID:  "123",
							UserName: "Admin User",
						},
					},
					ApprovalStatus:  "approved",
					ResourceSubtype: "default_task",
					Assignee:        "123",
					Parent:          "1234",
					Projects: []asana.SimpleProject{
						{
							GID:  "1234",
							Name: "Test Project",
						},
					},
				},
			},
		},
		{
			_type: "ok",
			name:  "EditProject task - remove",
			input: asana.TaskEditProjectInput{
				Action:     "edit project",
				ID:         "1234",
				ProjectID:  "1234",
				EditOption: "remove",
			},
			wantResp: asana.TaskTaskOutput{
				Task: asana.Task{
					GID:       "1234",
					Name:      "Test Task",
					Notes:     "Test Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-01",
					StartOn:   "2021-01-01",
					Completed: true,
					Liked:     true,
					Likes: []asana.Like{
						{
							LikeGID:  "123",
							UserGID:  "123",
							UserName: "Admin User",
						},
					},
					ApprovalStatus:  "approved",
					ResourceSubtype: "default_task",
					Assignee:        "123",
					Parent:          "1234",
					Projects: []asana.SimpleProject{
						{
							GID:  "1234",
							Name: "Test Project",
						},
					},
				},
			},
		},
		{
			_type: "nok",
			name:  "Edit Project task - 400 Bad Request",
			input: asana.TaskEditProjectInput{
				Action: "edit project",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, asana.TaskAsanaTask, t)
}
