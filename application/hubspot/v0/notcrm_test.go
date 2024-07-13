package hubspot

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

// This file is for testing tasks that are not CRM: Threads and Retrieve Association
// crm_test.go is for  Contacts, Deals, Companies, Tickets
// MockClient is created in crm_test.go

// Mock Thread struct and its functions
type MockThread struct{}

func (s *MockThread) Get(threadId string) (*ThreadResponseHSFormat, error) {

	var fakeThread ThreadResponseHSFormat
	if threadId == "7509711154" {
		fakeThread = ThreadResponseHSFormat{
			Results: []ThreadResultHSFormat{
				{
					CreatedAt: "2024-07-02T10:42:15Z",
					Senders: []ThreadUserHSFormat{
						{
							Name: "Brian Halligan (Sample Contact)",
							DeliveryIdentifier: ThreadDeliveryIdentifier{
								Type:  "HS_EMAIL_ADDRESS",
								Value: "bh@hubspot.com",
							},
						},
					},
					Recipients: []ThreadUserHSFormat{
						{
							DeliveryIdentifier: ThreadDeliveryIdentifier{
								Type:  "HS_EMAIL_ADDRESS",
								Value: "fake_email@gmail.com",
							},
						},
					},
					Text:    "Just random content inside",
					Subject: "A fake message",
				},
			},
		}
	}

	return &fakeThread, nil
}

// Mock Retrieve Association struct and its functions

type MockRetrieveAssociation struct{}

func (s *MockRetrieveAssociation) GetThreadId(contactId string) (*RetrieveThreadIdResponse, error) {

	var fakeThreadId RetrieveThreadIdResponse
	if contactId == "32027696539" {
		fakeThreadId = RetrieveThreadIdResponse{
			Results: []RetrieveThreadIdResult{
				{
					Id: "7509711154",
				},
			},
		}
	}
	return &fakeThreadId, nil
}

func (s *MockRetrieveAssociation) GetCrmId(contactId string, objectType string) (*RetrieveCrmIdResponseHSFormat, error) {

	var fakeCrmId RetrieveCrmIdResponseHSFormat
	if contactId == "32027696539" {
		fakeCrmId = RetrieveCrmIdResponseHSFormat{
			Results: []RetrieveCrmIdResultHSFormat{
				{
					IdArray: []RetrieveCrmId{
						{
							Id: "12345678900",
						},
					},
				},
			},
		}
	}
	return &fakeCrmId, nil

}

// Testing functions

func TestComponent_ExecuteGetThreadTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	tc := struct {
		name     string
		input    string
		wantResp ThreadResponseTaskFormat
	}{
		name:  "ok - get thread",
		input: "7509711154",
		wantResp: ThreadResponseTaskFormat{
			Results: []ThreadResultTaskFormat{
				{
					CreatedAt: "2024-07-02T10:42:15Z",
					Senders: []ThreadUserTaskFormat{
						{
							Name:  "Brian Halligan (Sample Contact)",
							Type:  "HS_EMAIL_ADDRESS",
							Value: "bh@hubspot.com",
						},
					},
					Recipients: []ThreadUserTaskFormat{
						{
							Type:  "HS_EMAIL_ADDRESS",
							Value: "fake_email@gmail.com",
						},
					},
					Text:    "Just random content inside",
					Subject: "A fake message",
				},
			},
		},
	}

	c.Run(tc.name, func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"token": bearerToken,
		})
		c.Assert(err, qt.IsNil)

		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: taskGetThread},
			client:             createMockClient(),
		}
		exec := &base.ExecutionWrapper{Execution: e}

		pbInput, err := structpb.NewStruct(map[string]any{
			"thread-id": tc.input,
		})

		c.Assert(err, qt.IsNil)

		res, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbInput})
		c.Assert(err, qt.IsNil)

		resJSON, err := protojson.Marshal(res[0])
		c.Assert(err, qt.IsNil)

		c.Check(resJSON, qt.JSONEquals, tc.wantResp)

	})
}

func TestComponent_ExecuteRetrieveAssociationTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    RetrieveAssociationInput
		wantResp interface{}
	}{
		{
			name: "ok - retrieve association: thread ID",
			input: RetrieveAssociationInput{
				ContactId:  "32027696539",
				ObjectType: "Threads",
			},
			wantResp: RetrieveThreadIdResponse{
				Results: []RetrieveThreadIdResult{
					{
						Id: "7509711154",
					},
				},
			},
		},
		{
			name: "ok - retrieve association: deal ID",
			input: RetrieveAssociationInput{
				ContactId:  "32027696539",
				ObjectType: "Deals",
			},
			wantResp: RetrieveCrmIdResultTaskFormat{
				IdArray: []RetrieveCrmId{
					{
						Id: "12345678900",
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"token": bearerToken,
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: taskRetrieveAssociation},
				client:             createMockClient(),
			}
			exec := &base.ExecutionWrapper{Execution: e}

			pbInput, err := base.ConvertToStructpb(tc.input)

			c.Assert(err, qt.IsNil)

			res, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbInput})
			c.Assert(err, qt.IsNil)

			resJSON, err := protojson.Marshal(res[0])
			c.Assert(err, qt.IsNil)

			c.Check(resJSON, qt.JSONEquals, tc.wantResp)

		})
	}

}
