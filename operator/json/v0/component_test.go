package json

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
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
	ctx := context.Background()

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
			name: "ok - marshal object",

			task:     taskMarshal,
			in:       map[string]any{"json": asMap},
			wantJSON: json.RawMessage(asJSON),
		},
		{
			name: "ok - marshal string",

			task:     taskMarshal,
			in:       map[string]any{"json": "dos"},
			wantJSON: json.RawMessage(`"dos"`),
		},
		{
			name: "ok - marshal array",

			task:     taskMarshal,
			in:       map[string]any{"json": []any{false, true, "dos", 3}},
			wantJSON: json.RawMessage(`[false, true, "dos", 3]`),
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
			name: "ok - unmarshal array",

			task: taskUnmarshal,
			in:   map[string]any{"string": `[false, true, "dos", 3]`},
			want: map[string]any{"json": []any{false, true, "dos", 3}},
		},
		{
			name: "ok - unmarshal string",

			task: taskUnmarshal,
			in:   map[string]any{"string": `"foo"`},
			want: map[string]any{"json": "foo"},
		},
		{
			name: "nok - unmarshal",

			task:    taskUnmarshal,
			in:      map[string]any{"string": `{`},
			wantErr: "Couldn't parse the JSON string. Please check the syntax is correct.",
		},
		{
			name: "ok - jq from string",

			task: taskJQ,
			in: map[string]any{
				"json-string": `{"a": {"b": 42}}`,
				"jq-filter":   ".a | .[]",
			},
			want: map[string]any{
				"results": []any{42},
			},
		},
		{
			name: "nok - jq invalid JSON string",

			task: taskJQ,
			in: map[string]any{
				"json-string": "{",
				"jq-filter":   ".",
			},
			wantErr: "Couldn't parse the JSON input. Please check the syntax is correct.",
		},
		{
			name: "ok - string value",

			task: taskJQ,
			in: map[string]any{
				"json-value": "foo",
				"jq-filter":  `. + "bar"`,
			},
			want: map[string]any{
				"results": []any{"foobar"},
			},
		},
		{
			name: "ok - from array",

			task: taskJQ,
			in: map[string]any{
				"json-value": []any{2, 3, 23},
				"jq-filter":  ".[2]",
			},
			want: map[string]any{
				"results": []any{23},
			},
		},
		{
			name: "ok - jq create object",

			task: taskJQ,
			in: map[string]any{
				"json-value": map[string]any{
					"id": "sample",
					"10": map[string]any{"b": 42},
				},
				"jq-filter": `{(.id): .["10"].b}`,
			},
			want: map[string]any{
				"results": []any{
					map[string]any{"sample": 42},
				},
			},
		},
		{
			name: "nok - jq invalid filter",

			task: taskJQ,
			in: map[string]any{
				"json-string": asJSON,
				"jq-filter":   ".foo & .bar",
			},
			wantErr: `Couldn't parse the jq filter: unexpected token "&". Please check the syntax is correct.`,
		},
	}

	bo := base.Component{}
	cmp := Init(bo)

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			exec, err := cmp.CreateExecution(base.ComponentExecution{
				Component: cmp,
				Task:      tc.task,
			})
			c.Assert(err, qt.IsNil)

			pbIn, err := structpb.NewStruct(tc.in)
			c.Assert(err, qt.IsNil)

			ir, ow, eh, job := base.GenerateMockJob(c)
			ir.ReadMock.Return(pbIn, nil)
			ow.WriteMock.Optional().Set(func(ctx context.Context, output *structpb.Struct) (err error) {

				if tc.wantJSON != nil {
					// Check JSON in the output string.
					b := output.Fields["string"].GetStringValue()
					c.Check([]byte(b), qt.JSONEquals, tc.wantJSON)
					return
				}

				gotJSON, err := output.MarshalJSON()
				c.Assert(err, qt.IsNil)
				c.Check(gotJSON, qt.JSONEquals, tc.want)
				return nil
			})
			eh.ErrorMock.Optional().Set(func(ctx context.Context, err error) {
				if tc.wantErr != "" {
					c.Check(errmsg.Message(err), qt.Matches, tc.wantErr)
				}
			})

			err = exec.Execute(ctx, []*base.Job{job})
			c.Check(err, qt.IsNil)

		})
	}
}

func TestOperator_CreateExecution(t *testing.T) {
	c := qt.New(t)

	bc := base.Component{}
	cmp := Init(bc)

	c.Run("nok - unsupported task", func(c *qt.C) {
		task := "FOOBAR"
		want := fmt.Sprintf("%s task is not supported.", task)

		_, err := cmp.CreateExecution(base.ComponentExecution{
			Component: cmp,
			Task:      task,
		})
		c.Check(err, qt.IsNotNil)
		c.Check(errmsg.Message(err), qt.Equals, want)
	})
}
