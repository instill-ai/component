package hubspot

import (
	"context"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/mock"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

// mockClient is in contact_test.go

// Mock GetAll struct and its functions
type MockGetAll struct{}

func (s *MockGetAll) Get(objectType string, param string) (*TaskGetAllResp, error) {
	var resource *TaskGetAllResp
	if param == "" {
		resource = &TaskGetAllResp{
			Results: []taskGetAllRespResult{
				{
					ID: "11111111111",
				},
				{
					ID: "22222222222",
				},
			},
			Paging: &taskGetAllRespPaging{
				Next: struct {
					Link  string `json:"link"`
					After string `json:"after"`
				}{
					Link:  "https://api.hubapi.com/crm/v3/objects/contacts?after=xxxxxxxxx",
					After: "xxxxxxxxxxx",
				},
			},
		}
	} else if strings.Contains(param, "after") {
		resource = &TaskGetAllResp{
			Results: []taskGetAllRespResult{
				{
					ID: "33333333333",
				},
				{
					ID: "44444444444",
				},
			},
		}
	}

	return resource, nil
}

func TestComponent_ExecuteGetAllTask(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	tc := struct {
		name     string
		input    string
		wantResp TaskGetAllOutput
	}{
		name:  "ok - get all contacts",
		input: "Contacts",
		wantResp: TaskGetAllOutput{
			ObjectIDs:       []string{"11111111111", "22222222222", "33333333333", "44444444444"},
			ObjectIDsLength: 4,
		},
	}

	c.Run(tc.name, func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"token": bearerToken,
		})
		c.Assert(err, qt.IsNil)

		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: taskGetAll},
			client:             createMockClient(),
		}
		e.execute = e.GetAll

		pbInput, err := structpb.NewStruct(map[string]any{
			"object-type": tc.input,
		})

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

		c.Assert(err, qt.IsNil)

		err = e.Execute(ctx, ir, ow)
		c.Assert(err, qt.IsNil)

	})
}
