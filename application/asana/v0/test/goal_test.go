package asana

import (
	"testing"

	"github.com/instill-ai/component/application/asana/v0"
)

func TestGetGoal(t *testing.T) {
	testcases := []taskCase[asana.GetGoalInput, asana.GoalTaskOutput]{
		{
			_type: "ok",
			name:  "Get goal",
			input: asana.GetGoalInput{
				Action: "get",
				ID:     "1234",
			},
			wantResp: asana.GoalTaskOutput{
				Goal: asana.Goal{
					GID:       "1234",
					Name:      "Test Goal",
					Owner:     asana.User{GID: "123", Name: "Admin User"},
					Notes:     "Test Notes",
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
				},
			},
		},
		{
			_type: "nok",
			name:  "Get goal - 404 Not Found",
			input: asana.GetGoalInput{
				Action: "get",
				ID:     "12345",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, asana.TaskAsanaGoal, t)
}
func TestUpdateGoal(t *testing.T) {
	testcases := []taskCase[asana.UpdateGoalInput, asana.GoalTaskOutput]{
		{
			_type: "ok",
			name:  "Update goal",
			input: asana.UpdateGoalInput{
				Action: "update",
				ID:     "1234",
				Notes:  "Modified Notes",
				DueOn:  "2021-01-02",
				Liked:  true,
			},
			wantResp: asana.GoalTaskOutput{
				Goal: asana.Goal{
					GID:       "1234",
					Name:      "Test Goal",
					Owner:     asana.User{GID: "123", Name: "Admin User"},
					Notes:     "Modified Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-02",
					StartOn:   "2021-01-01",
					Liked:     true,
					Likes: []asana.Like{
						{
							LikeGID:  "123",
							UserGID:  "123",
							UserName: "Admin User",
						},
					},
				},
			},
		},
		{
			_type: "nok",
			name:  "Update goal - 404 Not Found",
			input: asana.UpdateGoalInput{
				Action: "update",
				ID:     "12345",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, asana.TaskAsanaGoal, t)
}
func TestCreateGoal(t *testing.T) {
	testcases := []taskCase[asana.CreateGoalInput, asana.GoalTaskOutput]{
		{
			_type: "ok",
			name:  "Create goal",
			input: asana.CreateGoalInput{
				Action:  "create",
				Name:    "Test Goal",
				Notes:   "Modified Notes",
				DueOn:   "2021-01-02",
				StartOn: "2021-01-01",
				Liked:   true,
			},
			wantResp: asana.GoalTaskOutput{
				Goal: asana.Goal{
					GID:       "123456789",
					Name:      "Test Goal",
					Owner:     asana.User{GID: "123", Name: "Admin User"},
					Notes:     "Modified Notes",
					HTMLNotes: "Test HTML Notes",
					DueOn:     "2021-01-02",
					StartOn:   "2021-01-01",
					Liked:     true,
					Likes: []asana.Like{
						{
							LikeGID:  "123",
							UserGID:  "123",
							UserName: "Admin User",
						},
					},
				},
			},
		},
		{
			_type: "nok",
			name:  "Create goal - 400 Bad Request",
			input: asana.CreateGoalInput{
				Action: "create",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, asana.TaskAsanaGoal, t)
}

func TestDeleteGoal(t *testing.T) {
	testcases := []taskCase[asana.DeleteGoalInput, asana.GoalTaskOutput]{
		{
			_type: "ok",
			name:  "Delete goal",
			input: asana.DeleteGoalInput{
				Action: "delete",
				ID:     "1234567890",
			},
			wantResp: asana.GoalTaskOutput{
				Goal: asana.Goal{},
			},
		},
		{
			_type: "nok",
			name:  "Delete goal - 404 Not Found",
			input: asana.DeleteGoalInput{
				Action: "delete",
				ID:     "12345",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, asana.TaskAsanaGoal, t)
}
