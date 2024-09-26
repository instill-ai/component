package document

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/jpeg"
	"strings"

	"github.com/gen2brain/go-fitz"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/util"
)

type ConvertDocumentToImagesInput struct {
	Document string `json:"document"`
	Filename string `json:"filename"`
}

type ConvertDocumentToImagesOutput struct {
	Images    []string `json:"images"`
	Filenames []string `json:"filenames"`
}

func ConvertDocumentToImage(inputStruct *ConvertDocumentToImagesInput) (*ConvertDocumentToImagesOutput, error) {

	contentType, err := util.GetContentTypeFromBase64(inputStruct.Document)
	if err != nil {
		return nil, err
	}

	fileExtension := util.TransformContentTypeToFileExtension(contentType)

	if fileExtension == "" {
		return nil, fmt.Errorf("unsupported file type")
	}

	var base64PDF string
	if fileExtension != "pdf" {
		base64PDF, err = ConvertToPDF(inputStruct.Document, fileExtension)

		if err != nil {
			return nil, fmt.Errorf("failed to encode file to base64: %w", err)
		}
	} else {
		base64PDF = strings.Split(inputStruct.Document, ",")[1]
	}

	fileContent, err := base64.StdEncoding.DecodeString(base64PDF)

	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 string: %w", err)
	}

	pdfToBeConverted, err := fitz.NewFromMemory(fileContent)

	if err != nil {
		return nil, fmt.Errorf("failed to create PDF from memory: %w", err)
	}

	defer pdfToBeConverted.Close()

	images := make([]string, pdfToBeConverted.NumPage())
	filenames := make([]string, pdfToBeConverted.NumPage())

	for n := 0; n < pdfToBeConverted.NumPage(); n++ {
		img, err := pdfToBeConverted.Image(n)
		if err != nil {
			return nil, fmt.Errorf("failed to extract image from PDF: %w", err)
		}

		var buf bytes.Buffer
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: jpeg.DefaultQuality})

		if err != nil {
			return nil, fmt.Errorf("failed to encode image to JPEG: %w", err)
		}

		imgBase64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
		images[n] = fmt.Sprintf("data:image/jpeg;base64,%s", imgBase64Str)

		filename := strings.Split(inputStruct.Filename, ".")[0]
		filenames[n] = fmt.Sprintf("%s_%d.jpg", filename, n)
	}

	outputStruct := ConvertDocumentToImagesOutput{
		Images:    images,
		Filenames: filenames,
	}
	return &outputStruct, nil
}

func (e *execution) convertDocumentToImages(input *structpb.Struct) (*structpb.Struct, error) {

	inputStruct := ConvertDocumentToImagesInput{}
	err := base.ConvertFromStructpb(input, &inputStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to convert input struct: %w", err)
	}

	outputStruct, err := ConvertDocumentToImage(&inputStruct)
	if err != nil {
		return nil, err
	}

	return base.ConvertToStructpb(outputStruct)

}
