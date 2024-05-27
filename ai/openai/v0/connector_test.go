package openai

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/gojuno/minimock/v3"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/mock"
	"github.com/instill-ai/component/internal/util/httpclient"
	"github.com/instill-ai/x/errmsg"
)

const (
	apiKey  = "123"
	org     = "org1"
	errResp = `
{
  "error": {
    "message": "Incorrect API key provided."
  }
}`
)

func TestConnector_Execute(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()

	bc := base.Connector{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name        string
		task        string
		path        string
		contentType string
	}{
		{
			name:        "text generation",
			task:        TextGenerationTask,
			path:        completionsPath,
			contentType: httpclient.MIMETypeJSON,
		},
		{
			name:        "text embeddings",
			task:        TextEmbeddingsTask,
			path:        embeddingsPath,
			contentType: httpclient.MIMETypeJSON,
		},
		{
			name:        "speech recognition",
			task:        SpeechRecognitionTask,
			path:        transcriptionsPath,
			contentType: "multipart/form-data; boundary=.*",
		},
		{
			name:        "text to speech",
			task:        TextToSpeechTask,
			path:        createSpeechPath,
			contentType: httpclient.MIMETypeJSON,
		},
		{
			name:        "text to image",
			task:        TextToImageTask,
			path:        imgGenerationPath,
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

				c.Check(r.Header.Get("OpenAI-Organization"), qt.Equals, org)
				c.Check(r.Header.Get("Authorization"), qt.Equals, "Bearer "+apiKey)

				c.Check(r.Header.Get("Content-Type"), qt.Matches, tc.contentType)

				w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintln(w, errResp)
			})

			openAIServer := httptest.NewServer(h)
			c.Cleanup(openAIServer.Close)

			connection, err := structpb.NewStruct(map[string]any{
				"base_path":    openAIServer.URL,
				"api_key":      apiKey,
				"organization": org,
			})
			c.Assert(err, qt.IsNil)

			exec, err := connector.CreateExecution(nil, connection, tc.task)
			c.Assert(err, qt.IsNil)

			pbIn := new(structpb.Struct)
			_, err = exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})
			c.Check(err, qt.IsNotNil)

			want := "OpenAI responded with a 401 status code. Incorrect API key provided."
			c.Check(errmsg.Message(err), qt.Equals, want)
		})
	}

	c.Run("nok - unsupported task", func(c *qt.C) {
		task := "FOOBAR"
		exec, err := connector.CreateExecution(nil, new(structpb.Struct), task)
		c.Assert(err, qt.IsNil)

		pbIn := new(structpb.Struct)
		_, err = exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})
		c.Check(err, qt.IsNotNil)

		want := "FOOBAR task is not supported."
		c.Check(errmsg.Message(err), qt.Equals, want)
	})
}

func TestConnector_Test(t *testing.T) {
	c := qt.New(t)

	bc := base.Connector{Logger: zap.NewNop()}
	connector := Init(bc)

	c.Run("nok - error", func(c *qt.C) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Check(r.Method, qt.Equals, http.MethodGet)
			c.Check(r.URL.Path, qt.Equals, listModelsPath)

			w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, errResp)
		})

		openAIServer := httptest.NewServer(h)
		c.Cleanup(openAIServer.Close)

		connection, err := structpb.NewStruct(map[string]any{
			"base_path": openAIServer.URL,
		})
		c.Assert(err, qt.IsNil)

		err = connector.Test(nil, connection)
		c.Check(err, qt.IsNotNil)

		wantMsg := "OpenAI responded with a 401 status code. Incorrect API key provided."
		c.Check(errmsg.Message(err), qt.Equals, wantMsg)
	})

	c.Run("ok - disconnected", func(c *qt.C) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Check(r.Method, qt.Equals, http.MethodGet)
			c.Check(r.URL.Path, qt.Equals, listModelsPath)

			w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
			fmt.Fprintln(w, `{}`)
		})

		openAIServer := httptest.NewServer(h)
		c.Cleanup(openAIServer.Close)

		connection, err := structpb.NewStruct(map[string]any{
			"base_path": openAIServer.URL,
		})
		c.Assert(err, qt.IsNil)

		err = connector.Test(nil, connection)
		c.Check(err, qt.IsNotNil)
	})

	c.Run("ok - connected", func(c *qt.C) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Check(r.Method, qt.Equals, http.MethodGet)
			c.Check(r.URL.Path, qt.Equals, listModelsPath)

			w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
			fmt.Fprintln(w, `{"data": [{}]}`)
		})

		openAIServer := httptest.NewServer(h)
		c.Cleanup(openAIServer.Close)

		connection, err := structpb.NewStruct(map[string]any{
			"base_path": openAIServer.URL,
		})
		c.Assert(err, qt.IsNil)

		err = connector.Test(nil, connection)
		c.Check(err, qt.IsNil)
	})
}

