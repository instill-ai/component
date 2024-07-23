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

func (s *MockRetrieveAssociation) GetThreadId(contactId string) (*TaskRetrieveAssociationThreadResp, error) {

	var fakeThreadId TaskRetrieveAssociationThreadResp
	if contactId == "32027696539" {
		fakeThreadId = TaskRetrieveAssociationThreadResp{
			Results: []struct {
				Id string `json:"id"`
			}{
				{Id: "7509711154"},
			},
		}
	}
	return &fakeThreadId, nil
}

func (s *MockRetrieveAssociation) GetCrmId(contactId string, objectType string) (*TaskRetrieveAssociationCrmResp, error) {

	var fakeCrmId TaskRetrieveAssociationCrmResp
	if contactId == "32027696539" {
		fakeCrmId = TaskRetrieveAssociationCrmResp{
			Results: []taskRetrieveAssociationCrmRespResult{
				{
					IdArray: []struct {
						Id string `json:"id"`
					}{
						{Id: "12345678900"},
					},
				},
			},
		}
	}
	return &fakeCrmId, nil

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
				ContactId:  "32027696539",
				ObjectType: "Threads",
			},
			wantResp: TaskRetrieveAssociationOutput{
				ObjectIds: []string{
					"7509711154",
				},
			},
		},
		{
			name: "ok - retrieve association: deal ID",
			input: TaskRetrieveAssociationInput{
				ContactId:  "32027696539",
				ObjectType: "Deals",
			},
			wantResp: TaskRetrieveAssociationOutput{
				ObjectIds: []string{
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
