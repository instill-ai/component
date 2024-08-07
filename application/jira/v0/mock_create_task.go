package jira

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/instill-ai/component/tools/logger"
)

type AdditionalFields struct {
	Add    map[string]interface{} `json:"add,omitempty"`
	Copy   map[string]interface{} `json:"copy,omitempty"`
	Set    map[string]interface{} `json:"set,omitempty"`
	Edit   map[string]interface{} `json:"edit,omitempty"`
	Remove map[string]interface{} `json:"remove,omitempty"`
}
type mockCreateIssueRequest struct {
	Fields map[string]interface{}        `json:"fields"`
	Update map[string][]AdditionalFields `json:"update"`
}

type mockCreateIssueResponse struct {
	ID         string `json:"id"`
	Key        string `json:"key"`
	Self       string `json:"self"`
	Transition struct {
		Status          string `json:"status"`
		ErrorCollection struct {
			ErrorMessages []string               `json:"errorMessages"`
			Errors        map[string]interface{} `json:"errors"`
		} `json:"errorCollection"`
	}
}

func mockCreateIssueTask(res http.ResponseWriter, req *http.Request) {
	var debug logger.Session
	defer debug.SessionStart("MockCreateIssueTask", logger.Develop).SessionEnd()

	var err error
	if req.Method != http.MethodPost {
		http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body := mockCreateIssueRequest{}
	err = json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fields := body.Fields
	update := body.Update
	debug.Info("body: ", body)
	debug.Info("Fields: ", fields)
	debug.Info("Update: ", update)

	project, ok := fields["project"].(map[string]interface{})["key"].(string)
	if !ok {
		http.Error(res, "Invalid project", http.StatusBadRequest)
		return
	}
	badResp := mockCreateIssueResponse{
		ID:   "",
		Key:  "",
		Self: "",
		Transition: struct {
			Status          string `json:"status"`
			ErrorCollection struct {
				ErrorMessages []string               `json:"errorMessages"`
				Errors        map[string]interface{} `json:"errors"`
			} `json:"errorCollection"`
		}{
			Status: "Failed",
			ErrorCollection: struct {
				ErrorMessages []string               `json:"errorMessages"`
				Errors        map[string]interface{} `json:"errors"`
			}{
				ErrorMessages: []string{"Invalid project"},
				Errors:        map[string]interface{}{},
			},
		},
	}
	if project == "INVALID" {
		res.WriteHeader(http.StatusBadRequest)
		err = json.NewEncoder(res).Encode(badResp)
		if err != nil {
			fmt.Println(err)
			return
		}
		return
	}
	key := project + "-1"
	ID := "30000"
	successResp := mockCreateIssueResponse{
		ID:   ID,
		Key:  key,
		Self: "http://localhost:8080/rest/api/2/issue/10000",
		Transition: struct {
			Status          string `json:"status"`
			ErrorCollection struct {
				ErrorMessages []string               `json:"errorMessages"`
				Errors        map[string]interface{} `json:"errors"`
			} `json:"errorCollection"`
		}{
			Status: "Success",
			ErrorCollection: struct {
				ErrorMessages []string               `json:"errorMessages"`
				Errors        map[string]interface{} `json:"errors"`
			}{
				ErrorMessages: []string{},
				Errors:        map[string]interface{}{},
			},
		},
	}
	res.WriteHeader(http.StatusOK)
	err = json.NewEncoder(res).Encode(successResp)
	if err != nil {
		fmt.Println(err)
		return
	}

	fakeIssues = append(fakeIssues, FakeIssue{
		ID:     ID,
		Key:    key,
		Fields: fields,
	})
}
