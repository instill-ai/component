package hubspot

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	hubspot "github.com/belong-inc/go-hubspot"
	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/mock"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

// mockClient is in contact_test.go

// Mock Owner struct and its functions
type MockOwner struct{}

func (s *MockOwner) Get(ownerInfo string, infoType string) (*TaskGetOwnerResp, error) {

	var fakeOwner *TaskGetOwnerResp
	if infoType == "id" || infoType == "userId" {
		if ownerInfo == "1111111111" || ownerInfo == "22222222" {
			fakeOwner = &TaskGetOwnerResp{
				FirstName: "Random",
				LastName:  "Human",
				Email:     "randomhuman@gmail.com",
				OwnerID:   "1111111111",
				UserID:    22222222,
				CreatedAt: hubspot.NewTime(time.Date(2024, 7, 9, 0, 0, 0, 0, time.UTC)),
				UpdatedAt: hubspot.NewTime(time.Date(2024, 7, 9, 0, 0, 0, 0, time.UTC)),
				Archived:  false,
			}
		} else { //if the owner id/userId is not found

			// actual response from API if the owner id/userId is not found
			resp := `
			<html>
			<head>
				<meta http-equiv="Content-Type" content="text/html;charset=utf-8" />
				<title>Error 404 Not Found</title>
			</head>

			<body>
				<h2>HTTP ERROR 404</h2>
				<p>Resource not found</p>
			</body>

			</html>
			`

			err := json.Unmarshal([]byte(resp), &fakeOwner)

			// go-hubspot sdk will return in this format
			return nil, fmt.Errorf("404: unable to read response from hubspot:%v", err)

		}
	}

	return fakeOwner, nil
}

func TestComponent_ExecuteGetOwnerTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name      string
		inputID   string
		inputType string
		wantResp  TaskGetOwnerOutput
		wantErr   string
	}{
		{
			name:      "ok - get owner",
			inputID:   "1111111111",
			inputType: "Owner ID",
			wantResp: TaskGetOwnerOutput{
				FirstName: "Random",
				LastName:  "Human",
				Email:     "randomhuman@gmail.com",
				OwnerID:   "1111111111",
				UserID:    "22222222",
				CreatedAt: "2024-07-09 00:00:00 +0000 UTC",
				UpdatedAt: "2024-07-09 00:00:00 +0000 UTC",
				Archived:  false,
			},
			wantErr: "",
		},
		{
			name:      "nok - get owner: owner not found",
			inputID:   "9999999999",
			inputType: "Owner ID",
			wantErr:   "404: unable to read response from hubspot: no owner was found",
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			setup, err := structpb.NewStruct(map[string]any{
				"token": bearerToken,
			})
			c.Assert(err, qt.IsNil)

			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: taskGetOwner},
				client:             createMockClient(),
			}
			e.execute = e.GetOwner

			pbInput, err := structpb.NewStruct(map[string]any{
				"id":      tc.inputID,
				"id-type": tc.inputType,
			})

			c.Assert(err, qt.IsNil)

			input := []*structpb.Struct{pbInput}

			ir := mock.NewInputReaderMock(c)
			ir.ReadMock.Return(input, nil)
			ow := mock.NewOutputWriterMock(c)
			ow.WriteMock.Optional().Set(func(ctx context.Context, outputs []*structpb.Struct) (err error) {
				if tc.name == "error case" {
					c.Assert(outputs, qt.IsNil)
					return
				}
				c.Assert(outputs, qt.HasLen, 1)
				res := outputs[0]
				resJSON, err := protojson.Marshal(res)
				c.Check(resJSON, qt.JSONEquals, tc.wantResp)
				c.Assert(err, qt.IsNil)
				return nil
			})

			err = e.Execute(ctx, ir, ow)
			if tc.wantErr != "" {
				c.Assert(err, qt.ErrorMatches, tc.wantErr)
				return
			}

			c.Assert(err, qt.IsNil)

		})
	}
}
