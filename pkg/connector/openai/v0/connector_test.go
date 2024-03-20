package openai

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/connector/util/httpclient"
	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
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

	logger := zap.NewNop()
	connector := Init(logger)
	defID := uuid.Must(uuid.NewV4())

	testcases := []struct {
		name        string
		task        string
		path        string
		contentType string
	}{
		{
			name:        "text generation",
			task:        textGenerationTask,
			path:        completionsPath,
			contentType: httpclient.MIMETypeJSON,
		},
		{
			name:        "text embeddings",
			task:        textEmbeddingsTask,
			path:        embeddingsPath,
			contentType: httpclient.MIMETypeJSON,
		},
		{
			name:        "speech recognition",
			task:        speechRecognitionTask,
			path:        transcriptionsPath,
			contentType: "multipart/form-data; boundary=.*",
		},
		{
			name:        "text to speech",
			task:        textToSpeechTask,
			path:        createSpeechPath,
			contentType: httpclient.MIMETypeJSON,
		},
		{
			name:        "text to image",
			task:        textToImageTask,
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

			config, err := structpb.NewStruct(map[string]any{
				"base_path":    openAIServer.URL,
				"api_key":      apiKey,
				"organization": org,
			})
			c.Assert(err, qt.IsNil)

			exec, err := connector.CreateExecution(defID, tc.task, config, logger)
			c.Assert(err, qt.IsNil)

			pbIn := new(structpb.Struct)
			_, err = exec.Execute([]*structpb.Struct{pbIn})
			c.Check(err, qt.IsNotNil)

			want := "OpenAI responded with a 401 status code. Incorrect API key provided."
			c.Check(errmsg.Message(err), qt.Equals, want)
		})
	}

	c.Run("nok - unsupported task", func(c *qt.C) {
		task := "FOOBAR"
		exec, err := connector.CreateExecution(defID, task, new(structpb.Struct), logger)
		c.Assert(err, qt.IsNil)

		pbIn := new(structpb.Struct)
		_, err = exec.Execute([]*structpb.Struct{pbIn})
		c.Check(err, qt.IsNotNil)

		want := "FOOBAR task is not supported."
		c.Check(errmsg.Message(err), qt.Equals, want)
	})
}

func TestConnector_Test(t *testing.T) {
	c := qt.New(t)

	logger := zap.NewNop()
	connector := Init(logger)
	defID := uuid.Must(uuid.NewV4())

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

		config, err := structpb.NewStruct(map[string]any{
			"base_path": openAIServer.URL,
		})
		c.Assert(err, qt.IsNil)

		got, err := connector.Test(defID, config, logger)
		c.Check(err, qt.IsNotNil)
		c.Check(got, qt.Equals, pipelinePB.Connector_STATE_ERROR)

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

		config, err := structpb.NewStruct(map[string]any{
			"base_path": openAIServer.URL,
		})
		c.Assert(err, qt.IsNil)

		got, err := connector.Test(defID, config, logger)
		c.Check(err, qt.IsNil)
		c.Check(got, qt.Equals, pipelinePB.Connector_STATE_DISCONNECTED)
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

		config, err := structpb.NewStruct(map[string]any{
			"base_path": openAIServer.URL,
		})
		c.Assert(err, qt.IsNil)

		got, err := connector.Test(defID, config, logger)
		c.Check(err, qt.IsNil)
		c.Check(got, qt.Equals, pipelinePB.Connector_STATE_CONNECTED)
	})
}
