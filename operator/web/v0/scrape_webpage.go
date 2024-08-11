package web

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/util"
	"github.com/k3a/html2text"
	"google.golang.org/protobuf/types/known/structpb"
)

type ScrapeWebpageInput struct {
	URL             string   `json:"url"`
	IncludeHTML     bool     `json:"include-html"`
	OnlyMainContent bool     `json:"only-main-content"`
	RemoveTags      []string `json:"remove-tags,omitempty"`
	OnlyIncludeTags []string `json:"only-include-tags,omitempty"`
}

type ScrapeWebpageOutput struct {
	Content     string   `json:"content"`
	Markdown    string   `json:"markdown"`
	HTML        string   `json:"html"`
	Metadata    Metadata `json:"metadata"`
	LinksOnPage []string `json:"links-on-page"`
}

type Metadata struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	SourceURL   string `json:"source-url"`
}

func (e *execution) ScrapeWebpage(input *structpb.Struct) (*structpb.Struct, error) {

	inputStruct := ScrapeWebpageInput{}

	err := base.ConvertFromStructpb(input, &inputStruct)

	if err != nil {
		return nil, fmt.Errorf("error converting input to struct: %v", err)
	}

	output := ScrapeWebpageOutput{}

	doc, err := e.request(inputStruct.URL)

	if err != nil {
		return nil, fmt.Errorf("error getting HTML page doc: %v", err)
	}

	html := getRemovedTagsHTML(doc, inputStruct)

	err = setOutput(&output, inputStruct, doc, html)

	if err != nil {
		return nil, fmt.Errorf("error setting output: %v", err)
	}

	return base.ConvertToStructpb(output)

}

func httpRequest(url string) (*goquery.Document, error) {
	client := &http.Client{}
	res, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to %s: %v", url, err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML from %s: %v", url, err)
	}

	return doc, nil
}

func getRemovedTagsHTML(doc *goquery.Document, input ScrapeWebpageInput) string {
	if input.OnlyMainContent {
		removeSelectors := []string{"header", "nav", "footer"}
		for _, selector := range removeSelectors {
			doc.Find(selector).Remove()
		}
	}

	if input.RemoveTags != nil || len(input.RemoveTags) > 0 {
		for _, tag := range input.RemoveTags {
			doc.Find(tag).Remove()
		}
	}

	if input.OnlyIncludeTags == nil || len(input.OnlyIncludeTags) == 0 {
		html, err := doc.Html()
		if err != nil {
			log.Println("error getting HTML: ", err)
			return ""
		}
		return html
	}

	combinedHTML := ""

	tags := buildTags(input.OnlyIncludeTags)
	doc.Find(tags).Each(func(i int, s *goquery.Selection) {
		html, err := s.Html()
		if err != nil {
			log.Println("error getting HTML: ", err)
			combinedHTML += "\n"
		}
		combinedHTML += fmt.Sprintf("<%s>%s</%s>\n", s.Nodes[0].Data, html, s.Nodes[0].Data)
	})

	return combinedHTML
}

func buildTags(tags []string) string {
	tagsString := ""
	for i, tag := range tags {
		tagsString += tag
		if i < len(tags)-1 {
			tagsString += ","
		}
	}
	return tagsString
}

func setOutput(output *ScrapeWebpageOutput, input ScrapeWebpageInput, doc *goquery.Document, html string) error {
	plain := html2text.HTML2Text(html)

	output.Content = plain
	if input.IncludeHTML {
		output.HTML = html
	}

	markdown, err := getMarkdown(html, input.URL)

	if err != nil {
		return fmt.Errorf("failed to get markdown: %v", err)
	}

	output.Markdown = markdown

	title := util.ScrapeWebpageTitle(doc)
	description := util.ScrapeWebpageDescription(doc)

	metadata := Metadata{
		Title:       title,
		Description: description,
		SourceURL:   input.URL,
	}
	output.Metadata = metadata
	output.LinksOnPage = getAllLinksOnPage(doc)

	return nil

}

func getMarkdown(html, url string) (string, error) {
	domain, err := util.GetDomainFromURL(url)

	if err != nil {
		return "", fmt.Errorf("error getting domain from URL: %v", err)
	}

	markdown, err := util.ScrapeWebpageHTMLToMarkdown(html, domain)

	if err != nil {
		return "", fmt.Errorf("error converting HTML to Markdown: %v", err)
	}

	return markdown, nil
}

func getAllLinksOnPage(doc *goquery.Document) []string {
	links := []string{}

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, ok := s.Attr("href")
		if ok {
			links = append(links, link)
		}
	})

	return links
}
