package text

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"encoding/base64"

	"code.sajari.com/docconv"
	"gopkg.in/yaml.v2"

	"github.com/instill-ai/component/base"
)

// ConvertToTextInput defines the input for convert to text task
type ConvertToTextInput struct {
	// Doc: Document to convert
	Doc string `json:"doc"`
}

// ConvertToTextOutput defines the output for convert to text task
type ConvertToTextOutput struct {
	// Body: Plain text converted from the document
	Body string `json:"body"`
	// Meta: Metadata extracted from the document
	Meta map[string]string `json:"meta"`
	// MSecs: Time taken to convert the document
	MSecs uint32 `json:"msecs"`
	// Error: Error message if any during the conversion process
	Error string `json:"error"`
}

func getContentTypeFromBase64(base64String string) (string, error) {
	// Remove the "data:" prefix and split at the first semicolon
	contentType := strings.TrimPrefix(base64String, "data:")

	parts := strings.SplitN(contentType, ";", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid format")
	}

	// The first part is the content type
	return parts[0], nil
}

type converter interface {
	convert(contentType string, b []byte) (ConvertToTextOutput, error)
}

type docconvConverter struct{}

func (d docconvConverter) convert(contentType string, b []byte) (ConvertToTextOutput, error) {

	res, err := docconv.Convert(bytes.NewReader(b), contentType, false)
	if err != nil {
		return ConvertToTextOutput{}, err
	}

	if res.Meta == nil {
		res.Meta = map[string]string{}
	}

	return ConvertToTextOutput{
		Body:  res.Body,
		Meta:  res.Meta,
		MSecs: res.MSecs,
		Error: res.Error,
	}, nil
}

type markdownConverter struct{}

func extractMetadata(content string) map[string]string {
	var metadata map[string]string

	if !strings.HasPrefix(content, "---") {
		return metadata
	}

	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return metadata
	}

	yamlContent := parts[1]
	err := yaml.Unmarshal([]byte(yamlContent), &metadata)
	if err != nil {
		return metadata
	}

	return metadata
}

func (m markdownConverter) convert(contentType string, b []byte) (ConvertToTextOutput, error) {

	before := time.Now()
	content := string(b)
	metadata := extractMetadata(content)

	duration := time.Since(before)
	millis := duration.Milliseconds()

	if metadata == nil {
		metadata = map[string]string{}
	}

	return ConvertToTextOutput{
		Body:  content,
		Meta:  metadata,
		MSecs: uint32(millis),
		Error: "",
	}, nil
}

func convertToText(input ConvertToTextInput) (ConvertToTextOutput, error) {

	contentType, err := getContentTypeFromBase64(input.Doc)
	if err != nil {
		return ConvertToTextOutput{}, err
	}

	b, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(input.Doc))
	if err != nil {
		return ConvertToTextOutput{}, err
	}

	var converter converter
	switch contentType {
	case "application/octet-stream":
		converter = markdownConverter{}
	default:
		converter = docconvConverter{}
	}

	res, err := converter.convert(contentType, b)
	if err != nil {
		return ConvertToTextOutput{}, err
	}

	return res, nil
}