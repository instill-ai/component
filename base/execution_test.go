package base

import (
	"context"
	"fmt"
	"testing"

	_ "embed"

	"google.golang.org/protobuf/types/known/structpb"

	qt "github.com/frankban/quicktest"
)

func TestComponentExecution_GetComponent(t *testing.T) {
	c := qt.New(t)

	cmp := &testComp{
		Component: Component{
			NewUsageHandler: usageHandlerCreator(nil, nil),
		},
	}
	err := cmp.LoadDefinition(
		connectorDefJSON,
		connectorConfigJSON,
		connectorTasksJSON,
		map[string][]byte{"additional.json": connectorAdditionalJSON})
	c.Assert(err, qt.IsNil)

	x, err := cmp.CreateExecution(nil, nil, "TASK_TEXT_EMBEDDINGS")
	c.Assert(err, qt.IsNil)

	got := x.GetComponent()
	c.Check(got.GetDefinitionID(), qt.Equals, "openai")
}

func TestExecutionWrapper_Execute(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()

	inputValid := map[string]any{
		"text":  "What's Horace Andy's biggest hit?",
		"model": "text-embedding-ada-002",
	}
	outputValid := map[string]any{"embedding": []any{0.001}}

	testcases := []struct {
		name       string
		in         map[string]any
		checkErr   error
		collectErr error
		out        map[string]any
		outErr     error
		want       map[string]any
		wantErr    string
	}{
		{
			name:    "nok - invalid input",
			in:      map[string]any{"text": "What's Horace Andy's biggest hit?"},
			wantErr: `inputs\[0\]: missing properties: 'model'`,
		},
		{
			name:     "nok - check error",
			in:       inputValid,
			checkErr: fmt.Errorf("foo"),
			wantErr:  "foo",
		},
		{
			name:    "nok - invalid output",
			in:      inputValid,
			out:     map[string]any{"response": "Sky Larking, definitely!"},
			wantErr: `outputs\[0\]: missing properties: 'embedding'`,
		},
		{
			name:    "nok - execution error",
			in:      inputValid,
			outErr:  fmt.Errorf("bar"),
			wantErr: "bar",
		},
		{
			name:       "nok - collect error",
			in:         inputValid,
			out:        outputValid,
			collectErr: fmt.Errorf("zot"),
			wantErr:    "zot",
		},
		{
			name: "ok",
			in:   inputValid,
			out:  outputValid,
			want: outputValid,
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			cmp := &testComp{
				Component: Component{
					NewUsageHandler: usageHandlerCreator(tc.checkErr, tc.collectErr),
				},
				xOut: []map[string]any{tc.out},
				xErr: tc.outErr,
			}

			err := cmp.LoadDefinition(
				connectorDefJSON,
				connectorConfigJSON,
				connectorTasksJSON,
				map[string][]byte{"additional.json": connectorAdditionalJSON})
			c.Assert(err, qt.IsNil)

			x, err := cmp.CreateExecution(nil, nil, "TASK_TEXT_EMBEDDINGS")
			c.Assert(err, qt.IsNil)

			xw := &ExecutionWrapper{Execution: x}

			pbin, err := structpb.NewStruct(tc.in)
			c.Assert(err, qt.IsNil)

			got, err := xw.Execute(ctx, []*structpb.Struct{pbin})
			if tc.wantErr != "" {
				c.Check(err, qt.IsNotNil)
				c.Check(err, qt.ErrorMatches, tc.wantErr)
				return
			}

			c.Check(err, qt.IsNil)
			c.Assert(got, qt.HasLen, 1)

			gotJSON, err := got[0].MarshalJSON()
			c.Assert(err, qt.IsNil)
			c.Check(gotJSON, qt.JSONEquals, tc.want)
		})
	}
}

type testExec struct {
	ComponentExecution

	out []map[string]any
	err error
}

func (e *testExec) Execute(_ context.Context, _ []*structpb.Struct) ([]*structpb.Struct, error) {
	if e.out == nil {
		return nil, e.err
	}

	pbout := make([]*structpb.Struct, len(e.out))
	for i, o := range e.out {
		pbo, err := structpb.NewStruct(o)
		if err != nil {
			panic(err)
		}
		pbout[i] = pbo
	}

	return pbout, e.err
}

type testComp struct {
	Component

	// execution output
	xOut []map[string]any
	xErr error
}

func (c *testComp) CreateExecution(_ map[string]any, _ *structpb.Struct, task string) (IExecution, error) {
	x := ComponentExecution{Component: c, Task: task}
	return &testExec{
		ComponentExecution: x,

		out: c.xOut,
		err: c.xErr,
	}, nil
}

func usageHandlerCreator(checkErr, collectErr error) UsageHandlerCreator {
	return func(IExecution) (UsageHandler, error) {
		return &usageHandler{
			checkErr:   checkErr,
			collectErr: collectErr,
		}, nil
	}
}

type usageHandler struct {
	checkErr   error
	collectErr error
}

func (h *usageHandler) Check(context.Context, []*structpb.Struct) error          { return h.checkErr }
func (h *usageHandler) Collect(_ context.Context, _, _ []*structpb.Struct) error { return h.collectErr }
