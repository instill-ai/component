package document

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"code.sajari.com/docconv"
	qt "github.com/frankban/quicktest"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
)

func TestOperator(t *testing.T) {
	c := qt.New(t)

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
	}
	bc := base.Component{}
	ctx := context.Background()
	for i := range testcases {
		tc := &testcases[i]
		c.Run(tc.name, func(c *qt.C) {
			component := Init(bc)
			c.Assert(component, qt.IsNotNil)

			execution, err := component.CreateExecution(base.ComponentExecution{
				Component: component,
				Task:      tc.task,
			})
			c.Assert(err, qt.IsNil)
			c.Assert(execution, qt.IsNotNil)

			input := []*structpb.Struct{&tc.input}

			outputs, err := execution.Execute(ctx, input)

			c.Assert(err, qt.IsNil)
			c.Assert(outputs, qt.HasLen, 1)
		})
	}
}