package base

import "google.golang.org/protobuf/types/known/structpb"

// UsageHandler allows the component execution wrapper to add checks and
// collect usage metrics around a component execution.
type UsageHandler interface {
	Check(inputs []*structpb.Struct) error
	Collect(inputs, outputs []*structpb.Struct) error
}

// UsageHandlerCreator returns a function to initialize a UsageHandler.
type UsageHandlerCreator func(IExecution) UsageHandler

type noopUsageHandler struct{}

func (h *noopUsageHandler) Check([]*structpb.Struct) error { return nil }
func (h *noopUsageHandler) Collect(_, _ []*structpb.Struct) error {
	return nil
}

func newNoopUsageHandler(IExecution) UsageHandler {
	return new(noopUsageHandler)
}
