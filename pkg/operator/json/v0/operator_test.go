package json

import (
	"encoding/json"
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/x/errmsg"
)

const asJSON = `
{
  "a": "27",
  "b": 27
}`

var asMap = map[string]any{"a": "27", "b": 27}

func TestOperator_Execute(t *testing.T) {
	c := qt.New(t)

	testcases := []struct {
		name string

		task    string
		in      map[string]any
		want    map[string]any
		wantErr string

		// The marshal task will return a string with a valid JSON in the
		// output. However, the format of the JSON may vary (e.g. spaces), so
		// this field will be used to do a JSON comparison instead of a value
		// one.
		wantJSON json.RawMessage
	}{
		{
			name: "ok - marshal",

			task:     taskMarshal,
			in:       map[string]any{"json": asMap},
			wantJSON: json.RawMessage(asJSON),
		},
		{
			name: "nok - marshal",

			task:    taskMarshal,
			in:      map[string]any{},
			wantErr: "Couldn't convert the provided object to JSON.",
		},
		{
			name: "ok - unmarshal",

			task: taskUnmarshal,
			in:   map[string]any{"string": asJSON},
			want: map[string]any{"json": asMap},
		},
		{
			name: "nok - unmarshal",

			task:    taskUnmarshal,
			in:      map[string]any{"string": `{`},
			wantErr: "Couldn't parse the JSON string. Please check the syntax is correct.",
		},
		{
			name: "ok - jq",

			task: taskJQ,
			in: map[string]any{
				"jsonInput": `{"a": {"b": 42}}`,
				"jqFilter":  ".a | .[]",
			},
			want: map[string]any{
				"results": []any{42},
			},
		},
		{
			name: "ok - jq create object",

			task: taskJQ,
			in: map[string]any{
				"jsonInput": `{"id": "sample", "10": {"b": 42}}`,
				"jqFilter":  `{(.id): .["10"].b}`,
			},
			want: map[string]any{
				"results": []any{
					map[string]any{"sample": 42},
				},
			},
		},
		{
			name: "nok - jq invalid JSON input",

			task: taskJQ,
			in: map[string]any{
				"jsonInput": "{",
				"jqFilter":  ".",
			},
			wantErr: "Couldn't parse the JSON input. Please check the syntax is correct.",
		},
		{
			name: "nok - jq invalid filter",

			task: taskJQ,
			in: map[string]any{
				"jsonInput": asJSON,
				"jqFilter":  ".foo & .bar",
			},
			wantErr: `Couldn't parse the jq filter: unexpected token "&". Please check the syntax is correct.`,
		},
	}

	logger := zap.NewNop()
	operator := Init(logger, nil)

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			exec, err := operator.CreateExecution(nil, tc.task)
			c.Assert(err, qt.IsNil)

			pbIn, err := structpb.NewStruct(tc.in)
			c.Assert(err, qt.IsNil)

			got, err := exec.Execution.Execute([]*structpb.Struct{pbIn})
			if tc.wantErr != "" {
				c.Check(errmsg.Message(err), qt.Matches, tc.wantErr)
				return
			}

			c.Check(err, qt.IsNil)
			c.Assert(got, qt.HasLen, 1)

			if tc.wantJSON != nil {
				// Check JSON in the output string.
				b := got[0].Fields["string"].GetStringValue()
				c.Check([]byte(b), qt.JSONEquals, tc.wantJSON)
				return
			}

			gotJSON, err := got[0].MarshalJSON()
			c.Assert(err, qt.IsNil)
			c.Check(gotJSON, qt.JSONEquals, tc.want)
		})
	}
}

func TestOperator_CreateExecution(t *testing.T) {
	c := qt.New(t)

	logger := zap.NewNop()
	operator := Init(logger, nil)

	c.Run("nok - unsupported task", func(c *qt.C) {
		task := "FOOBAR"
		want := fmt.Sprintf("%s task is not supported.", task)

		_, err := operator.CreateExecution(nil, task)
		c.Check(err, qt.IsNotNil)
		c.Check(errmsg.Message(err), qt.Equals, want)
	})
}
