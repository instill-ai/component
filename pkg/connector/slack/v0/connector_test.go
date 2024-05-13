package slack

import (
	"context"
	"encoding/json"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/pkg/base"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	apiKey = "testkey"
)

func TestConnector_ExecuteWriteTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Connector{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    UserInputWriteTask
		wantResp WriteTaskResp
		wantErr  string
	}{
		{
			name: "ok to write",
			input: UserInputWriteTask{
				ChannelName:     "test_channel",
				Message:         "I am unit test",
				IsPublicChannel: true,
			},
			wantResp: WriteTaskResp{
				Result: "succeed",
			},
		},
		{
			name: "fail to write",
			input: UserInputWriteTask{
				ChannelName:     "test_channel_1",
				Message:         "I am unit test",
				IsPublicChannel: true,
			},
			wantErr: `there is no match name in slack channel \[test_channel_1\]`,
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {

			connection, err := structpb.NewStruct(map[string]any{
				"api_key": apiKey,
			})
			c.Assert(err, qt.IsNil)

			// It will increase the modification range if we do DI to CreateExecution.
			// So, I replaced it with the code below.
			// exec, err := connector.CreateExecution(nil, connection, taskWriteMessage)
			// c.Assert(err, qt.IsNil)
			e := &execution{
				ConnectorExecution: base.ConnectorExecution{Connector: connector, SystemVariables: nil, Connection: connection, Task: taskWriteMessage},
				client:             &MockSlackClient{},
			}
			e.execute = e.sendMessage
			exec := &base.ExecutionWrapper{Execution: e}

			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			got, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})

			if tc.wantErr != "" {
				c.Assert(err, qt.ErrorMatches, tc.wantErr)
				return
			}

			wantJSON, err := json.Marshal(tc.wantResp)
			c.Assert(err, qt.IsNil)
			c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
		})
	}
}

func TestConnector_ExecuteReadTask(t *testing.T) {

	c := qt.New(t)
	ctx := context.Background()
	bc := base.Connector{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name     string
		input    UserInputReadTask
		wantResp ReadTaskResp
		wantErr  string
	}{
		{
			name: "ok to read",
			input: UserInputReadTask{
				ChannelName: "test_channel",
			},
			wantResp: ReadTaskResp{
				Conversations: []Conversation{
					{
						UserID:     "user123",
						Message:    "Hello, world!",
						StartDate:  "2024-05-08",
						LastDate:   "2024-05-08",
						TS:         "1715159446.644219",
						ReplyCount: 1,
						ThreadReplyMessage: []ThreadReplyMessage{
							{
								UserID:   "user456",
								Message:  "Hello, how are you",
								DateTime: "2024-05-08 10:10:49",
							},
						},
					},
				},
			},
		},
		{
			name: "fail to read",
			input: UserInputReadTask{
				ChannelName: "test_channel_1",
			},
			wantErr: `there is no match name in slack channel \[test_channel_1\]`,
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			connection, err := structpb.NewStruct(map[string]any{
				"api_key": apiKey,
			})
			c.Assert(err, qt.IsNil)

			// It will increase the modification range if we do DI to CreateExecution.
			// So, I replaced it with the code below.
			// exec, err := connector.CreateExecution(nil, connection, taskWriteMessage)
			// c.Assert(err, qt.IsNil)

			e := &execution{
				ConnectorExecution: base.ConnectorExecution{Connector: connector, SystemVariables: nil, Connection: connection, Task: taskReadMessage},
				client:             &MockSlackClient{},
			}
			e.execute = e.readMessage
			exec := &base.ExecutionWrapper{Execution: e}

			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			got, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})

			if tc.wantErr != "" {
				c.Assert(err, qt.ErrorMatches, tc.wantErr)
				return
			}

			wantJSON, err := json.Marshal(tc.wantResp)
			c.Assert(err, qt.IsNil)
			c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())

		})

	}

}
