package asana

import (
	"testing"

	"github.com/instill-ai/component/application/asana/v0"
)

func TestGetPortfolio(t *testing.T) {
	testcases := []taskCase[asana.GetPortfolioInput, asana.PortfolioTaskOutput]{
		{
			_type: "ok",
			name:  "Get portfolio",
			input: asana.GetPortfolioInput{
				Action: "get",
				ID:     "1234",
			},
			wantResp: asana.PortfolioTaskOutput{
				Portfolio: asana.Portfolio{
					GID:                 "1234",
					Name:                "Test Portfolio",
					Owner:               asana.User{GID: "123", Name: "Admin User"},
					DueOn:               "2021-01-01",
					StartOn:             "2021-01-01",
					Color:               "red",
					Public:              true,
					CreatedBy:           asana.User{GID: "123", Name: "Admin User"},
					CurrentStatus:       map[string]string{"title": "On track"},
					CustomFields:        map[string]string{"field": "value"},
					CustomFieldSettings: map[string]string{"field": "value"},
				},
			},
		},
		{
			_type: "nok",
			name:  "Get portfolio - 404 Not Found",
			input: asana.GetPortfolioInput{
				Action: "get",
				ID:     "12345",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, asana.TaskAsanaPortfolio, t)
}
func TestUpdatePortfolio(t *testing.T) {
	testcases := []taskCase[asana.UpdatePortfolioInput, asana.PortfolioTaskOutput]{
		{
			_type: "ok",
			name:  "Update portfolio",
			input: asana.UpdatePortfolioInput{
				Action: "update",
				ID:     "1234",
				Public: false,
				Color:  "blue",
			},
			wantResp: asana.PortfolioTaskOutput{
				Portfolio: asana.Portfolio{
					GID:                 "1234",
					Name:                "Test Portfolio",
					Owner:               asana.User{GID: "123", Name: "Admin User"},
					DueOn:               "2021-01-01",
					StartOn:             "2021-01-01",
					Color:               "blue",
					Public:              false,
					CreatedBy:           asana.User{GID: "123", Name: "Admin User"},
					CurrentStatus:       map[string]string{"title": "On track"},
					CustomFields:        map[string]string{"field": "value"},
					CustomFieldSettings: map[string]string{"field": "value"},
				},
			},
		},
		{
			_type: "nok",
			name:  "Update portfolio - 404 Not Found",
			input: asana.UpdatePortfolioInput{
				Action: "update",
				ID:     "12345",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, asana.TaskAsanaPortfolio, t)
}
func TestCreatePortfolio(t *testing.T) {
	testcases := []taskCase[asana.CreatePortfolioInput, asana.PortfolioTaskOutput]{
		{
			_type: "ok",
			name:  "Create portfolio",
			input: asana.CreatePortfolioInput{
				Action:    "create",
				Name:      "Test Portfolio",
				Color:     "blue",
				Public:    true,
				Workspace: "123",
			},
			wantResp: asana.PortfolioTaskOutput{
				Portfolio: asana.Portfolio{
					GID:                 "123456789",
					Name:                "Test Portfolio",
					Owner:               asana.User{GID: "123", Name: "Admin User"},
					DueOn:               "2021-01-01",
					StartOn:             "2021-01-01",
					Color:               "blue",
					Public:              true,
					CreatedBy:           asana.User{GID: "123", Name: "Admin User"},
					CurrentStatus:       map[string]string{"title": "On track"},
					CustomFields:        map[string]string{"field": "value"},
					CustomFieldSettings: map[string]string{"field": "value"},
				},
			},
		},
		{
			_type: "nok",
			name:  "Create portfolio - 400 Bad Request",
			input: asana.CreatePortfolioInput{
				Action: "create",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, asana.TaskAsanaPortfolio, t)
}

func TestDeletePortfolio(t *testing.T) {
	testcases := []taskCase[asana.DeletePortfolioInput, asana.PortfolioTaskOutput]{
		{
			_type: "ok",
			name:  "Delete portfolio",
			input: asana.DeletePortfolioInput{
				Action: "delete",
				ID:     "1234567890",
			},
			wantResp: asana.PortfolioTaskOutput{
				Portfolio: asana.Portfolio{},
			},
		},
		{
			_type: "nok",
			name:  "Delete portfolio - 404 Not Found",
			input: asana.DeletePortfolioInput{
				Action: "delete",
				ID:     "12345",
			},
			wantErr: `unsuccessful HTTP response.*`,
		},
	}
	taskTesting(testcases, asana.TaskAsanaPortfolio, t)
}
