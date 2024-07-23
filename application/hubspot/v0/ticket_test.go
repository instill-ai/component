package hubspot

import (
	"context"
	"testing"

	hubspot "github.com/belong-inc/go-hubspot"
	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

// mockClient is in contact_test.go

// Mock Ticket struct and its functions
type MockTicket struct{}

func (s *MockTicket) Get(ticketId string) (*hubspot.ResponseResource, error) {
	var fakeTicket TaskGetTicketResp
	if ticketId == "2865646368" {
		fakeTicket = TaskGetTicketResp{
			TicketName:   "HubSpot - New Query (Sample Query)",
			TicketStatus: "1",
			Pipeline:     "0",
			Category:     "PRODUCT_ISSUE;BILLING_ISSUE",
		}
	}

	ret := &hubspot.ResponseResource{
		Properties: &fakeTicket,
	}

	return ret, nil
}
func (s *MockTicket) Create(ticket *TaskCreateTicketReq) (*hubspot.ResponseResource, error) {
	arbitraryTicketId := "99987654321"

	fakeTicketInfo := ticket

	fakeTicketInfo.TicketId = arbitraryTicketId

	ret := &hubspot.ResponseResource{
		Properties: fakeTicketInfo,
	}

	return ret, nil
}

func TestComponent_ExecuteGetTicketTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	tc := struct {
		name     string
		input    string
		wantResp TaskGetTicketOutput
	}{
		name:  "ok - get ticket",
		input: "2865646368",
		wantResp: TaskGetTicketOutput{
			TicketName:   "HubSpot - New Query (Sample Query)",
			TicketStatus: "1",
			Pipeline:     "0",
			Category:     []string{"PRODUCT_ISSUE", "BILLING_ISSUE"},
		},
	}

	c.Run(tc.name, func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"token": bearerToken,
		})
		c.Assert(err, qt.IsNil)

		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: taskGetTicket},
			client:             createMockClient(),
		}
		e.execute = e.GetTicket
		exec := &base.ExecutionWrapper{Execution: e}

		pbInput, err := structpb.NewStruct(map[string]any{
			"ticket-id": tc.input,
		})

		c.Assert(err, qt.IsNil)

		res, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbInput})

		c.Assert(err, qt.IsNil)

		resJSON, err := protojson.Marshal(res[0])
		c.Assert(err, qt.IsNil)

		c.Check(resJSON, qt.JSONEquals, tc.wantResp)

	})
}

func TestComponent_ExecuteCreateTicketTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	tc := struct {
		name           string
		inputTicket    TaskCreateTicketInput
		inputContactId string //used to associate contact with ticket
		wantResp       string
	}{
		name: "ok - create ticket",
		inputTicket: TaskCreateTicketInput{
			TicketName:   "Fake Ticket",
			TicketStatus: "2",
			Pipeline:     "0",
			Category:     []string{"FEATURE_REQUEST", "GENERAL_INQUIRY"},
		},
		inputContactId: "32027696539",
		wantResp:       "99987654321",
	}

	c.Run(tc.name, func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"token": bearerToken,
		})
		c.Assert(err, qt.IsNil)

		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: taskCreateTicket},
			client:             createMockClient(),
		}
		e.execute = e.CreateTicket
		exec := &base.ExecutionWrapper{Execution: e}

		pbInput, err := base.ConvertToStructpb(tc.inputTicket)
		pbInput.Fields["contact-id-or-email"] = structpb.NewStringValue(tc.inputContactId)

		c.Assert(err, qt.IsNil)

		res, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbInput})
		c.Assert(err, qt.IsNil)

		resString := res[0].Fields["ticket-id"].GetStringValue()

		c.Check(resString, qt.Equals, tc.wantResp)

	})
}
