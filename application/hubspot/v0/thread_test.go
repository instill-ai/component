package hubspot

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/mock"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

// mockClient is in contact_test.go

// Mock Thread struct and its functions
type MockThread struct{}

func (s *MockThread) Get(threadID string) (*TaskGetThreadResp, error) {

	var fakeThread TaskGetThreadResp
	if threadID == "7509711154" {
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
					Text:             "Just random content inside",
					Subject:          "A fake message",
					ChannelID:        "1002",
					ChannelAccountID: "638727358",
					Type:             "MESSAGE",
				},
			},
		}
	}

	return &fakeThread, nil
}

func (s *MockThread) Insert(threadID string, message *TaskInsertMessageReq) (*TaskInsertMessageResp, error) {

	res := &TaskInsertMessageResp{}
	if threadID == "7509711154" {
		res.Status = taskInsertMessageRespStatusType{
			StatusType: "SENT",
		}
	}
	return res, nil
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
					Sender: taskGetThreadOutputSender{
						Name:  "Brian Halligan (Sample Contact)",
						Type:  "HS_EMAIL_ADDRESS",
						Value: "bh@hubspot.com",
					},
					Recipients: []taskGetThreadOutputRecipient{
						{
							Type:  "HS_EMAIL_ADDRESS",
							Value: "fake_email@gmail.com",
						},
					},
					Text:             "Just random content inside",
					Subject:          "A fake message",
					ChannelID:        "1002",
					ChannelAccountID: "638727358",
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

		pbInput, err := structpb.NewStruct(map[string]any{
			"thread-id": tc.input,
		})

		c.Assert(err, qt.IsNil)

		ir := mock.NewInputReaderMock(c)
		ow := mock.NewOutputWriterMock(c)
		ir.ReadMock.Return([]*structpb.Struct{pbInput}, nil)
		ow.WriteMock.Optional().Set(func(ctx context.Context, outputs []*structpb.Struct) (err error) {
			resJSON, err := protojson.Marshal(outputs[0])
			c.Assert(err, qt.IsNil)

			c.Check(resJSON, qt.JSONEquals, tc.wantResp)
			return nil
		})
		err = e.Execute(ctx, ir, ow)
		c.Assert(err, qt.IsNil)

	})
}

func TestComponent_ExecuteInsertMessageTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	tc := struct {
		name     string
		input    TaskInsertMessageInput
		wantResp string
	}{

		name: "ok - insert message",
		input: TaskInsertMessageInput{
			ThreadID:         "7509711154",
			SenderActorID:    "A-12345678",
			Recipients:       []string{"randomemail@gmail.com"},
			ChannelAccountID: "123456789",
			Subject:          "A fake message",
			Text:             "A message with random content inside",
		},
		wantResp: "SENT",
	}

	c.Run(tc.name, func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"token": bearerToken,
		})
		c.Assert(err, qt.IsNil)

		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: taskInsertMessage},
			client:             createMockClient(),
		}
		e.execute = e.InsertMessage

		pbInput, err := base.ConvertToStructpb(tc.input)

		c.Assert(err, qt.IsNil)

		ir := mock.NewInputReaderMock(c)
		ow := mock.NewOutputWriterMock(c)
		ir.ReadMock.Return([]*structpb.Struct{pbInput}, nil)
		ow.WriteMock.Optional().Set(func(ctx context.Context, outputs []*structpb.Struct) (err error) {
			resString := outputs[0].Fields["status"].GetStringValue()
			c.Check(resString, qt.Equals, tc.wantResp)
			return nil
		})
		err = e.Execute(ctx, ir, ow)

		c.Assert(err, qt.IsNil)

	})

}
