package text

import (
	"bytes"
	"fmt"
	"strings"

	"encoding/base64"

	"code.sajari.com/docconv"

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

func convertToText(input ConvertToTextInput) (ConvertToTextOutput, error) {

	contentType, err := getContentTypeFromBase64(input.Doc)
	if err != nil {
		return ConvertToTextOutput{}, err
	}

	b, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(input.Doc))
	if err != nil {
		return ConvertToTextOutput{}, err
	}

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
