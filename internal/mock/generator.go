package mock

//go:generate minimock -g -i github.com/instill-ai/component/base.UsageHandler -o ./ -s "_mock.gen.go"
//go:generate minimock -g -i github.com/instill-ai/component/operator/document/v0.commandRunner -o ./ -s "_mock.gen.go"
//go:generate minimock -g -i io.WriteCloser -o ./ -s "_mock.gen.go"
