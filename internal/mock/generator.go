package mock

//go:generate minimock -g -i github.com/instill-ai/component/base.UsageHandler -o ./ -s "_mock.gen.go"
//go:generate minimock -g -i github.com/instill-ai/component/operator/document/v0.commandRunner -o ./ -s "_mock.gen.go"
//go:generate minimock -g -i io.WriteCloser -o ./ -s "_mock.gen.go"
//go:generate minimock -g -i github.com/instill-ai/protogen-go/artifact/artifact/v1alpha.ArtifactPublicServiceClient -o ./ -s "_mock.gen.go"

// Ollama mock is generated in the source package to avoid import cycles.
//go:generate minimock -i github.com/instill-ai/component/ai/ollama/v0.OllamaClientInterface -o ../../ai/ollama/v0 -s "_mock.gen.go" -p ollama
