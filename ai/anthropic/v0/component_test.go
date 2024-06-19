package anthropic

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/util/httpclient"
	"github.com/instill-ai/x/errmsg"
)

const (
	apiKey  = "123"
	errResp = `
{
  "error": {
    "message": "Incorrect API key provided."
  }
}`
)

func TestComponent_Execute(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()

	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name        string
		task        string
		path        string
		contentType string
	}{
		{
			name:        "text generation",
			task:        textGenerationTask,
			path:        messagesPath,
			contentType: httpclient.MIMETypeJSON,
		},
	}

	// TODO we'll likely want to have a test function per task and test at
	// least OK, NOK. For now, only errors are tested in order to verify
	// end-user messages.
	for _, tc := range testcases {
		c.Run("nok - "+tc.name+" 401", func(c *qt.C) {
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.Check(r.Method, qt.Equals, http.MethodPost)
				c.Check(r.URL.Path, qt.Equals, tc.path)
				c.Check(r.Header.Get("X-Api-Key"), qt.Equals, apiKey)

				c.Check(r.Header.Get("Content-Type"), qt.Matches, tc.contentType)

				w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintln(w, errResp)
			})

			anthropicServer := httptest.NewServer(h)
			c.Cleanup(anthropicServer.Close)

			setup, err := structpb.NewStruct(map[string]any{
				"base-path": anthropicServer.URL,
				"api-key":   apiKey,
			})
			c.Assert(err, qt.IsNil)

			exec, err := connector.CreateExecution(nil, setup, tc.task)
			c.Assert(err, qt.IsNil)

			pbIn := new(structpb.Struct)
			_, err = exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})
			c.Check(err, qt.IsNotNil)

			want := "Anthropic responded with a 401 status code. Incorrect API key provided."
			c.Check(errmsg.Message(err), qt.Equals, want)
		})
	}
	c.Run("nok - unsupported task", func(c *qt.C) {
		task := "FOOBAR"

		_, err := connector.CreateExecution(nil, nil, task)
		c.Check(err, qt.ErrorMatches, "unsupported task")
	})
}

func TestComponent_Test(t *testing.T) {
	c := qt.New(t)

	bc := base.Component{Logger: zap.NewNop()}
	connector := Init(bc)

	c.Run("nok - error", func(c *qt.C) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Check(r.Method, qt.Equals, http.MethodGet)

			w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, errResp)
		})

		anthropicServer := httptest.NewServer(h)
		c.Cleanup(anthropicServer.Close)

		_, err := structpb.NewStruct(map[string]any{
			"base-path": anthropicServer.URL,
		})
		c.Assert(err, qt.IsNil)
	})

	c.Run("ok - disconnected", func(c *qt.C) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Check(r.Method, qt.Equals, http.MethodGet)

			w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
			fmt.Fprintln(w, `{}`)
		})

		openAIServer := httptest.NewServer(h)
		c.Cleanup(openAIServer.Close)

		_, err := structpb.NewStruct(map[string]any{
			"base_path": openAIServer.URL,
		})
		c.Assert(err, qt.IsNil)
	})

	c.Run("ok - connected", func(c *qt.C) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Check(r.Method, qt.Equals, http.MethodGet)

			w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
			fmt.Fprintln(w, `{"data": [{}]}`)
		})

		anthropicServer := httptest.NewServer(h)
		c.Cleanup(anthropicServer.Close)

		setup, err := structpb.NewStruct(map[string]any{
			"base_path": anthropicServer.URL,
		})
		c.Assert(err, qt.IsNil)

		err = connector.Test(nil, setup)
		c.Check(err, qt.IsNil)
	})
}
