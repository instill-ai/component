package stabilityai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/component/pkg/connector/util/httpclient"
	"github.com/instill-ai/x/errmsg"
)

const (
	apiKey  = "123"
	errResp = `
{
  "id": "6e958442e7911ffb2e0bf89c6efe804f",
  "message": "Incorrect API key provided",
  "name": "unauthorized"
}`

	okResp = `
{
  "artifacts": [
    {
      "base64": "a",
      "seed": 1234,
      "finishReason": "SUCCESS"
    }
  ]
}
`
)

func TestConnector_ExecuteImageFromText(t *testing.T) {
	c := qt.New(t)

	weight := 0.5
	text := "a cat and a dog"
	engine := "engine"

	bc := base.Connector{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name      string
		gotStatus int
		gotResp   string
		wantResp  TextToImageOutput
		wantErr   string
	}{
		{
			name:      "ok - 200",
			gotStatus: http.StatusOK,
			gotResp:   okResp,
			wantResp: TextToImageOutput{
				Images: []string{"data:image/png;base64,a"},
				Seeds:  []uint32{1234},
			},
		},
		{
			name:      "nok - 401",
			gotStatus: http.StatusUnauthorized,
			gotResp:   errResp,
			wantErr:   "Stability AI responded with a 401 status code. Incorrect API key provided",
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.Check(r.Method, qt.Equals, http.MethodPost)
				c.Check(r.URL.Path, qt.Matches, `/v1/generation/.*/text-to-image`)

				c.Check(r.Header.Get("Authorization"), qt.Equals, "Bearer "+apiKey)
				c.Check(r.Header.Get("Content-Type"), qt.Equals, httpclient.MIMETypeJSON)

				w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
				w.WriteHeader(tc.gotStatus)
				fmt.Fprintln(w, tc.gotResp)
			})

			srv := httptest.NewServer(h)
			c.Cleanup(srv.Close)

			connection, err := structpb.NewStruct(map[string]any{
				"base_path": srv.URL,
				"api_key":   apiKey,
			})
			c.Assert(err, qt.IsNil)

			exec, err := connector.CreateExecution(nil, connection, textToImageTask)
			c.Assert(err, qt.IsNil)

			weights := []float64{weight}
			pbIn, err := base.ConvertToStructpb(TextToImageInput{
				Engine:  engine,
				Prompts: []string{text},
				Weights: &weights,
			})
			c.Assert(err, qt.IsNil)

			got, err := exec.Execution.Execute([]*structpb.Struct{pbIn})
			if tc.wantErr != "" {
				c.Check(errmsg.Message(err), qt.Equals, tc.wantErr)
				return
			}

			wantJSON, err := json.Marshal(tc.wantResp)
			c.Assert(err, qt.IsNil)
			c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
		})
	}

	c.Run("nok - unsupported task", func(c *qt.C) {
		task := "FOOBAR"
		exec, err := connector.CreateExecution(nil, new(structpb.Struct), task)
		c.Assert(err, qt.IsNil)

		pbIn := new(structpb.Struct)
		_, err = exec.Execution.Execute([]*structpb.Struct{pbIn})
		c.Check(err, qt.IsNotNil)

		want := "FOOBAR task is not supported."
		c.Check(errmsg.Message(err), qt.Equals, want)
	})
}

func TestConnector_ExecuteImageFromImage(t *testing.T) {
	c := qt.New(t)

	weight := 0.5
	text := "a cat and a dog"
	engine := "engine"

	bc := base.Connector{Logger: zap.NewNop()}
	connector := Init(bc)

	testcases := []struct {
		name      string
		gotStatus int
		gotResp   string
		wantResp  ImageToImageOutput
		wantErr   string
	}{
		{
			name:      "ok - 200",
			gotStatus: http.StatusOK,
			gotResp:   okResp,
			wantResp: ImageToImageOutput{
				Images: []string{"data:image/png;base64,a"},
				Seeds:  []uint32{1234},
			},
		},
		{
			name:      "nok - 401",
			gotStatus: http.StatusUnauthorized,
			gotResp:   errResp,
			wantErr:   "Stability AI responded with a 401 status code. Incorrect API key provided",
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.Check(r.Method, qt.Equals, http.MethodPost)
				c.Check(r.URL.Path, qt.Matches, `/v1/generation/.*/image-to-image`)

				c.Check(r.Header.Get("Authorization"), qt.Equals, "Bearer "+apiKey)
				c.Check(r.Header.Get("Content-Type"), qt.Matches, "multipart/form-data; boundary=.*")

				w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
				w.WriteHeader(tc.gotStatus)
				fmt.Fprintln(w, tc.gotResp)
			})

			srv := httptest.NewServer(h)
			c.Cleanup(srv.Close)

			connection, err := structpb.NewStruct(map[string]any{
				"base_path": srv.URL,
				"api_key":   apiKey,
			})
			c.Assert(err, qt.IsNil)

			exec, err := connector.CreateExecution(nil, connection, imageToImageTask)
			c.Assert(err, qt.IsNil)

			weights := []float64{weight}
			pbIn, err := base.ConvertToStructpb(ImageToImageInput{
				Engine:  engine,
				Prompts: []string{text},
				Weights: &weights,
			})
			c.Assert(err, qt.IsNil)

			got, err := exec.Execution.Execute([]*structpb.Struct{pbIn})
			if tc.wantErr != "" {
				c.Check(errmsg.Message(err), qt.Equals, tc.wantErr)
				return
			}

			wantJSON, err := json.Marshal(tc.wantResp)
			c.Assert(err, qt.IsNil)
			c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
		})
	}

	c.Run("nok - unsupported task", func(c *qt.C) {
		task := "FOOBAR"
		exec, err := connector.CreateExecution(nil, new(structpb.Struct), task)
		c.Assert(err, qt.IsNil)

		pbIn := new(structpb.Struct)
		_, err = exec.Execution.Execute([]*structpb.Struct{pbIn})
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
			c.Check(r.URL.Path, qt.Equals, listEnginesPath)

			w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, errResp)
		})

		srv := httptest.NewServer(h)
		c.Cleanup(srv.Close)

		connection, err := structpb.NewStruct(map[string]any{
			"base_path": srv.URL,
		})
		c.Assert(err, qt.IsNil)

		err = connector.Test(nil, connection)
		c.Check(err, qt.IsNotNil)

		wantMsg := "Stability AI responded with a 401 status code. Incorrect API key provided"
		c.Check(errmsg.Message(err), qt.Equals, wantMsg)
	})

	c.Run("ok - disconnected", func(c *qt.C) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Check(r.Method, qt.Equals, http.MethodGet)
			c.Check(r.URL.Path, qt.Equals, listEnginesPath)

			w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
			fmt.Fprintln(w, `[]`)
		})

		srv := httptest.NewServer(h)
		c.Cleanup(srv.Close)

		connection, err := structpb.NewStruct(map[string]any{
			"base_path": srv.URL,
		})
		c.Assert(err, qt.IsNil)

		err = connector.Test(nil, connection)
		c.Check(err, qt.IsNotNil)
	})

	c.Run("ok - connected", func(c *qt.C) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Check(r.Method, qt.Equals, http.MethodGet)
			c.Check(r.URL.Path, qt.Equals, listEnginesPath)

			w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
			fmt.Fprintln(w, `[{}]`)
		})

		srv := httptest.NewServer(h)
		c.Cleanup(srv.Close)

		connection, err := structpb.NewStruct(map[string]any{
			"base_path": srv.URL,
		})
		c.Assert(err, qt.IsNil)

		err = connector.Test(nil, connection)
		c.Check(err, qt.IsNil)
	})
}
