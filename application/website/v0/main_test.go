package website

import (
	"io"
	"strings"
	"testing"

	"github.com/frankban/quicktest"
	"github.com/instill-ai/component/base"
)

func TestScrapSiteMap(t *testing.T) {
	c := quicktest.New(t)

	c.Run("ScrapeSitemap", func(c *quicktest.C) {
		component := Init(base.Component{})

		e := &execution{
			ComponentExecution: base.ComponentExecution{Component: component, SystemVariables: nil, Setup: nil, Task: taskScrapeSitemap},
			externalCaller:     fakeScrapSitemapCaller,
		}

		e.execute = e.ScrapeSitemap

		input := &ScrapeSitemapInput{
			URL: "https://www.example.com/sitemap.xml",
		}

		inputStruct, err := base.ConvertToStructpb(input)
		c.Assert(err, quicktest.IsNil)

		output, err := e.execute(inputStruct)

		c.Assert(err, quicktest.IsNil)

		var outputStruct ScrapeSitemapOutput
		err = base.ConvertFromStructpb(output, &outputStruct)
		c.Assert(err, quicktest.IsNil)

		c.Assert(len(outputStruct.List), quicktest.Equals, 1)

		siteInfo := outputStruct.List[0]
		c.Assert(siteInfo.Loc, quicktest.Equals, "https://www.example.com")
		c.Assert(siteInfo.LastModifiedTime, quicktest.Equals, "2021-01-01T00:00:00Z")
		c.Assert(siteInfo.ChangeFrequency, quicktest.Equals, "daily")
		c.Assert(siteInfo.Priority, quicktest.Equals, 0.8)
	})
}

func fakeScrapSitemapCaller(url string) (ioCloser io.ReadCloser, err error) {

	xml := `<?xml version="1.0" encoding="UTF-8"?>`
	xml += `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`
	xml += `<url>`
	xml += `<loc>https://www.example.com</loc>`
	xml += `<lastmod>2021-01-01T00:00:00Z</lastmod>`
	xml += `<changefreq>daily</changefreq>`
	xml += `<priority>0.8</priority>`
	xml += `</url>`
	xml += `</urlset>`
	return io.NopCloser(strings.NewReader(xml)), nil
}
