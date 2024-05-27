package archetypeai

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/component/pkg/util/httpclient"
	"github.com/instill-ai/x/errmsg"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	apiKey = "213bac"
)

const errJSON = `{ "error": "Invalid access." }`
const describeJSON = `
{
  "query_id": "2401242b4d59e48bbf6e0d",
  "status": "completed",
  "inference_time_sec": 1.6635565757751465,
  "query_response_time_sec": 6.018876314163208,
  "response": [
	{
	  "timestamp": 2.0,
	  "frame_id": 60,
	  "description": "The group of people is walking across a bridge."
	},
	{
	  "timestamp": 6.0,
	  "frame_id": 180,
	  "description": "The man is walking across a bridge, and he is surrounded by people."
	}
  ]
}`
const describeErrJSON = `
{
  "query_id": "2401242b4d59e48bbf6e0d",
  "status": "failed",
  "inference_time_sec": 1.6635565757751465,
  "query_response_time_sec": 6.018876314163208,
  "response": [
    {
      "timestamp": 2.0,
      "frame_id": 60,
      "description": "The group of people is walking across a bridge."
    }
  ]
}`
const summarizeJSON = `
{
  "query_id": "240123b93a83a79e9907a5",
  "status": "completed",
  "file_ids": [
    "test_image.jpg"
  ],
  "inference_time_sec": 2.1776912212371826,
  "query_response_time_sec": 2.1914472579956055,
  "response": {
    "processed_text": "A family of four is hiking together on a trail."
  }
}`
const summarizeErrJSON = `
{
  "query_id": "2401233472bde249e60260",
  "status": "failed",
  "file_ids": [
    "test_image.jpg"
  ]
}`
const uploadFileJSON = `
{
  "is_valid": true,
  "file_id": "2084fa42-8452-4fa6-bed9-6aac6d6153bb",
  "file_uid": "2401242e3cb25122835a17"
}`
const uploadErrJSON = `
{
  "is_valid": false,
  "errors": [
    "Invalid file type: application/octet-stream. Supported file types are: ('image/jpeg', 'image/png', 'video/mp4')."
  ]
}`

var (
	queryIn = fileQueryParams{
		Query:   "Describe what's happening",
		FileIDs: []string{"test.file"},
	}
	uploadFileIn = uploadFileParams{
		File: "data:text/plain;base64,aG9sYQ==",
	}
)

