package document

import (
	"fmt"
	"strings"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/util"
	"google.golang.org/protobuf/types/known/structpb"
)

type convertDocumentToMarkdownInput struct {
	Document        string `json:"document"`
	DisplayImageTag bool   `json:"display-image-tag"`
	Converter       string `json:"converter"`
	Filename        string `json:"filename"`
}

type convertDocumentToMarkdownOutput struct {
	Body     string `json:"body"`
	Filename string `json:"filename"`
}

func (e *execution) convertDocumentToMarkdown(input *structpb.Struct) (*structpb.Struct, error) {
	inputStruct := convertDocumentToMarkdownInput{}
	err := base.ConvertFromStructpb(input, &inputStruct)
	if err != nil {
		return nil, err
	}

	contentType, err := util.GetContentTypeFromBase64(inputStruct.Document)
	if err != nil {
		return nil, err
	}

	fileExtension := util.TransformContentTypeToFileExtension(contentType)

	if fileExtension == "" {
		return nil, fmt.Errorf("unsupported file type")
	}

	var transformer MarkdownTransformer

	transformer, err = e.getMarkdownTransformer(fileExtension, inputStruct)
	if err != nil {
		return nil, err
	}
	extractedTextInMarkdownFormat, err := transformer.Transform()
	if err != nil {
		return nil, err
	}

	outputStruct := convertDocumentToMarkdownOutput{
		Body: extractedTextInMarkdownFormat,
	}

	if inputStruct.Filename != "" {
		filename := strings.Split(inputStruct.Filename, ".")[0] + ".md"
		outputStruct.Filename = filename
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func getMarkdownTransformer(fileExtension string, inputStruct convertDocumentToMarkdownInput) (MarkdownTransformer, error) {
	switch fileExtension {
	case "pdf":
		return PDFToMarkdownTransformer{
			Base64EncodedText: inputStruct.Document,
			FileExtension:     fileExtension,
			DisplayImageTag:   inputStruct.DisplayImageTag,
			Converter:         inputStruct.Converter,
		}, nil
	case "doc", "docx":
		return DocxDocToMarkdownTransformer{
			Base64EncodedText: inputStruct.Document,
			FileExtension:     fileExtension,
			DisplayImageTag:   inputStruct.DisplayImageTag,
			Converter:         inputStruct.Converter,
		}, nil
	case "ppt", "pptx":
		return PptPptxToMarkdownTransformer{
			Base64EncodedText: inputStruct.Document,
			FileExtension:     fileExtension,
			DisplayImageTag:   inputStruct.DisplayImageTag,
			Converter:         inputStruct.Converter,
		}, nil
	case "html":
		return HTMLToMarkdownTransformer{
			Base64EncodedText: inputStruct.Document,
			FileExtension:     fileExtension,
			DisplayImageTag:   inputStruct.DisplayImageTag,
			Converter:         inputStruct.Converter,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported file type")
	}
}
