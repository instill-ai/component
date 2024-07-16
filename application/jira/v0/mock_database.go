package jira

import "fmt"

var fakeBoards = []FakeBoard{
	{
		Board: Board{
			ID:        1,
			Name:      "KAN",
			BoardType: "kanban",
		},
	},
	{
		Board: Board{
			ID:        2,
			Name:      "SCR",
			BoardType: "scrum",
		},
	},
	{
		Board: Board{
			ID:        3,
			Name:      "TST",
			BoardType: "simple",
		},
	},
}

type FakeBoard struct {
	Board
}

func (f *FakeBoard) getSelf() string {
	if f.Self == "" {
		f.Self = fmt.Sprintf("https://test.atlassian.net/rest/agile/1.0/board/%d", f.ID)
	}
	return f.Self
}
