package document

import (
	"fmt"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/util"
	"google.golang.org/protobuf/types/known/structpb"
)

type convertDocumentToMarkdownInput struct {
	Document        string `json:"document"`
	DisplayImageTag bool   `json:"display-image-tag"`
}

type convertDocumentToMarkdownOutput struct {
	Body string `json:"body"`
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
		}, nil
	case "doc", "docx":
		return DocxDocToMarkdownTransformer{
			Base64EncodedText: inputStruct.Document,
			FileExtension:     fileExtension,
			DisplayImageTag:   inputStruct.DisplayImageTag,
		}, nil
	case "ppt", "pptx":
		return PptPptxToMarkdownTransformer{
			Base64EncodedText: inputStruct.Document,
			FileExtension:     fileExtension,
			DisplayImageTag:   inputStruct.DisplayImageTag,
		}, nil
	case "html":
		return HTMLToMarkdownTransformer{
			Base64EncodedText: inputStruct.Document,
			FileExtension:     fileExtension,
			DisplayImageTag:   inputStruct.DisplayImageTag,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported file type")
	}
}
