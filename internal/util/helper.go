package util

import (
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
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
func ScrapeWebpageHTMLToMarkdown(html, domain string) (string, error) {
	// Initialize the markdown converter
	converter := md.NewConverter(domain, true, nil)

	// Convert the HTML to Markdown
	markdown, err := converter.ConvertString(html)
	if err != nil {
		return "", err
	}

	return markdown, nil
}

func GetDomainFromURL(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)

	if err != nil {
		return "", fmt.Errorf("error when parse url: %v", err)
	}
	return u.Host, nil
}

// DecodeBase64 takes a base64-encoded blob, trims the MIME type (if present)
// and decodes the remaining bytes.
func DecodeBase64(input string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(base.TrimBase64Mime(input))
}

func GetContentTypeFromBase64(base64String string) (string, error) {
	// Remove the "data:" prefix and split at the first semicolon
	contentType := strings.TrimPrefix(base64String, "data:")

	parts := strings.SplitN(contentType, ";", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid format")
	}

	// The first part is the content type
	return parts[0], nil
}

func GetFileBase64Content(base64String string) string {
	return strings.SplitN(base64String, ",", 2)[1]
}

func TransformContentTypeToFileExtension(contentType string) string {
	// https://gist.github.com/AshHeskes/6038140
	// We can integrate more Content-Type to file extension mappings in the future
	switch contentType {
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return "docx"
	case "application/msword":
		return "doc"
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		return "pptx"
	case "application/vnd.ms-powerpoint":
		return "ppt"
	case "text/html":
		return "html"
	case "application/pdf":
		return "pdf"
	}
	return ""
}

func StripProtocolFromURL(url string) string {
	index := strings.Index(url, "://")
	if index > 0 {
		return url[strings.Index(url, "://")+3:]
	}
	return url
}

func GetHeaderAuthorization(vars map[string]any) string {
	if v, ok := vars["__PIPELINE_HEADER_AUTHORIZATION"]; ok {
		return v.(string)
	}
	return ""
}
func GetInstillUserUID(vars map[string]any) string {
	return vars["__PIPELINE_USER_UID"].(string)
}

func GetInstillRequesterUID(vars map[string]any) string {
	return vars["__PIPELINE_REQUESTER_UID"].(string)
}
