package base

import "google.golang.org/protobuf/types/known/structpb"

// UsageHandler allows the component execution wrapper to add checks and
// collect usage metrics around a component execution.
type UsageHandler interface {
	Check(task string, usesSecret bool, inputs []*structpb.Struct) error
	Collect(task string, usesSecret bool, inputs, outputs []*structpb.Struct) error
}

type noopUsageHandler struct{}

func (h *noopUsageHandler) Check(string, bool, []*structpb.Struct) error { return nil }
func (h *noopUsageHandler) Collect(_ string, _ bool, _, _ []*structpb.Struct) error {
	return nil
}

func newNoopUsageHandler(IExecution) UsageHandler {
	return new(noopUsageHandler)
}
