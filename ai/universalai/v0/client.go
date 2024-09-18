package universalai

import (
	"fmt"

	openaiv1 "github.com/instill-ai/component/ai/openai/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func newClient(setup *structpb.Struct, logger *zap.Logger, vendor string) (interface{}, error) {
	switch vendor {
	case "openai":
		return openaiv1.NewClient(setup, logger), nil
	default:
		return nil, fmt.Errorf("unsupported vendor: %s", vendor)
	}
}
