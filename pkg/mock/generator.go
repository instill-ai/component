package mock

//go:generate minimock -g -i github.com/instill-ai/component/pkg/base.UsageHandler -o ./ -s "_mock.gen.go"
//go:generate minimock -g -i github.com/instill-ai/component/pkg/operator/pdf/v0.commandRunner -o ./ -s "_mock.gen.go"
//go:generate minimock -g -i io.WriteCloser -o ./ -s "_mock.gen.go"
