package asana

import (
	"testing"
)

func TestGetTask(t *testing.T) {
	testcases := []taskCase[GetTaskInput, TaskTaskOutput]{
		{
			_type: "ok",
			name:  "Get task",
			input: GetTaskInput{
				Action: "get",
				ID:     "1234",
			},
			wantResp: TaskTaskOutput{
				Task: Task{
					GID:       "1234",
					Name:      "Test Task",
					Notes:     "Test Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-01",
					StartOn:   "2021-01-01",
					Projects: []SimpleProject{
						{
							GID:  "1234",
							Name: "Test Project",
						},
					},
					Liked: true,
					Likes: []Like{
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
			input: GetTaskInput{
				Action: "get",
				ID:     "12345",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, TaskAsanaTask, t)
}
func TestUpdateTask(t *testing.T) {
	testcases := []taskCase[UpdateTaskInput, TaskTaskOutput]{
		{
			_type: "ok",
			name:  "Update task",
			input: UpdateTaskInput{
				Action:         "update",
				ID:             "43210",
				Notes:          "Modified Notes",
				ApprovalStatus: "approved",
				Liked:          true,
			},
			wantResp: TaskTaskOutput{
				Task: Task{
					GID:       "43210",
					Name:      "Test Task",
					Notes:     "Modified Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-01",
					StartOn:   "2021-01-01",
					Liked:     true,
					Likes: []Like{
						{
							LikeGID:  "123",
							UserGID:  "123",
							UserName: "Admin User",
						},
					},
					Assignee: "123",
					Parent:   "1234",
					Projects: []SimpleProject{
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
			input: UpdateTaskInput{
				Action: "update",
				ID:     "12345",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, TaskAsanaTask, t)
}
func TestCreateTask(t *testing.T) {
	testcases := []taskCase[CreateTaskInput, TaskTaskOutput]{
		{
			_type: "ok",
			name:  "Create task",
			input: CreateTaskInput{
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
			wantResp: TaskTaskOutput{
				Task: Task{
					GID:       "123456789",
					Name:      "Test Task",
					Notes:     "Test Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-01",
					StartOn:   "2021-01-01",
					Completed: true,
					Liked:     true,
					Likes: []Like{
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
					Projects:        []SimpleProject{},
				},
			},
		},
		{
			_type: "nok",
			name:  "Create task - 400 Bad Request",
			input: CreateTaskInput{
				Action: "create",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, TaskAsanaTask, t)
}

func TestDeleteTask(t *testing.T) {
	testcases := []taskCase[DeleteTaskInput, TaskTaskOutput]{
		{
			_type: "ok",
			name:  "Delete task",
			input: DeleteTaskInput{
				Action: "delete",
				ID:     "1234567890",
			},
			wantResp: TaskTaskOutput{
				Task: Task{},
			},
		},
		{
			_type: "nok",
			name:  "Delete task - 404 Not Found",
			input: DeleteTaskInput{
				Action: "delete",
				ID:     "12345",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, TaskAsanaTask, t)
}
func TestDuplicateTask(t *testing.T) {
	testcases := []taskCase[DuplicateTaskInput, TaskTaskOutput]{
		{
			_type: "ok",
			name:  "Duplicate task",
			input: DuplicateTaskInput{
				Action: "duplicate",
				ID:     "1234",
				Name:   "Test Task",
			},
			wantResp: TaskTaskOutput{
				Task: Task{
					GID:       "4321",
					Name:      "Test Task",
					Notes:     "Test Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-01",
					StartOn:   "2021-01-01",
					Completed: true,
					Liked:     true,
					Likes: []Like{
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
					Projects: []SimpleProject{
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
			input: DuplicateTaskInput{
				Action: "duplicate",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, TaskAsanaTask, t)
}

func TestTaskSetParent(t *testing.T) {
	testcases := []taskCase[TaskSetParentInput, TaskTaskOutput]{
		{
			_type: "ok",
			name:  "Set Parent task",
			input: TaskSetParentInput{
				Action: "set parent",
				ID:     "1234",
				Parent: "1234",
			},
			wantResp: TaskTaskOutput{
				Task: Task{
					GID:       "1234",
					Name:      "Test Task",
					Notes:     "Test Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-01",
					StartOn:   "2021-01-01",
					Completed: true,
					Liked:     true,
					Likes: []Like{
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
					Projects: []SimpleProject{
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
			input: TaskSetParentInput{
				Action: "set parent",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, TaskAsanaTask, t)
}

func TestTaskEditTag(t *testing.T) {
	testcases := []taskCase[TaskEditTagInput, TaskTaskOutput]{
		{
			_type: "ok",
			name:  "Edit Tag task - add",
			input: TaskEditTagInput{
				Action:     "edit tag",
				ID:         "1234",
				TagID:      "1234",
				EditOption: "add",
			},
			wantResp: TaskTaskOutput{
				Task: Task{
					GID:       "1234",
					Name:      "Test Task",
					Notes:     "Test Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-01",
					StartOn:   "2021-01-01",
					Completed: true,
					Liked:     true,
					Likes: []Like{
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
					Projects: []SimpleProject{
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
			input: TaskEditTagInput{
				Action:     "edit tag",
				ID:         "1234",
				TagID:      "1234",
				EditOption: "remove",
			},
			wantResp: TaskTaskOutput{
				Task: Task{
					GID:       "1234",
					Name:      "Test Task",
					Notes:     "Test Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-01",
					StartOn:   "2021-01-01",
					Completed: true,
					Liked:     true,
					Likes: []Like{
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
					Projects: []SimpleProject{
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
			input: TaskEditTagInput{
				Action: "edit tag",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, TaskAsanaTask, t)
}

func TestTaskEditFollowers(t *testing.T) {
	testcases := []taskCase[TaskEditFollowerInput, TaskTaskOutput]{
		{
			_type: "ok",
			name:  "Edit Follower task - add",
			input: TaskEditFollowerInput{
				Action:     "edit follower",
				ID:         "1234",
				Followers:  "1234,test@instill.tech",
				EditOption: "add",
			},
			wantResp: TaskTaskOutput{
				Task: Task{
					GID:       "1234",
					Name:      "Test Task",
					Notes:     "Test Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-01",
					StartOn:   "2021-01-01",
					Completed: true,
					Liked:     true,
					Likes: []Like{
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
					Projects: []SimpleProject{
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
			input: TaskEditFollowerInput{
				Action:     "edit follower",
				ID:         "1234",
				Followers:  "1234",
				EditOption: "remove",
			},
			wantResp: TaskTaskOutput{
				Task: Task{
					GID:       "1234",
					Name:      "Test Task",
					Notes:     "Test Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-01",
					StartOn:   "2021-01-01",
					Completed: true,
					Liked:     true,
					Likes: []Like{
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
					Projects: []SimpleProject{
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
			input: TaskEditFollowerInput{
				Action: "edit follower",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, TaskAsanaTask, t)
}

func TestTaskEditProject(t *testing.T) {
	testcases := []taskCase[TaskEditProjectInput, TaskTaskOutput]{
		{
			_type: "ok",
			name:  "Edit Project task - add",
			input: TaskEditProjectInput{
				Action:     "edit project",
				ID:         "1234",
				ProjectID:  "1234",
				EditOption: "add",
			},
			wantResp: TaskTaskOutput{
				Task: Task{
					GID:       "1234",
					Name:      "Test Task",
					Notes:     "Test Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-01",
					StartOn:   "2021-01-01",
					Completed: true,
					Liked:     true,
					Likes: []Like{
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
					Projects: []SimpleProject{
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
			input: TaskEditProjectInput{
				Action:     "edit project",
				ID:         "1234",
				ProjectID:  "1234",
				EditOption: "remove",
			},
			wantResp: TaskTaskOutput{
				Task: Task{
					GID:       "1234",
					Name:      "Test Task",
					Notes:     "Test Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-01",
					StartOn:   "2021-01-01",
					Completed: true,
					Liked:     true,
					Likes: []Like{
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
					Projects: []SimpleProject{
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
			input: TaskEditProjectInput{
				Action: "edit project",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, TaskAsanaTask, t)
}
