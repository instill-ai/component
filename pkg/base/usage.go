package base

// UsageHandler allows the component execution wrapper to add checks and
// collect usage metrics around a component execution.
type UsageHandler interface {
	Check() error
	Collect() error
}

type noopUsageHandler struct{}

func (h *noopUsageHandler) Check() error   { return nil }
func (h *noopUsageHandler) Collect() error { return nil }
