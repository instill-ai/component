package hubspot

import (
	"context"
	"testing"
	"time"

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

func (s *MockTicket) Get(ticketID string) (*hubspot.ResponseResource, error) {
	var fakeTicket TaskGetTicketResp
	if ticketID == "2865646368" {
		fakeTicket = TaskGetTicketResp{
			TicketName:       "HubSpot - New Query (Sample Query)",
			TicketStatus:     "1",
			Pipeline:         "0",
			Category:         "PRODUCT_ISSUE;BILLING_ISSUE",
			CreateDate:       hubspot.NewTime(time.Date(2024, 7, 9, 0, 0, 0, 0, time.UTC)),
			LastModifiedDate: hubspot.NewTime(time.Date(2024, 7, 9, 0, 0, 0, 0, time.UTC)),
		}
	}

	ret := &hubspot.ResponseResource{
		Properties: &fakeTicket,
	}

	return ret, nil
}
func (s *MockTicket) Create(ticket *TaskCreateTicketReq) (*hubspot.ResponseResource, error) {
	arbitraryTicketID := "99987654321"

	fakeTicketInfo := ticket

	fakeTicketInfo.TicketID = arbitraryTicketID

	ret := &hubspot.ResponseResource{
		Properties: fakeTicketInfo,
	}

	return ret, nil
}

func (s *MockTicket) Update(ticketID string, ticket *TaskUpdateTicketReq) (*hubspot.ResponseResource, error) {
	return nil, nil
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
			TicketName:       "HubSpot - New Query (Sample Query)",
			TicketStatus:     "1",
			Pipeline:         "0",
			Category:         []string{"PRODUCT_ISSUE", "BILLING_ISSUE"},
			CreateDate:       "2024-07-09 00:00:00 +0000 UTC",
			LastModifiedDate: "2024-07-09 00:00:00 +0000 UTC",
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
		

		pbInput, err := structpb.NewStruct(map[string]any{
			"ticket-id": tc.input,
		})

		c.Assert(err, qt.IsNil)

		res, err := e.Execute(ctx, []*structpb.Struct{pbInput})

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
		name        string
		inputTicket TaskCreateTicketInput
		wantResp    string
	}{
		name: "ok - create ticket",
		inputTicket: TaskCreateTicketInput{
			TicketName:   "Fake Ticket",
			TicketStatus: "2",
			Pipeline:     "0",
			Category:     []string{"FEATURE_REQUEST", "GENERAL_INQUIRY"},
		},
		wantResp: "99987654321",
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
		

		pbInput, err := base.ConvertToStructpb(tc.inputTicket)

		c.Assert(err, qt.IsNil)

		res, err := e.Execute(ctx, []*structpb.Struct{pbInput})
		c.Assert(err, qt.IsNil)

		resString := res[0].Fields["ticket-id"].GetStringValue()

		c.Check(resString, qt.Equals, tc.wantResp)

	})
}
