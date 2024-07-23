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

// Mock Thread struct and its functions
type MockThread struct{}

func (s *MockThread) Get(threadId string) (*TaskGetThreadResp, error) {

	var fakeThread TaskGetThreadResp
	if threadId == "7509711154" {
		fakeThread = TaskGetThreadResp{
			Results: []taskGetThreadRespResult{
				{
					CreatedAt: "2024-07-02T10:42:15Z",
					Senders: []taskGetThreadRespUser{
						{
							Name: "Brian Halligan (Sample Contact)",
							DeliveryIdentifier: taskGetThreadRespIdentifier{
								Type:  "HS_EMAIL_ADDRESS",
								Value: "bh@hubspot.com",
							},
						},
					},
					Recipients: []taskGetThreadRespUser{
						{
							DeliveryIdentifier: taskGetThreadRespIdentifier{
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

func TestComponent_ExecuteGetThreadTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	tc := struct {
		name     string
		input    string
		wantResp TaskGetThreadOutput
	}{
		name:  "ok - get thread",
		input: "7509711154",
		wantResp: TaskGetThreadOutput{
			Results: []taskGetThreadOutputResult{
				{
					CreatedAt: "2024-07-02T10:42:15Z",
					Senders: []taskGetThreadOutputUser{
						{
							Name:  "Brian Halligan (Sample Contact)",
							Type:  "HS_EMAIL_ADDRESS",
							Value: "bh@hubspot.com",
						},
					},
					Recipients: []taskGetThreadOutputUser{
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
		e.execute = e.GetThread
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
