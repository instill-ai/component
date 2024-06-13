package website

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/instill-ai/component/internal/util"
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
	IncludeLinkText *bool `json:"include_link_text"`
	// IncludeLinkHTML: Whether to include the scraped HTML of the scraped web page.
	IncludeLinkHTML *bool `json:"include_link_html"`
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

// getHTMLPageDoc returns the *goquery.Document of a webpage
func getHTMLPageDoc(url string) (*goquery.Document, error) {
	// Request the HTML page.
	client := &http.Client{Transport: &http.Transport{
		DisableKeepAlives: true,
	}}
	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

// Scrape crawls a webpage and returns a slice of PageInfo
func Scrape(input ScrapeWebsiteInput) (ScrapeWebsiteOutput, error) {
	output := ScrapeWebsiteOutput{}

	if input.IncludeLinkHTML == nil {
		b := false
		input.IncludeLinkHTML = &b
	}
	if input.IncludeLinkText == nil {
		b := false
		input.IncludeLinkText = &b
	}
	if input.MaxK < 0 {
		input.MaxK = 0
	}

	pageLinks := []string{}

	c := colly.NewCollector()
	if len(input.AllowedDomains) > 0 {
		c.AllowedDomains = input.AllowedDomains
	}
	c.AllowURLRevisit = false

	// On every a element which has href attribute call callback
	// Wont be called if error occurs
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		_ = c.Visit(e.Request.AbsoluteURL(link))
	})

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	c.OnRequest(func(r *colly.Request) {

		if input.MaxK > 0 && len(output.Pages) >= input.MaxK {
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
			doc, err := getHTMLPageDoc(strippedURL.String())
			if err != nil {
				fmt.Printf("Error parsing %s: %v", strippedURL.String(), err)
				return
			}
			page := PageInfo{}
			title := util.ScrapeWebpageTitle(doc)
			page.Title = title
			page.Link = strippedURL.String()

			if *input.IncludeLinkHTML || *input.IncludeLinkText {
				html, err := util.ScrapeWebpageHTML(doc)
				if err != nil {
					fmt.Printf("Error scraping HTML from %s: %v", strippedURL.String(), err)
					return
				}

				if *input.IncludeLinkHTML {
					page.LinkHTML = html
				}

				if *input.IncludeLinkText {
					markdown, err := util.ScrapeWebpageHTMLToMarkdown(html)
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
	if !strings.HasPrefix(input.TargetURL, "http://") && !strings.HasPrefix(input.TargetURL, "https://") {
		input.TargetURL = "https://" + input.TargetURL
	}
	_ = c.Visit(input.TargetURL)

	return output, nil
}
