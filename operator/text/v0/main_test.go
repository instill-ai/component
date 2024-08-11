package text

import (
	"context"
	"testing"

	"github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestOperator(t *testing.T) {
	c := quicktest.New(t)

	testcases := []struct {
		name  string
		task  string
		input structpb.Struct
	}{
		{
			name: "chunk texts",
			task: "TASK_CHUNK_TEXT",
			input: structpb.Struct{
				Fields: map[string]*structpb.Value{
					"text": {Kind: &structpb.Value_StringValue{StringValue: "Hello world. This is a test."}},
					"strategy": {Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"setting": {Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"chunk-method": {Kind: &structpb.Value_StringValue{StringValue: "Token"}},
								},
							}}},
						},
					}}},
					"tokenization": {Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"choice": {Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"tokenization-method": {Kind: &structpb.Value_StringValue{StringValue: "Model"}},
									"model":               {Kind: &structpb.Value_StringValue{StringValue: "gpt-3.5-turbo"}},
								},
							}}},
						},
					}}},
				},
			},
		},
		{
			name:  "error case",
			task:  "FAKE_TASK",
			input: structpb.Struct{},
		},
	}
	bc := base.Component{}
	ctx := context.Background()
	for i := range testcases {
		tc := &testcases[i]
		c.Run(tc.name, func(c *quicktest.C) {
			component := Init(bc)
			c.Assert(component, quicktest.IsNotNil)

			execution, err := component.CreateExecution(base.ComponentExecution{
				Component: component,
				Task:      tc.task,
			})
			if tc.name == "error case" {
				c.Assert(err, quicktest.ErrorMatches, "FAKE_TASK task is not supported.")
				return
			}
			c.Assert(err, quicktest.IsNil)
			c.Assert(execution, quicktest.IsNotNil)

			input := []*structpb.Struct{&tc.input}

			outputs, err := execution.Execute(ctx, input)

			c.Assert(err, quicktest.IsNil)
			c.Assert(outputs, quicktest.HasLen, 1)
		})
	}
}
