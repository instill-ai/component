package jira

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func router(res http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	var logger DebugSession
	logger.SessionStart("router", StaticVerboseLevel)
	defer logger.SessionEnd()
	switch path {
	case "/_edge/tenant_info":
		res.WriteHeader(http.StatusOK)
		_, err := res.Write([]byte(`{"id":"12345678-1234-1234-1234-123456789012"}`))
		if err != nil {
			fmt.Println("/_edge/tenant_info", err)
		}
	case "/rest/agile/1.0/board":
		mockListBoards(res, req)
	case "/rest/agile/1.0/board/1":
	case "/rest/api/2/issue":
	default:
		http.NotFound(res, req)
	}
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
