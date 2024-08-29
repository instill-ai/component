package freshdesk

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/gojuno/minimock/v3"
	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/mock"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestComponent_ExecuteGetProductTask(t *testing.T) {
	mc := minimock.NewController(t)
	c := qt.New(t)
	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)
	ctx := context.Background()

	FreshdeskClientMock := NewFreshdeskInterfaceMock(mc)

	FreshdeskClientMock.GetProductMock.
		When(154000129735).
		Then(
			&TaskGetProductResponse{
				Name:         "Fake Product",
				Description:  "This is a fake product",
				PrimaryEmail: "randomemail@gmail.com",
				CreatedAt:    "2024-08-29T09:35:16Z",
				UpdatedAt:    "2024-08-29T09:35:16Z",
				Default:      true,
			}, nil)

	tc := struct {
		name       string
		input      TaskGetProductInput
		wantOutput TaskGetProductOutput
	}{
		name: "ok - task get product",
		input: TaskGetProductInput{
			ProductID: 154000129735,
		},
		wantOutput: TaskGetProductOutput{
			Name:         "Fake Product",
			Description:  "This is a fake product",
			PrimaryEmail: "randomemail@gmail.com",
			CreatedAt:    "2024-08-29 09:35:16 UTC",
			UpdatedAt:    "2024-08-29 09:35:16 UTC",
			Default:      true,
		},
	}

	c.Run(tc.name, func(c *qt.C) {
		setup, err := structpb.NewStruct(map[string]any{
			"api-key": apiKey,
			"domain":  domain,
		})
		c.Assert(err, qt.IsNil)

		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: connector, SystemVariables: nil, Setup: setup, Task: taskGetProduct},
			client:             FreshdeskClientMock,
		}
		e.execute = e.TaskGetProduct

		pbIn, err := base.ConvertToStructpb(tc.input)
		c.Assert(err, qt.IsNil)

		ir := mock.NewInputReaderMock(c)
		ow := mock.NewOutputWriterMock(c)
		ir.ReadMock.Return([]*structpb.Struct{pbIn}, nil)
		ow.WriteMock.Optional().Set(func(ctx context.Context, outputs []*structpb.Struct) (err error) {

			outJSON, err := protojson.Marshal(outputs[0])
			c.Assert(err, qt.IsNil)

			c.Check(outJSON, qt.JSONEquals, tc.wantOutput)
			return nil
		})

		err = e.Execute(ctx, ir, ow)

		c.Assert(err, qt.IsNil)

	})
}