func TestConnector_WithConfig(t *testing.T) {
	c := qt.New(t)
	cleanupConn := func() { once = sync.Once{} }
	ctx := context.Background()

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

	pbOut, err := base.ConvertToStructpb(want)
	c.Assert(err, qt.IsNil)
	outputs := []*structpb.Struct{pbOut}

	openAIServer := httptest.NewServer(h)
	c.Cleanup(openAIServer.Close)

	task := TextGenerationTask
	bc := base.Connector{Logger: zap.NewNop()}

	c.Run("nok - usage handler check error", func(c *qt.C) {
		c.Cleanup(cleanupConn)

		uh := mock.NewUsageHandlerMock(c)
		uh.CheckMock.When(minimock.AnyContext, inputs).Then(fmt.Errorf("check error"))
		creator := usageHandlerCreator{uh}
		connector := Init(bc).WithUsageHandlerCreator(creator.newUH)

		connection, err := structpb.NewStruct(map[string]any{})
		c.Assert(err, qt.IsNil)

		exec, err := connector.CreateExecution(nil, connection, task)
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
		connector := Init(bc).WithUsageHandlerCreator(creator.newUH)

		connection, err := structpb.NewStruct(map[string]any{
			"base_path": openAIServer.URL,
			"api_key":   apiKey,
		})
		c.Assert(err, qt.IsNil)

		exec, err := connector.CreateExecution(nil, connection, task)
		c.Assert(err, qt.IsNil)

		_, err = exec.Execute(ctx, inputs)
		c.Check(err, qt.IsNotNil)
		c.Check(err, qt.ErrorMatches, "collect error")
	})

	c.Run("ok - with usage handler", func(c *qt.C) {
		c.Cleanup(cleanupConn)

		uh := mock.NewUsageHandlerMock(c)
		uh.CheckMock.When(minimock.AnyContext, inputs).Then(nil)
		uh.CollectMock.When(minimock.AnyContext, inputs, outputs).Then(nil)
		creator := usageHandlerCreator{uh}
		connector := Init(bc).WithUsageHandlerCreator(creator.newUH)

		connection, err := structpb.NewStruct(map[string]any{
			"base_path": openAIServer.URL,
			"api_key":   apiKey,
		})
		c.Assert(err, qt.IsNil)

		exec, err := connector.CreateExecution(nil, connection, task)
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

		connection, err := structpb.NewStruct(map[string]any{
			"base_path": openAIServer.URL,
			"api_key":   "__INSTILL_SECRET", // will be replaced by secrets.apikey
		})
		c.Assert(err, qt.IsNil)

		exec, err := connector.CreateExecution(nil, connection, task)
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
		connection, err := structpb.NewStruct(map[string]any{
			"api_key": "__INSTILL_SECRET",
		})
		c.Assert(err, qt.IsNil)

		_, err = connector.CreateExecution(nil, connection, task)
		c.Check(err, qt.IsNotNil)
		c.Check(err, qt.ErrorMatches, "unresolved global secret")
		c.Check(errmsg.Message(err), qt.Equals, "The configuration field api_key can't reference a global secret.")
	})
}

type usageHandlerCreator struct {
	uh base.UsageHandler
}

func (c usageHandlerCreator) newUH(base.IExecution) (base.UsageHandler, error) {
	return c.uh, nil
}
