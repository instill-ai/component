package util

import (
	"encoding/base64"
	"mime/multipart"
	"net/http"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/h2non/filetype"
	"github.com/instill-ai/component/base"
)

func GetFileExt(fileData []byte) string {
	kind, _ := filetype.Match(fileData)
	if kind != filetype.Unknown && kind.Extension != "" {
		return kind.Extension
	}
	//fallback to DetectContentType
	mimeType := http.DetectContentType(fileData)
	return mimeType[strings.LastIndex(mimeType, "/")+1:]
}

func WriteFile(writer *multipart.Writer, fileName string, fileData []byte) error {
	part, err := writer.CreateFormFile(fileName, "file."+GetFileExt(fileData))
	if err != nil {
		return err
	}
	_, err = part.Write(fileData)
	return err
}

func WriteField(writer *multipart.Writer, key string, value string) {
	if key != "" && value != "" {
		_ = writer.WriteField(key, value)
	}
}

// ScrapeWebpageHTML scrape the HTML content of a webpage
func ScrapeWebpageHTML(doc *goquery.Document) (string, error) {
	return doc.Selection.Html()
}

// ScrapeWebpageTitle extracts and returns the title from the *goquery.Document
func ScrapeWebpageTitle(doc *goquery.Document) string {
	// Find the title tag and get its text content
	title := doc.Find("title").Text()

	// Return the trimmed title
	return strings.TrimSpace(title)
}

// ScrapeWebpageHTMLToMarkdown converts an HTML string to Markdown format
func ScrapeWebpageHTMLToMarkdown(html string) (string, error) {
	// Initialize the markdown converter
	converter := md.NewConverter("", true, nil)

	// Convert the HTML to Markdown
	markdown, err := converter.ConvertString(html)
	if err != nil {
		return "", err
	}

	return markdown, nil
}

// DecodeBase64 takes a base64-encoded blob, trims the MIME type (if present)
// and decodes the remaining bytes.
func DecodeBase64(input string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(base.TrimBase64Mime(input))
}
