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

	got := x.Execution.GetComponent()
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

			pbin, err := structpb.NewStruct(tc.in)
			c.Assert(err, qt.IsNil)

			got, err := x.Execute(ctx, []*structpb.Struct{pbin})
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

/*
// TODO check usage handler - check error
// TODO check usage handler - collect error
// TODO check usage handler - no error
func TestExecutionWrapper_Execute(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()

	cmp := &testComp{Component{
		// NewUsageHandler: creator.newUH,
	}}
	err := cmp.LoadDefinition(
		connectorDefJSON,
		connectorConfigJSON,
		connectorTasksJSON,
		map[string][]byte{"additional.json": connectorAdditionalJSON})
		c.Assert(err, qt.IsNil)

	want := TextCompletionOutput{
		Texts: []string{"hello!"},
		Usage: usage{TotalTokens: 25},
	}
	resp := `{"usage": {"total_tokens": 25}, "choices": [{"message": {"content": "hello!"}}]}`

	pbIn, err := base.ConvertToStructpb(TextCompletionInput{
		Model:  "gpt-3.5-turbo",
		Prompt: "what instrument did Yusef Lateef play?",
		Images: []string{},
	})
	c.Assert(err, qt.IsNil)
	inputs := []*structpb.Struct{pbIn}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Check(r.Header.Get("Authorization"), qt.Equals, "Bearer "+apiKey)

		w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
		fmt.Fprintln(w, resp)
	})

	task := TextGenerationTask
	bc := new(base.Component)

	/*
	c.Run("nok - usage handler check error", func(c *qt.C) {
		c.Cleanup(cleanupConn)

		uh := mock.NewUsageHandlerMock(c)
		uh.CheckMock.When(minimock.AnyContext, inputs).Then(fmt.Errorf("check error"))
		creator := usageHandlerCreator{uh}

		bc := base.Component{NewUsageHandler: creator.newUH}
		connector := Init(bc)

		setup, err := structpb.NewStruct(map[string]any{})
		c.Assert(err, qt.IsNil)

		exec, err := connector.CreateExecution(nil, setup, task)
		c.Assert(err, qt.IsNil)

		_, err = exec.Execute(ctx, inputs)
		c.Check(err, qt.IsNotNil)
		c.Check(err, qt.ErrorMatches, "check error")
	})

	c.Run("nok - usage handler collect error", func(c *qt.C) {
		c.Cleanup(cleanupConn)

		uh := mock.NewUsageHandlerMock(c)
		uh.CheckMock.When(minimock.AnyContext, inputs).Then(nil)
		uh.CollectMock.When(minimock.AnyContext, inputs, outputs).Then(fmt.Errorf("collect error"))
		creator := usageHandlerCreator{uh}

		bc := base.Component{NewUsageHandler: creator.newUH}
		connector := Init(bc)

		setup, err := structpb.NewStruct(map[string]any{
			"base-path": openAIServer.URL,
			"api-key":   apiKey,
		})
		c.Assert(err, qt.IsNil)

		exec, err := connector.CreateExecution(nil, setup, task)
		c.Assert(err, qt.IsNil)

		_, err = exec.Execute(ctx, inputs)
		c.Check(err, qt.IsNotNil)
		c.Check(err, qt.ErrorMatches, "collect error")
	})

	c.Run("ok - with usage handler", func(c *qt.C) {
		c.Cleanup(cleanupConn)

		uh := mock.NewUsageHandlerMock(c)
		uh.CheckMock.Return(nil)   //.When(minimock.AnyContext, minimock.Any).Then(nil)
		uh.CollectMock.Return(nil) //.When(minimock.AnyContext, inputs, minimock.Any).Then(nil)
		creator := usageHandlerCreator{uh}

		bc.NewUsageHandler = creator.newUH

		c.Cleanup(func() { bc.NewUsageHandler = nil })

		connector := Init(bc)

		setup, err := structpb.NewStruct(map[string]any{
			"base-path": openAIServer.URL,
			"api-key":   apiKey,
		})
		c.Assert(err, qt.IsNil)

		exec, err := connector.CreateExecution(nil, setup, task)
		c.Assert(err, qt.IsNil)
		c.Check(exec.Execution.UsesSecret(), qt.IsFalse)

		got, err := exec.Execute(ctx, inputs)
		c.Check(err, qt.IsNil)
		c.Assert(got, qt.HasLen, 1)

		gotJSON, err := got[0].MarshalJSON()
		c.Assert(err, qt.IsNil)
		c.Check(gotJSON, qt.JSONEquals, want)
	})


	c.Run("ok - with secrets & usage handler", func(c *qt.C) {
		c.Cleanup(cleanupConn)

		secrets := map[string]any{"apikey": apiKey}
		connector := Init(bc).WithSecrets(secrets)

		setup, err := structpb.NewStruct(map[string]any{
			"base-path": openAIServer.URL,
			"api-key":   "__INSTILL_SECRET", // will be replaced by secrets.apikey
		})
		c.Assert(err, qt.IsNil)

		exec, err := connector.CreateExecution(nil, setup, task)
		c.Assert(err, qt.IsNil)
		c.Check(exec.Execution.UsesSecret(), qt.IsTrue)

		got, err := exec.Execute(ctx, inputs)
		c.Check(err, qt.IsNil)
		c.Assert(got, qt.HasLen, 1)

		gotJSON, err := got[0].MarshalJSON()
		c.Assert(err, qt.IsNil)
		c.Check(gotJSON, qt.JSONEquals, want)
	})

	c.Run("nok - secret not injected", func(c *qt.C) {
		c.Cleanup(cleanupConn)

		connector := Init(bc)
		setup, err := structpb.NewStruct(map[string]any{
			"api-key": "__INSTILL_SECRET",
		})
		c.Assert(err, qt.IsNil)

		_, err = connector.CreateExecution(nil, setup, task)
		c.Check(err, qt.IsNotNil)
		c.Check(err, qt.ErrorMatches, "unresolved global secret")
		c.Check(errmsg.Message(err), qt.Equals, "The configuration field api-key can't reference a global secret.")
	})
}
*/

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

func (c *testComp) CreateExecution(_ map[string]any, _ *structpb.Struct, task string) (*ExecutionWrapper, error) {
	x := ComponentExecution{Component: c, Task: task}
	return &ExecutionWrapper{Execution: &testExec{
		ComponentExecution: x,

		out: c.xOut,
		err: c.xErr,
	}}, nil
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
