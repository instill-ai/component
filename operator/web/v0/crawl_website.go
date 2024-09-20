package web

import (
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"strings"

	colly "github.com/gocolly/colly/v2"
	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/util"
	"google.golang.org/protobuf/types/known/structpb"
)

type PageInfo struct {
	Link     string `json:"link"`
	Title    string `json:"title"`
	LinkText string `json:"link-text"`
	LinkHTML string `json:"link-html"`
}

// ScrapeWebsiteInput defines the input of the scrape website task
type ScrapeWebsiteInput struct {
	// TargetURL: The URL of the website to scrape.
	TargetURL string `json:"target-url"`
	// AllowedDomains: The list of allowed domains to scrape.
	AllowedDomains []string `json:"allowed-domains"`
	// MaxK: The maximum number of pages to scrape.
	MaxK int `json:"max-k"`
	// IncludeLinkText: Whether to include the scraped text of the scraped web page.
	IncludeLinkText *bool `json:"include-link-text"`
	// IncludeLinkHTML: Whether to include the scraped HTML of the scraped web page.
	IncludeLinkHTML *bool `json:"include-link-html"`
	// OnlyMainContent: Whether to scrape only the main content of the web page. If true, the scraped text wull exclude the header, nav, footer.
	OnlyMainContent bool `json:"only-main-content"`
	// RemoveTags: The list of tags to remove from the scraped text.
	RemoveTags []string `json:"remove-tags"`
	// OnlyIncludeTags: The list of tags to include in the scraped text.
	OnlyIncludeTags []string `json:"only-include-tags"`
	// Timeout: The number of milliseconds to wait before scraping the web page. Min 0, Max 60000.
	Timeout int `json:"timeout"`
}

// ScrapeWebsiteOutput defines the output of the scrape website task
type ScrapeWebsiteOutput struct {
	// Pages: The list of pages that were scraped.
	Pages []PageInfo `json:"pages"`
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// randomString generates a random string of length 10-20
func randomString() string {
	b := make([]byte, rand.Intn(10)+10)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// stripQueryAndTrailingSlash removes query parameters and trailing '/' from a URL
func stripQueryAndTrailingSlash(u *url.URL) *url.URL {
	// Remove query parameters by setting RawQuery to an empty string
	u.RawQuery = ""

	// Remove trailing '/' from the path
	u.Path = strings.TrimSuffix(u.Path, "/")

	return u
}

// existsInSlice checks if a string exists in a slice
func existsInSlice(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true // Item already exists, so don't add it again
		}
	}
	return false // Item doesn't exist, so add it to the slice
}

// Scrape crawls a webpage and returns a slice of PageInfo
func (e *execution) CrawlWebsite(input *structpb.Struct) (*structpb.Struct, error) {
	inputStruct := ScrapeWebsiteInput{}
	err := base.ConvertFromStructpb(input, &inputStruct)

	if err != nil {
		return nil, fmt.Errorf("error converting input to struct: %v", err)
	}

	output := ScrapeWebsiteOutput{}

	if inputStruct.IncludeLinkHTML == nil {
		b := false
		inputStruct.IncludeLinkHTML = &b
	}
	if inputStruct.IncludeLinkText == nil {
		b := false
		inputStruct.IncludeLinkText = &b
	}
	if inputStruct.MaxK < 0 {
		inputStruct.MaxK = 0
	}

	pageLinks := []string{}

	c := colly.NewCollector(
		colly.Async(),
	)
	if len(inputStruct.AllowedDomains) > 0 {
		c.AllowedDomains = inputStruct.AllowedDomains
	}
	c.AllowURLRevisit = false

	// On every a element which has href attribute call callback
	// Wont be called if error occurs
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		err := c.Visit(e.Request.AbsoluteURL(link))
		if err != nil {
			log.Println("Error visiting link:", link, "Error:", err)
		}
	})

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	c.OnRequest(func(r *colly.Request) {

		if inputStruct.MaxK > 0 && len(output.Pages) >= inputStruct.MaxK {
			r.Abort()
			return
		}

		// Set a random user agent to avoid being blocked by websites
		r.Headers.Set("User-Agent", randomString())
		// Strip query parameters and trailing '/' from the URL
		strippedURL := stripQueryAndTrailingSlash(r.URL)
		// Check if the URL already exists in the slice
		if !existsInSlice(pageLinks, strippedURL.String()) {
			// Add the URL to the slice if it doesn't already exist
			pageLinks = append(pageLinks, strippedURL.String())
			// Scrape the webpage information
			doc, err := getDocAfterRequestURL(strippedURL.String(), inputStruct.Timeout)
			if err != nil {
				fmt.Printf("Error parsing %s: %v", strippedURL.String(), err)
				return
			}
			page := PageInfo{}
			title := util.ScrapeWebpageTitle(doc)
			page.Title = title
			page.Link = strippedURL.String()

			if *inputStruct.IncludeLinkHTML || *inputStruct.IncludeLinkText {
				html, err := util.ScrapeWebpageHTML(doc)
				if err != nil {
					fmt.Printf("Error scraping HTML from %s: %v", strippedURL.String(), err)
					return
				}

				if *inputStruct.IncludeLinkHTML {
					page.LinkHTML = html
				}

				if *inputStruct.IncludeLinkText {
					domain, err := util.GetDomainFromURL(strippedURL.String())
					if err != nil {
						fmt.Printf("Error getting domain from %s: %v", strippedURL.String(), err)
						return
					}
					markdown, err := util.ScrapeWebpageHTMLToMarkdown(html, domain)
					if err != nil {
						fmt.Printf("Error scraping text from %s: %v", strippedURL.String(), err)
						return
					}
					page.LinkText = markdown
				}
			}
			output.Pages = append(output.Pages, page)
		}
	})

	// Start scraping
	if !strings.HasPrefix(inputStruct.TargetURL, "http://") && !strings.HasPrefix(inputStruct.TargetURL, "https://") {
		inputStruct.TargetURL = "https://" + inputStruct.TargetURL
	}
	_ = c.Visit(inputStruct.TargetURL)
	c.Wait()

	outputStruct, err := base.ConvertToStructpb(output)
	if err != nil {
		return nil, fmt.Errorf("error converting output to struct: %v", err)
	}

	return outputStruct, nil

}
