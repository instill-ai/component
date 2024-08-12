package ai21labs

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/util/httpclient"
	"github.com/instill-ai/x/errmsg"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	correctAPIKey = "### CORRECT API KEY ###"
	wrongAPIKey   = "### WRONG API KEY ###"
	errResp       = `
{
  "error": {
    "message": "Incorrect API key provided."
  }
}`
)

func TestComponent_Execute(t *testing.T) {
	c := qt.New(t)

	bc := base.Component{Logger: zap.NewNop()}
	cmp := Init(bc)

	supportedTasks := []string{
		"TASK_TEXT_GENERATION_CHAT",
		"TASK_TEXT_EMBEDDINGS",
		"TASK_CONTEXTUAL_ANSWERING",
		"TASK_TEXT_SUMMARIZATION",
		"TASK_TEXT_SUMMARIZATION_SEGMENT",
		"TASK_TEXT_PARAPHRASING",
		"TASK_GRAMMAR_CHECK",
		"TASK_TEXT_IMPROVEMENT",
		"TASK_TEXT_SEGMENTATION",
	}
	for _, supportedTask := range supportedTasks {
		c.Run("ok - supported task", func(c *qt.C) {
			task := supportedTask

			_, err := cmp.CreateExecution(base.ComponentExecution{
				Component: cmp,
				Task:      task,
			})
			c.Check(err, qt.IsNil)
		})
	}

	c.Run("nok - unsupported task", func(c *qt.C) {
		task := "FOOBAR"
		_, err := cmp.CreateExecution(base.ComponentExecution{
			Component: cmp,
			Task:      task,
		})
		c.Check(err, qt.ErrorMatches, "unsupported task")
	})
}

func TestComponent_Connection(t *testing.T) {
	c := qt.New(t)

	bc := base.Component{Logger: zap.NewNop()}
	cmp := Init(bc)

	ctx := context.Background()

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Check(r.Method, qt.Equals, http.MethodPost)

		w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
		if r.Header.Get("Authorization") != "Bearer "+correctAPIKey {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, errResp)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
			fmt.Fprintln(w, `{
				"id": "5c307fd0-aaf9-44e5-a72c-6fb97d67395f",
				"choices":[{
					"index": 0,
					"message": {
						"role": "assistant",
						"content": "Hello, how can I help you today?"
					},
					"finish_reason": "stop"
				}],
				"usage": {
					"prompt_tokens": 0,
					"completion_tokens": 0,
					"total_tokens": 0
				}
			}`)
		}

	})

	c.Run("ok - correct API key", func(c *qt.C) {
		ai21labsServer := httptest.NewServer(h)
		c.Cleanup(ai21labsServer.Close)

		setup, err := structpb.NewStruct(map[string]any{
			"base-path": ai21labsServer.URL,
			"api-key":   correctAPIKey,
		})
		c.Assert(err, qt.IsNil)

		exec, err := cmp.CreateExecution(base.ComponentExecution{
			Component: cmp,
			Task:      "TASK_TEXT_GENERATION_CHAT",
			Setup:     setup,
		})
		c.Assert(err, qt.IsNil)
		pbIn := new(structpb.Struct)
		_, err = exec.Execute(ctx, []*structpb.Struct{pbIn})
		c.Check(err, qt.IsNil)

	})

	c.Run("nok - incorrect API key", func(c *qt.C) {
		ai21labsServer := httptest.NewServer(h)
		c.Cleanup(ai21labsServer.Close)

		setup, err := structpb.NewStruct(map[string]any{
			"base-path": ai21labsServer.URL,
			"api-key":   wrongAPIKey,
		})
		c.Assert(err, qt.IsNil)

		exec, err := cmp.CreateExecution(base.ComponentExecution{
			Component: cmp,
			Task:      "TASK_TEXT_GENERATION_CHAT",
			Setup:     setup,
		})
		c.Assert(err, qt.IsNil)
		pbIn := new(structpb.Struct)
		_, err = exec.Execute(ctx, []*structpb.Struct{pbIn})
		c.Check(err, qt.IsNotNil)

		want := "AI21labs responded with a 401 status code. Incorrect API key provided."
		c.Check(errmsg.Message(err), qt.Equals, want)

	})

	c.Run("nok - disconnected", func(c *qt.C) {
		ai21labsServer := httptest.NewServer(h)
		c.Cleanup(ai21labsServer.Close)

		setup, err := structpb.NewStruct(map[string]any{
			"base-path": ai21labsServer.URL,
		})
		c.Assert(err, qt.IsNil)
		err = cmp.TestConnection(nil, setup)
		c.Check(err, qt.IsNotNil)
	})
}

func (c *component) TestConnection(_ map[string]any, setup *structpb.Struct) error {
	clt := newClient(getAPIKey(setup), getBasePath(setup), c.GetLogger())
	req := clt.httpClient.R().SetResult(&struct{}{})

	if _, err := req.Post("/"); err != nil {
		return err
	}

	return nil
}
