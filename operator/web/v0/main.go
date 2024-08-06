//go:generate compogen readme ./config ./README.mdx
package web

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"sync"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
)

const (
	taskScrapeWebsite = "TASK_SCRAPE_WEBSITE"
	taskScrapeSitemap = "TASK_SCRAPE_SITEMAP"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte

	once sync.Once
	comp *component
)

type component struct {
	base.Component
}

type execution struct {
	base.ComponentExecution
	execute        func(*structpb.Struct) (*structpb.Struct, error)
	externalCaller func(url string) (ioCloser io.ReadCloser, err error)
}

func Init(bc base.Component) *component {
	once.Do(func() {
		comp = &component{Component: bc}
		err := comp.LoadDefinition(definitionJSON, nil, tasksJSON, nil)
		if err != nil {
			panic(err)
		}
	})
	return comp
}

func (c *component) CreateExecution(x base.ComponentExecution) (base.IExecution, error) {
	e := &execution{
		ComponentExecution: x,
	}

	switch x.Task {
	case taskScrapeWebsite:
		e.execute = e.Scrape
	case taskScrapeSitemap:
		// To make mocking easier
		e.externalCaller = scrapSitemapCaller
		e.execute = e.ScrapeSitemap
	default:
		return nil, fmt.Errorf(x.Task + " task is not supported.")
	}

	return e, nil
}

func (e *execution) Execute(_ context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := make([]*structpb.Struct, len(inputs))

	for i, input := range inputs {
		output, err := e.execute(input)
		if err != nil {
			return nil, err
		}

		outputs[i] = output
	}

	return outputs, nil
}
