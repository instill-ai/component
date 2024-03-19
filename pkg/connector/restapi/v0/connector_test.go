package restapi

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/component/pkg/connector/util/httpclient"
	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
	"github.com/instill-ai/x/errmsg"
)

const (
	errResp = `{"message": "Bad request"}`
	okResp  = `{"title": "Be the wheel"}`
)

var (
	path = "/good-songs/10"
)

func TestConnector_Execute(t *testing.T) {
	c := qt.New(t)

	logger := zap.NewNop()
	connector := Init(logger)
	defID := uuid.Must(uuid.NewV4())
	reqBody := map[string]any{
		"title": "Be the wheel",
	}

	c.Run("nok - unsupported task", func(c *qt.C) {
		task := "FOOBAR"

		exec, err := connector.CreateExecution(defID, task, cfg(noAuthType), logger)
		c.Assert(err, qt.IsNil)

		pbIn := new(structpb.Struct)
		_, err = exec.Execute([]*structpb.Struct{pbIn})
		c.Check(err, qt.IsNotNil)

		want := "FOOBAR task is not supported."
		c.Check(errmsg.Message(err), qt.Equals, want)
	})

	c.Run("ok - POST, 400, basic auth", func(c *qt.C) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Check(r.Method, qt.Equals, http.MethodPost)
			c.Check(r.URL.Path, qt.Matches, path)

			auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
			c.Check(r.Header.Get("Authorization"), qt.Equals, "Basic "+auth)

			c.Assert(r.Body, qt.IsNotNil)
			defer r.Body.Close()

			body, err := io.ReadAll(r.Body)
			c.Assert(err, qt.IsNil)
			c.Check(body, qt.JSONEquals, reqBody)

			w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, errResp)
		})

		srv := httptest.NewServer(h)
		c.Cleanup(srv.Close)

		exec, err := connector.CreateExecution(defID, taskPost, cfg(basicAuthType), logger)
		c.Assert(err, qt.IsNil)

		pbIn, err := base.ConvertToStructpb(TaskInput{
			EndpointURL: srv.URL + path,
			Body:        reqBody,
		})
		c.Assert(err, qt.IsNil)

		got, err := exec.Execute([]*structpb.Struct{pbIn})
		c.Check(err, qt.IsNil)

		resp := got[0].AsMap()
		c.Check(resp["status_code"], qt.Equals, float64(http.StatusBadRequest))
		c.Check(resp["body"], qt.ContentEquals, map[string]any{"message": "Bad request"})
	})

	c.Run("ok - PUT + query auth", func(c *qt.C) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Check(r.Method, qt.Equals, http.MethodPut)
			c.Check(r.URL.Path, qt.Matches, path)

			c.Check(r.FormValue(authKey), qt.Equals, authValue)
			w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
			fmt.Fprintln(w, okResp)
		})

		srv := httptest.NewServer(h)
		c.Cleanup(srv.Close)

		exec, err := connector.CreateExecution(defID, taskPut, cfg(apiKeyType), logger)
		c.Assert(err, qt.IsNil)

		pbIn, err := base.ConvertToStructpb(TaskInput{
			EndpointURL: srv.URL + path,
			Body:        reqBody,
		})

		c.Assert(err, qt.IsNil)

		got, err := exec.Execute([]*structpb.Struct{pbIn})
		c.Check(err, qt.IsNil)

		resp := got[0].AsMap()
		c.Check(resp["status_code"], qt.Equals, float64(http.StatusOK))
		c.Check(resp["body"], qt.ContentEquals, map[string]any{"title": "Be the wheel"})
	})

	c.Run("ok - GET + bearer auth", func(c *qt.C) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Check(r.Method, qt.Equals, http.MethodGet)
			c.Check(r.URL.Path, qt.Matches, path)

			c.Check(r.Header.Get("Authorization"), qt.Equals, "Bearer "+token)

			w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
			fmt.Fprintln(w, okResp)
		})

		srv := httptest.NewServer(h)
		c.Cleanup(srv.Close)

		exec, err := connector.CreateExecution(defID, taskGet, cfg(bearerTokenType), logger)
		c.Assert(err, qt.IsNil)

		pbIn, err := base.ConvertToStructpb(TaskInput{
			EndpointURL: srv.URL + path,
			Body:        reqBody,
		})
		c.Assert(err, qt.IsNil)

		got, err := exec.Execute([]*structpb.Struct{pbIn})
		c.Check(err, qt.IsNil)

		resp := got[0].AsMap()
		c.Check(resp["status_code"], qt.Equals, float64(http.StatusOK))
		c.Check(resp["body"], qt.ContentEquals, map[string]any{"title": "Be the wheel"})
	})
}

func TestConnector_Test(t *testing.T) {
	c := qt.New(t)

	logger := zap.NewNop()
	connector := Init(logger)
	defID := uuid.Must(uuid.NewV4())

	c.Run("ok - connected (even with non-2xx status", func(c *qt.C) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Check(r.Method, qt.Equals, http.MethodGet)

			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "plain/text")
			fmt.Fprintln(w, "Not Found")
		})

		srv := httptest.NewServer(h)
		c.Cleanup(srv.Close)

		got, err := connector.Test(defID, cfg(noAuthType), logger)
		c.Check(err, qt.IsNil)
		c.Check(got, qt.Equals, pipelinePB.Connector_STATE_CONNECTED)
	})
}

const (
	username  = "foo"
	password  = "bar"
	token     = "123"
	authKey   = "api-key"
	authValue = "321"
)

var testAuth = map[authType]map[string]any{
	noAuthType: map[string]any{},
	basicAuthType: map[string]any{
		"username": username,
		"password": password,
	},
	bearerTokenType: map[string]any{"token": token},
	apiKeyType: map[string]any{
		"auth_location": string(query),
		"key":           authKey,
		"value":         authValue,
	},
}

func cfg(atype authType) *structpb.Struct {
	auth := testAuth[atype]
	auth["auth_type"] = string(atype)
	config, _ := structpb.NewStruct(map[string]any{
		"authentication": auth,
	})

	return config
}