func TestConnector_Execute(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()

	testcases := []struct {
		name string

		task    string
		in      any
		want    any
		wantErr string

		// server expectations and response
		wantPath        string
		wantReq         any
		wantContentType string
		gotStatus       int
		gotResp         string
	}{
		{
			name: "ok - describe",

			task: taskDescribe,
			in:   queryIn,
			want: describeOutput{
				Descriptions: []frameDescription{
					{
						Timestamp:   2.0,
						FrameID:     60,
						Description: "The group of people is walking across a bridge.",
					},
					{
						Timestamp:   6.0,
						FrameID:     180,
						Description: "The man is walking across a bridge, and he is surrounded by people.",
					},
				},
			},

			wantPath:        describePath,
			wantReq:         queryIn,
			wantContentType: httpclient.MIMETypeJSON,
			gotStatus:       http.StatusOK,
			gotResp:         describeJSON,
		},
		{
			name: "nok - describe error",

			task:    taskDescribe,
			in:      queryIn,
			wantErr: `Archetype AI didn't complete query 2401242b4d59e48bbf6e0d: status is "failed".`,

			wantPath:        describePath,
			wantReq:         queryIn,
			wantContentType: httpclient.MIMETypeJSON,
			gotStatus:       http.StatusOK,
			gotResp:         describeErrJSON,
		},
		{
			name: "ok - summarize",

			task: taskSummarize,
			in:   queryIn,
			want: summarizeOutput{
				Response: "A family of four is hiking together on a trail.",
			},

			wantPath:        summarizePath,
			wantReq:         queryIn,
			wantContentType: httpclient.MIMETypeJSON,
			gotStatus:       http.StatusOK,
			gotResp:         summarizeJSON,
		},
		{
			name: "nok - summarize wrong file",

			task:    taskSummarize,
			in:      queryIn,
			wantErr: `Archetype AI didn't complete query 2401233472bde249e60260: status is "failed".`,

			wantPath:        summarizePath,
			wantReq:         queryIn,
			wantContentType: httpclient.MIMETypeJSON,
			gotStatus:       http.StatusOK,
			gotResp:         summarizeErrJSON,
		},
		{
			name: "ok - upload file",

			task: taskUploadFile,
			in:   uploadFileIn,
			want: uploadFileOutput{FileID: "2084fa42-8452-4fa6-bed9-6aac6d6153bb"},

			wantPath:        uploadFilePath,
			wantReq:         "hola",
			wantContentType: "multipart/form-data.*",
			gotStatus:       http.StatusOK,
			gotResp:         uploadFileJSON,
		},
		{
			name: "nok - upload invalid file",

			task:    taskUploadFile,
			in:      uploadFileIn,
			wantErr: "Couldn't complete upload: Invalid file type.*",

			wantPath:        uploadFilePath,
			wantReq:         "hola",
			wantContentType: "multipart/form-data.*",
			gotStatus:       http.StatusOK,
			gotResp:         uploadErrJSON,
		},
		{
			name: "nok - unauthorized",

			task:    taskSummarize,
			in:      queryIn,
			wantErr: "Archetype AI responded with a 401 status code. Invalid access.",

			wantPath:        summarizePath,
			wantReq:         queryIn,
			wantContentType: httpclient.MIMETypeJSON,
			gotStatus:       http.StatusUnauthorized,
			gotResp:         errJSON,
		},
	}

	bc := base.Connector{Logger: zap.NewNop()}
	connector := Init(bc)

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.Check(r.Method, qt.Equals, http.MethodPost)
				c.Check(r.URL.Path, qt.Matches, tc.wantPath)

				c.Check(r.Header.Get("Authorization"), qt.Equals, "Bearer "+apiKey)
				c.Check(r.Header.Get("Content-Type"), qt.Matches, tc.wantContentType)

				body, err := io.ReadAll(r.Body)
				c.Assert(err, qt.IsNil)
				if tc.wantContentType == httpclient.MIMETypeJSON {
					c.Check(body, qt.JSONEquals, tc.wantReq)
				} else {
					// We just do partial match to avoid matching every field
					// in multipart bodies.
					c.Check(string(body), qt.Contains, tc.wantReq)
				}

				w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
				w.WriteHeader(tc.gotStatus)
				fmt.Fprintln(w, tc.gotResp)
			})

			srv := httptest.NewServer(h)
			c.Cleanup(srv.Close)

			connection, _ := structpb.NewStruct(map[string]any{
				"base_path": srv.URL,
				"api_key":   apiKey,
			})

			exec, err := connector.CreateExecution(nil, connection, tc.task)
			c.Assert(err, qt.IsNil)

			pbIn, err := base.ConvertToStructpb(tc.in)
			c.Assert(err, qt.IsNil)

			got, err := exec.Execution.Execute(ctx, []*structpb.Struct{pbIn})
			if tc.wantErr != "" {
				c.Check(errmsg.Message(err), qt.Matches, tc.wantErr)
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

func TestConnector_CreateExecution(t *testing.T) {
	c := qt.New(t)

	bc := base.Connector{Logger: zap.NewNop()}
	connector := Init(bc)

	c.Run("nok - unsupported task", func(c *qt.C) {
		task := "FOOBAR"
		want := fmt.Sprintf("%s task is not supported.", task)

		_, err := connector.CreateExecution(nil, new(structpb.Struct), task)
		c.Check(err, qt.IsNotNil)
		c.Check(errmsg.Message(err), qt.Equals, want)
	})
}

func TestConnector_Test(t *testing.T) {
	c := qt.New(t)

	bc := base.Connector{Logger: zap.NewNop()}
	connector := Init(bc)

	c.Run("ok - connected", func(c *qt.C) {
		err := connector.Test(nil, nil)
		c.Check(err, qt.IsNil)
	})
}
