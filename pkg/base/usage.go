package base

// TODO: add parameters for usage check and collection
type UsageHandler interface {
	Check() error
	Collect() error
}
