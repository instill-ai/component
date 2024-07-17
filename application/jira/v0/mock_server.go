package jira

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func router(middlewares ...func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	for _, m := range middlewares {
		r.Use(m)
	}
	r.Get("/_edge/tenant_info", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"cloudId":"12345678-1234-1234-1234-123456789012"}`))
		if err != nil {
			fmt.Println("/_edge/tenant_info", err)
		}
	})
	r.Get("/rest/agile/1.0/board", mockListBoards)
	r.Get("/rest/agile/1.0/issue/{issueIdOrKey:[a-zA-z0-9-]+}", mockGetIssue)
	r.Get("/rest/agile/1.0/sprint/{sprintId}", mockGetSprint)
	return r
}

func mockListBoards(res http.ResponseWriter, req *http.Request) {
	var err error
	opt := req.URL.Query()
	boardType := opt.Get("type")
	startAt := opt.Get("startAt")
	maxResults := opt.Get("maxResults")
	name := opt.Get("name")
	projectKeyOrID := opt.Get("projectKeyOrId")
	// filter boards
	var boards []FakeBoard
	pjNotFound := projectKeyOrID != ""
	for _, board := range fakeBoards {
		if boardType != "" && board.BoardType != boardType {
			continue
		}
		if name != "" && !strings.Contains(board.Name, name) {
			continue
		}
		if projectKeyOrID != "" {
			if !strings.EqualFold(board.Name, projectKeyOrID) {
				continue
			}
			pjNotFound = false
		}
		boards = append(boards, board)
	}
	if pjNotFound {
		res.WriteHeader(http.StatusBadRequest)
		_, err := res.Write([]byte(fmt.Sprintf(`{"errorMessages":["No project could be found with key or id '%s'"]}`, projectKeyOrID)))
		if err != nil {
			fmt.Println("/rest/agile/1.0/board", err)
		}
		return
	}
	// pagination
	start, end := 0, len(boards)
	if startAt != "" {
		start, err = strconv.Atoi(startAt)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			_, err := res.Write([]byte(`{"errorMessages":["The 'startAt' parameter must be a number"]}`))
			if err != nil {
				fmt.Println("/rest/agile/1.0/board", err)
			}
			return
		}
	}
	maxResultsNum := len(boards)
	if maxResults != "" {
		maxResultsNum, err = strconv.Atoi(maxResults)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			_, err := res.Write([]byte(`{"errorMessages":["The 'maxResults' parameter must be a number"]}`))
			if err != nil {
				fmt.Println("/rest/agile/1.0/board", err)
			}
			return
		}
		end = start + maxResultsNum
		if end > len(boards) {
			end = len(boards)
		}
	}
	// response
	res.WriteHeader(http.StatusOK)
	respText := `{"values":[`
	if len(boards) != 0 {
		for i, board := range boards[start:end] {
			if i > 0 {
				respText += ","
			}
			respText += fmt.Sprintf(`{"id":%d,"name":"%s","type":"%s","self":"%s"}`, board.ID, board.Name, board.BoardType, board.getSelf())
		}
	}
	respText += `],`
	respText += `"total":` + strconv.Itoa(len(boards)) + `,"startAt":` + strconv.Itoa(start) + `,"maxResults":` + strconv.Itoa(maxResultsNum) + `,"isLast":` + strconv.FormatBool(end == len(boards)) + `}`
	_, err = res.Write([]byte(respText))
	if err != nil {
		fmt.Println("/rest/agile/1.0/board", err)
	}
}

func mockGetIssue(res http.ResponseWriter, req *http.Request) {
	var err error

	issueID := chi.URLParam(req, "issueIdOrKey")
	if issueID == "" {
		res.WriteHeader(http.StatusBadRequest)
		_, err := res.Write([]byte(`{"errorMessages":["Issue id or key is required"]}`))
		if err != nil {
			fmt.Println("/rest/agile/1.0/issue", err)
		}
		return
	}
	// find issue
	var issue *FakeIssue
	for _, i := range fakeIssues {
		if i.ID == issueID || i.Key == issueID {
			issue = &i
			issue.getSelf()
			break
		}
	}
	if issue == nil {
		res.WriteHeader(http.StatusNotFound)
		_, err := res.Write([]byte(`{"errorMessages":["Issue does not exist or you do not have permission to see it"]}`))
		if err != nil {
			fmt.Println("/rest/agile/1.0/issue", err)
		}
		return
	}
	fmt.Println(issue)
	// response
	res.WriteHeader(http.StatusOK)
	respText, err := json.Marshal(issue)
	if err != nil {
		fmt.Println("/rest/agile/1.0/issue", err)
	}
	_, err = res.Write(respText)
	if err != nil {
		fmt.Println("/rest/agile/1.0/issue", err)
	}
}

func mockGetSprint(res http.ResponseWriter, req *http.Request) {
	var err error
	sprintID := chi.URLParam(req, "sprintId")
	if sprintID == "" {
		res.WriteHeader(http.StatusBadRequest)
		_, err := res.Write([]byte(`{"errorMessages":["Sprint id is required"]}`))
		if err != nil {
			fmt.Println("/rest/agile/1.0/sprint", err)
		}
		return
	}
	// find sprint
	var sprint *FakeSprint
	for _, s := range fakeSprints {
		if strconv.Itoa(s.ID) == sprintID {
			sprint = &s
			sprint.getSelf()
			break
		}
	}
	if sprint == nil {
		res.WriteHeader(http.StatusNotFound)
		_, err := res.Write([]byte(`{"errorMessages":["Sprint does not exist or you do not have permission to see it"]}`))
		if err != nil {
			fmt.Println("/rest/agile/1.0/sprint", err)
		}
		return
	}
	// response
	res.WriteHeader(http.StatusOK)
	respText, err := json.Marshal(sprint)
	if err != nil {
		fmt.Println("/rest/agile/1.0/sprint", err)
	}
	_, err = res.Write(respText)
	if err != nil {
		fmt.Println("/rest/agile/1.0/sprint", err)
	}
}
