package jira

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/util/httpclient"
	"github.com/instill-ai/x/errmsg"
	"google.golang.org/protobuf/types/known/structpb"
)

type Client struct {
	*httpclient.Client
	APIBaseURL string `json:"api-base-url"`
	Domain     string `json:"domain"`
	CloudID    string `json:"cloud-id"`
}

type CloudID struct {
	ID string `json:"cloudId"`
}

type AuthConfig struct {
	Email   string `json:"email"`
	Token   string `json:"token"`
	BaseURL string `json:"base-url"`
}

func newClient(_ context.Context, setup *structpb.Struct) (*Client, error) {
	var authConfig AuthConfig
	if err := base.ConvertFromStructpb(setup, &authConfig); err != nil {
		return nil, err
	}

	email := authConfig.Email
	token := authConfig.Token
	baseURL := authConfig.BaseURL
	if token == "" {
		return nil, errmsg.AddMessage(
			fmt.Errorf("token not provided"),
			"token not provided",
		)
	}
	if email == "" {
		return nil, errmsg.AddMessage(
			fmt.Errorf("email not provided"),
			"email not provided",
		)
	}
	cloudID, err := getCloudID(baseURL)
	if err != nil {
		return nil, err
	}

	jiraClient := httpclient.New(
		"Jira-Client",
		baseURL,
		httpclient.WithEndUserError(new(errBody)),
	)
	jiraClient.
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetBasicAuth(email, token)
	client := &Client{
		Client:     jiraClient,
		APIBaseURL: apiBaseURL,
		Domain:     baseURL,
		CloudID:    cloudID,
	}
	return client, nil
}

func getCloudID(baseURL string) (string, error) {
	client := httpclient.New("Get-Domain-ID", baseURL, httpclient.WithEndUserError(new(errBody)))
	resp := CloudID{}
	req := client.R().SetResult(&resp)
	// See https://developer.atlassian.com/cloud/jira/software/rest/intro/#base-url-differences
	if _, err := req.Get("_edge/tenant_info"); err != nil {
		return "", err
	}
	return resp.ID, nil
}

type errBody struct {
	Body struct {
		Msg []string `json:"errorMessages"`
	} `json:"body"`
}

func (e errBody) Message() string {
	return strings.Join(e.Body.Msg, " ")
}

func addQueryOptions(req *resty.Request, opt interface{}) error {
	var debug DebugSession
	debug.SessionStart("addQueryOptions", StaticVerboseLevel)
	defer debug.SessionEnd()

	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil
	}
	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			if v.MapIndex(key).IsValid() && v.MapIndex(key).CanInterface() {
				val := v.MapIndex(key).Interface()
				var stringVal string
				switch val := val.(type) {
				case string:
					stringVal = val
				case int:
					stringVal = fmt.Sprintf("%d", val)
				case bool:
					stringVal = fmt.Sprintf("%t", val)
				default:
					continue
				}
				if stringVal == fmt.Sprintf("%v", reflect.Zero(reflect.TypeOf(val))) {
					debug.AddMessage(key.String(), "Default value is not set. Skipping.")
					continue
				}
				paramName := key.String()
				req.SetQueryParam(paramName, stringVal)
			}
		}
	} else if v.Kind() == reflect.Struct {
		typeOfS := v.Type()
		for i := 0; i < v.NumField(); i++ {
			if !v.Field(i).IsValid() || !v.Field(i).CanInterface() {
				debug.AddMessage(typeOfS.Field(i).Name, "Not a valid field")
				continue
			}
			val := v.Field(i).Interface()
			var stringVal string
			switch val := val.(type) {
			case string:
				stringVal = val
			case int:
				stringVal = fmt.Sprintf("%d", val)
			case bool:
				stringVal = fmt.Sprintf("%t", val)
			default:
				continue
			}
			if stringVal == fmt.Sprintf("%v", reflect.Zero(reflect.TypeOf(val))) {
				debug.AddMessage(typeOfS.Field(i).Name, "Default value is not set. Skipping.")
				continue
			}
			paramName := typeOfS.Field(i).Tag.Get("struct")
			if paramName == "" {
				paramName = typeOfS.Field(i).Name
			}
			req.SetQueryParam(paramName, stringVal)
		}
	}
	return nil
}
