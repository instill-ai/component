package asana

import (
	"context"
	"fmt"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/tools/logger"
	"google.golang.org/protobuf/types/known/structpb"
)

type PortfolioTaskOutput struct {
	Portfolio
}

type PortfolioTaskResp struct {
	Data struct {
		GID                 string            `json:"gid"`
		Name                string            `json:"name"`
		Owner               User              `json:"owner"`
		DueOn               string            `json:"due_on"`
		StartOn             string            `json:"start_on"`
		Color               string            `json:"color"`
		Public              bool              `json:"public"`
		CreatedBy           User              `json:"created_by"`
		CurrentStatus       map[string]string `json:"current_status"`
		CustomFields        map[string]string `json:"custom_fields"`
		CustomFieldSettings map[string]string `json:"custom_field_settings"`
	} `json:"data"`
}

func portfolioResp2Output(resp *PortfolioTaskResp) PortfolioTaskOutput {
	out := PortfolioTaskOutput{
		Portfolio: Portfolio{
			GID:                 resp.Data.GID,
			Name:                resp.Data.Name,
			Owner:               resp.Data.Owner,
			DueOn:               resp.Data.DueOn,
			StartOn:             resp.Data.StartOn,
			Color:               resp.Data.Color,
			Public:              resp.Data.Public,
			CreatedBy:           resp.Data.CreatedBy,
			CurrentStatus:       resp.Data.CurrentStatus,
			CustomFields:        resp.Data.CustomFields,
			CustomFieldSettings: resp.Data.CustomFieldSettings,
		},
	}
	return out
}

type GetPortfolioInput struct {
	Action string `json:"action"`
	ID     string `json:"portfolio-gid"`
}

func (c *Client) GetPortfolio(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("GetPortfolio", logger.Develop).SessionEnd()
	var input GetPortfolioInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}

	debug.Info("input", input)

	apiEndpoint := fmt.Sprintf("/portfolios/%s", input.ID)
	req := c.Client.R().SetResult(&PortfolioTaskResp{})

	wantOptFields := parseWantOptionFields(Portfolio{})
	debug.Info("wantOptFields", wantOptFields)
	if err := addQueryOptions(req, map[string]interface{}{"opt_fields": wantOptFields}); err != nil {
		return nil, err
	}
	resp, err := req.Get(apiEndpoint)
	if err != nil {
		return nil, err
	}

	debug.Info("resp", resp)
	portfolio := resp.Result().(*PortfolioTaskResp)
	debug.Info("portfolio", portfolio)
	out := portfolioResp2Output(portfolio)
	return base.ConvertToStructpb(out)
}

type UpdatePortfolioInput struct {
	Action    string `json:"action"`
	ID        string `json:"portfolio-gid"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	Public    bool   `json:"public"`
	Workspace string `json:"workspace"`
}

type UpdatePortfolioReq struct {
	Name      string `json:"name"`
	Color     string `json:"color"`
	Public    bool   `json:"public"`
	Workspace string `json:"workspace"`
}

func (c *Client) UpdatePortfolio(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("UpdatePortfolio", logger.Develop).SessionEnd()
	var input UpdatePortfolioInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}

	debug.Info("input", input)

	apiEndpoint := fmt.Sprintf("/portfolios/%s", input.ID)
	req := c.Client.R().SetResult(&PortfolioTaskResp{}).SetBody(
		map[string]interface{}{
			"data": &UpdatePortfolioReq{
				Name:      input.Name,
				Color:     input.Color,
				Public:    input.Public,
				Workspace: input.Workspace,
			},
		})

	wantOptFields := parseWantOptionFields(Portfolio{})
	debug.Info("wantOptFields", wantOptFields)
	if err := addQueryOptions(req, map[string]interface{}{"opt_fields": wantOptFields}); err != nil {
		return nil, err
	}

	resp, err := req.Put(apiEndpoint)

	if err != nil {
		return nil, err
	}
	debug.Info("resp", resp)
	portfolio := resp.Result().(*PortfolioTaskResp)
	debug.Info("portfolio", portfolio)
	out := portfolioResp2Output(portfolio)
	return base.ConvertToStructpb(out)
}

type CreatePortfolioInput struct {
	Action    string `json:"action"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	Public    bool   `json:"public"`
	Workspace string `json:"workspace"`
}

type CreatePortfolioReq struct {
	Name      string `json:"name"`
	Color     string `json:"color"`
	Public    bool   `json:"public"`
	Workspace string `json:"workspace"`
}

func (c *Client) CreatePortfolio(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("CreatePortfolio", logger.Develop).SessionEnd()
	var input CreatePortfolioInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}

	debug.Info("input", input)

	apiEndpoint := "/portfolios"
	req := c.Client.R().SetResult(&PortfolioTaskResp{}).SetBody(
		map[string]interface{}{
			"data": &CreatePortfolioReq{
				Name:      input.Name,
				Color:     input.Color,
				Public:    input.Public,
				Workspace: input.Workspace,
			},
		})
	wantOptFields := parseWantOptionFields(Portfolio{})
	debug.Info("wantOptFields", wantOptFields)
	if err := addQueryOptions(req, map[string]interface{}{"opt_fields": wantOptFields}); err != nil {
		return nil, err
	}

	resp, err := req.Post(apiEndpoint)
	if err != nil {
		return nil, err
	}
	portfolio := resp.Result().(*PortfolioTaskResp)
	out := portfolioResp2Output(portfolio)
	return base.ConvertToStructpb(out)
}

type DeletePortfolioInput struct {
	Action string `json:"action"`
	ID     string `json:"portfolio-gid"`
}

func (c *Client) DeletePortfolio(ctx context.Context, props *structpb.Struct) (*structpb.Struct, error) {
	var debug logger.Session
	defer debug.SessionStart("DeletePortfolio", logger.Develop).SessionEnd()
	var input DeletePortfolioInput
	if err := base.ConvertFromStructpb(props, &input); err != nil {
		return nil, err
	}

	debug.Info("input", input)

	apiEndpoint := fmt.Sprintf("/portfolios/%s", input.ID)
	req := c.Client.R()

	_, err := req.Delete(apiEndpoint)
	if err != nil {
		return nil, err
	}
	out := PortfolioTaskOutput{}
	return base.ConvertToStructpb(out)
}
