package jira

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-resty/resty/v2"
	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/util/httpclient"
	"github.com/instill-ai/x/errmsg"
	"google.golang.org/protobuf/types/known/structpb"
)

type Client struct {
	*httpclient.Client
	APIBaseURL string `json:"api_base_url"`
	Domain     string `json:"domain"`
	CloudID    string `json:"cloud_id"`
}

type CloudID struct {
	ID string `json:"cloudId"`
}

type AuthConfig struct {
	Email   string `json:"email"`
	Token   string `json:"token"`
	BaseURL string `json:"base_url"`
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
	jiraClient.SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json")
	jiraClient.SetBasicAuth(email, token)
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
	Msg    string `json:"message"`
	Status int    `json:"status"`
}

func (e errBody) Message() string {
	return fmt.Sprintf("%d %s", e.Status, e.Msg)
}

func addQueryOptions(req *resty.Request, opt interface{}) error {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil
	}
	typeOfS := v.Type()
	for i := 0; i < v.NumField(); i++ {
		stringVal, ok := v.Field(i).Interface().(string)
		if !ok {
			intVal, ok := v.Field(i).Interface().(int)
			if !ok {
				continue
			}
			stringVal = fmt.Sprintf("%d", intVal)
		}
		if stringVal == "" {
			continue
		}
		paramName := typeOfS.Field(i).Tag.Get("struct")
		if paramName == "" {
			paramName = typeOfS.Field(i).Name
		}
		req.SetQueryParam(paramName, stringVal)
	}
	return nil
}
