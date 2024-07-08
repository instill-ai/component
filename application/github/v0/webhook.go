package github

import (
	"context"

	"github.com/google/go-github/v62/github"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

type CreateWebHookInput struct {
	RepoInfo
	HookURL     string   `json:"hook_url"`
	HookSecret  string   `json:"hook_secret"`
	Events      []string `json:"events"`
	Active      bool     `json:"active"`
	ContentType string   `json:"content_type"` // including `json`, `form`
}

type HookConfig struct {
	URL         string `json:"url"`
	InsecureSSL string `json:"insecure_ssl"`
	Secret      string `json:"secret,omitempty"`
	ContentType string `json:"content_type"`
}

type HookInfo struct {
	ID      int64      `json:"id"`
	URL     string     `json:"url"`
	PingURL string     `json:"ping_url"`
	TestURL string     `json:"test_url"`
	Config  HookConfig `json:"config"`
}
type CreateWebHookResp struct {
	Hook HookInfo `json:"hook"`
}

func (githubClient *Client) createWebhookTask(props *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct CreateWebHookInput
	err := base.ConvertFromStructpb(props, &inputStruct)
	if err != nil {
		return nil, err
	}

	owner, repository, err := parseTargetRepo(inputStruct)
	if err != nil {
		return nil, err
	}
	hookURL := inputStruct.HookURL
	hookSecret := inputStruct.HookSecret
	originalEvents := inputStruct.Events
	active := inputStruct.Active
	contentType := inputStruct.ContentType
	if contentType != "json" && contentType != "form" {
		contentType = "json"
	}

	hook := &github.Hook{
		Name: github.String("web"), // only webhooks are supported
		Config: &github.HookConfig{
			InsecureSSL: github.String("0"), // SSL verification is required
			URL:         &hookURL,
			Secret:      &hookSecret,
			ContentType: &contentType,
		},
		Events: originalEvents,
		Active: &active,
	}

	hook, _, err = githubClient.Repositories.CreateHook(context.Background(), owner, repository, hook)
	if err != nil {
		return nil, err
	}

	var resp CreateWebHookResp
	hookStruct, err := base.ConvertToStructpb(hook)
	if err != nil {
		return nil, err
	}
	var hookInfo HookInfo
	err = base.ConvertFromStructpb(hookStruct, &hookInfo)
	if err != nil {
		return nil, err
	}
	resp.Hook = hookInfo
	out, err := base.ConvertToStructpb(resp)
	if err != nil {
		return nil, err
	}
	return out, nil
}
