package slack

import (
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

	logger := zap.NewNop()
	connector := Init(logger, nil)

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
				BaseConnectorExecution: base.BaseConnectorExecution{Connector: connector, SystemVariables: nil, Connection: connection, Task: taskWriteMessage},
				client:                 &MockSlackClient{},
			}
			e.execute = e.sendMessage
			exec := &base.ExecutionWrapper{Execution: e}

			pbIn, err := base.ConvertToStructpb(tc.input)
			c.Assert(err, qt.IsNil)

			got, err := exec.Execution.Execute([]*structpb.Struct{pbIn})

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

