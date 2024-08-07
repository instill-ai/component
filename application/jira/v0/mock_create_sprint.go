package jira

import (
	"encoding/json"
	"net/http"

	"github.com/instill-ai/component/tools/logger"
)

type mockCreateSprintRequest struct {
	Name          string `json:"name"`
	Goal          string `json:"goal"`
	StartDate     string `json:"startDate"`
	EndDate       string `json:"endDate"`
	OriginBoardID int    `json:"originBoardId"`
}

func mockCreateSprint(res http.ResponseWriter, req *http.Request) {
	var debug logger.Session
	defer debug.SessionStart("MockCreateIssueTask", logger.Develop).SessionEnd()
	var err error

	debug.Info("MockCreateIssueTask called")
	debug.Info(req.Method)
	if req.Method != http.MethodPost {
		http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body := mockCreateSprintRequest{}
	err = json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		http.Error(res, "Bad Request", http.StatusBadRequest)
		return
	}
	var newSprint = FakeSprint{
		ID:            1,
		Self:          "https://test.atlassian.net/rest/agile/1.0/sprint/1",
		State:         "active",
		Name:          body.Name,
		StartDate:     body.StartDate,
		EndDate:       body.EndDate,
		CompleteDate:  "",
		OriginBoardID: body.OriginBoardID,
		Goal:          body.Goal,
	}
	fakeSprints = append(fakeSprints, newSprint)

	res.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(res).Encode(newSprint)
	if err != nil {
		http.Error(res, "Bad Request", http.StatusBadRequest)
		return
	}
}
