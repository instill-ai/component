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
)

type ConvertPDFToImagesInput struct {
	PDF      string `json:"pdf"`
	Filename string `json:"filename"`
}

type ConvertPDFToImagesOutput struct {
	Images    []string `json:"images"`
	Filenames []string `json:"filenames"`
}

func ConvertPDFToImage(inputStruct *ConvertPDFToImagesInput) (*ConvertPDFToImagesOutput, error) {
	base64String := strings.Split(inputStruct.PDF, ",")[1]
	fileContent, err := base64.StdEncoding.DecodeString(base64String)

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

	outputStruct := ConvertPDFToImagesOutput{
		Images:    images,
		Filenames: filenames,
	}
	return &outputStruct, nil
}

func (e *execution) convertPDFToImages(input *structpb.Struct) (*structpb.Struct, error) {

	inputStruct := ConvertPDFToImagesInput{}
	err := base.ConvertFromStructpb(input, &inputStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to convert input struct: %w", err)
	}

	outputStruct, err := ConvertPDFToImage(&inputStruct)
	if err != nil {
		return nil, err
	}

	return base.ConvertToStructpb(outputStruct)

}
