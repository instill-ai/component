package text

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"code.sajari.com/docconv"
	"github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestConnector(t *testing.T) {
	c := quicktest.New(t)

	fileContent, _ := os.ReadFile("testdata/test.txt")
	base64DataURI := fmt.Sprintf("data:%s;base64,%s", docconv.MimeTypeByExtension("testdata/test.txt"), base64.StdEncoding.EncodeToString(fileContent))
	testcases := []struct {
		name  string
		task  string
		input structpb.Struct
	}{
		{
			name: "convert to text",
			task: "TASK_CONVERT_TO_TEXT",
			input: structpb.Struct{
				Fields: map[string]*structpb.Value{
					"doc": {Kind: &structpb.Value_StringValue{StringValue: base64DataURI}},
				},
			},
		},
		{
			name: "split by token",
			task: "TASK_SPLIT_BY_TOKEN",
			input: structpb.Struct{
				Fields: map[string]*structpb.Value{
					"text":  {Kind: &structpb.Value_StringValue{StringValue: "Hello world. This is a test."}},
					"model": {Kind: &structpb.Value_StringValue{StringValue: "gpt-3.5-turbo"}},
				},
			},
		},
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
									"chunk_method": {Kind: &structpb.Value_StringValue{StringValue: "Token"}},
								},
							}}},
						},
					}}},
				},
			},
		},
	}
	bc := base.Operator{Logger: zap.NewNop()}
	ctx := context.Background()
	for i := range testcases {
		tc := &testcases[i]
		c.Run(tc.name, func(c *quicktest.C) {
			connector := Init(bc)
			execution, err := connector.CreateExecution(map[string]any{}, tc.task)
			c.Assert(err, quicktest.IsNil)

			input := []*structpb.Struct{&tc.input}
			outputs, err := execution.Execute(ctx, input)
			
			c.Assert(err, quicktest.IsNil)
			c.Assert(outputs, quicktest.HasLen, 1)
		})
	}
}
