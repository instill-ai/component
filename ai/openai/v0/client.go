package openai

import (
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	openaiclient "github.com/sashabaranov/go-openai"
)

func newClient(setup *structpb.Struct, logger *zap.Logger) *openaiclient.Client {

	cfg := openaiclient.DefaultConfig(getAPIKey(setup))
	org := getOrg(setup)
	if org != "" {
		cfg.OrgID = org
	}

	return openaiclient.NewClientWithConfig(cfg)
}

func getAPIKey(setup *structpb.Struct) string {
	return setup.GetFields()[cfgAPIKey].GetStringValue()
}

func getOrg(setup *structpb.Struct) string {
	val, ok := setup.GetFields()[cfgOrganization]
	if !ok {
		return ""
	}
	return val.GetStringValue()
}
