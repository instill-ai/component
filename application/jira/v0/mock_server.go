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
	r.Get("/rest/agile/1.0/issue/{issueIdOrKey:[a-zA-z0-9-]+}", mockGetIssue)
	r.Get("/rest/agile/1.0/sprint/{sprintId}", mockGetSprint)
	r.Get("/rest/agile/1.0/board/{boardId}/issue", mockListIssues)           // list all issues
	r.Get("/rest/agile/1.0/board/{boardId}/epic", mockListIssues)            // list all epic
	r.Get("/rest/agile/1.0/board/{boardId}/sprint", mockListIssues)          // list all sprint
	r.Get("/rest/agile/1.0/board/{boardId}/backlog", mockListIssues)         // list all issues in backlog
	r.Get("/rest/agile/1.0/board/{boardId}/epic/none/issue", mockListIssues) // list all issues without epic assigned
	r.Get("/rest/agile/1.0/board/{boardId}", mockGetBoard)
	r.Get("/rest/agile/1.0/board", mockListBoards)
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
func mockGetBoard(res http.ResponseWriter, req *http.Request) {
	var err error
	boardID := chi.URLParam(req, "boardId")
	// filter boards
	var board *FakeBoard
	for _, b := range fakeBoards {
		if boardID != "" && strconv.Itoa(b.ID) != boardID {
			continue
		}
		board = &b
	}
	if board == nil {
		res.WriteHeader(http.StatusNotFound)
		_, err := res.Write([]byte(`{"errorMessages":["Board does not exist or you do not have permission to see it"]}`))
		if err != nil {
			fmt.Println("mockGetBoard", err)
		}
		return
	}
	// response
	res.WriteHeader(http.StatusOK)
	respText, err := json.Marshal(board)
	if err != nil {
		fmt.Println("mockGetBoard", err)
	}
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

type MockListIssuesResponse struct {
	Issues     []FakeIssue `json:"issues"`
	Total      int         `json:"total"`
	StartAt    int         `json:"start-at"`
	MaxResults int         `json:"max-results"`
}

func mockListIssues(res http.ResponseWriter, req *http.Request) {
	var err error
	opt := req.URL.Query()
	boardID := chi.URLParam(req, "boardId")
	jql := opt.Get("jql")
	startAt := opt.Get("startAt")
	maxResults := opt.Get("maxResults")
	// find board
	var board *FakeBoard
	for _, b := range fakeBoards {
		if strconv.Itoa(b.ID) == boardID {
			board = &b
			break
		}
	}
	if board == nil {
		res.WriteHeader(http.StatusNotFound)
		_, err := res.Write([]byte(`{"errorMessages":["Board does not exist or you do not have permission to see it"]}`))
		if err != nil {
			fmt.Println("mockListIssues", err)
		}
		return
	}
	// filter issues
	var issues []FakeIssue
	for _, issue := range fakeIssues {
		prefix := strings.Split(issue.Key, "-")[0]
		if board.Name != "" && prefix != board.Name {
			fmt.Println("prefix", prefix, "board.Name", board.Name)
			continue
		}
		if jql != "" {
			// Skip JQL filter as there is no need to implement it
			continue
		}
		issue.getSelf()
		issues = append(issues, issue)
	}
	// response
	res.WriteHeader(http.StatusOK)
	startAtNum := 0
	if startAt != "" {
		startAtNum, err = strconv.Atoi(startAt)
		if err != nil {
			fmt.Println("mockListIssues", err)
			return
		}
	}
	maxResultsNum, err := strconv.Atoi(maxResults)
	if err != nil {
		fmt.Println("mockListIssues", err)
		return
	}
	resp := MockListIssuesResponse{
		Issues:     issues,
		Total:      len(issues),
		StartAt:    startAtNum,
		MaxResults: maxResultsNum,
	}
	respText, err := json.Marshal(resp)
	if err != nil {
		fmt.Println("mockListIssues", err)
	}
	_, err = res.Write([]byte(respText))
	if err != nil {
		fmt.Println("mockListIssues", err)
	}
}
