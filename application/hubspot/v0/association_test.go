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

// mockClient is in contact_test.go

// Mock Retrieve Association struct and its functions

type MockRetrieveAssociation struct{}

func (s *MockRetrieveAssociation) GetThreadID(contactID string) (*TaskRetrieveAssociationThreadResp, error) {

	var fakeThreadID TaskRetrieveAssociationThreadResp
	if contactID == "32027696539" {
		fakeThreadID = TaskRetrieveAssociationThreadResp{
			Results: []struct {
				ID string `json:"id"`
			}{
				{ID: "7509711154"},
			},
		}
	}
	return &fakeThreadID, nil
}

func (s *MockRetrieveAssociation) GetCrmID(contactID string, objectType string) (*TaskRetrieveAssociationCrmResp, error) {

	var fakeCrmID TaskRetrieveAssociationCrmResp
	if contactID == "32027696539" {
		fakeCrmID = TaskRetrieveAssociationCrmResp{
			Results: []taskRetrieveAssociationCrmRespResult{
				{
					IDArray: []struct {
						ID string `json:"id"`
					}{
						{ID: "12345678900"},
					},
				},
			},
		}
	}
	return &fakeCrmID, nil

}

func TestComponent_ExecuteRetrieveAssociationTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    TaskRetrieveAssociationInput
		wantResp interface{}
	}{
		{
			name: "ok - retrieve association: thread ID",
			input: TaskRetrieveAssociationInput{
				ContactID:  "32027696539",
				ObjectType: "Threads",
			},
			wantResp: TaskRetrieveAssociationOutput{
				ObjectIDs: []string{
					"7509711154",
				},
			},
		},
		{
			name: "ok - retrieve association: deal ID",
			input: TaskRetrieveAssociationInput{
				ContactID:  "32027696539",
				ObjectType: "Deals",
			},
			wantResp: TaskRetrieveAssociationOutput{
				ObjectIDs: []string{
					"12345678900",
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
			e.execute = e.RetrieveAssociation

			pbInput, err := base.ConvertToStructpb(tc.input)

			c.Assert(err, qt.IsNil)

			res, err := e.Execute(ctx, []*structpb.Struct{pbInput})
			c.Assert(err, qt.IsNil)

			resJSON, err := protojson.Marshal(res[0])
			c.Assert(err, qt.IsNil)

			c.Check(resJSON, qt.JSONEquals, tc.wantResp)

		})
	}

}
