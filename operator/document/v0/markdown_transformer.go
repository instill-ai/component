package document

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/util"
	"github.com/xuri/excelize/v2"
)

type MarkdownTransformer interface {
	Transform() (string, error)
}

type PDFToMarkdownTransformer struct {
	Base64EncodedText string
	FileExtension     string
	DisplayImageTag   bool
	PDFConvertFunc    func(string, bool) (string, error)
}

func (t PDFToMarkdownTransformer) Transform() (string, error) {
	return t.PDFConvertFunc(t.Base64EncodedText, t.DisplayImageTag)
}

type DocxDocToMarkdownTransformer struct {
	Base64EncodedText string
	FileExtension     string
	DisplayImageTag   bool
	PDFConvertFunc    func(string, bool) (string, error)
}

func (t DocxDocToMarkdownTransformer) Transform() (string, error) {

	base64PDF, err := ConvertToPDF(t.Base64EncodedText, t.FileExtension)

	if err != nil {
		return "", fmt.Errorf("failed to encode file to base64: %w", err)
	}

	return t.PDFConvertFunc(base64PDF, t.DisplayImageTag)
}

type PptPptxToMarkdownTransformer struct {
	Base64EncodedText string
	FileExtension     string
	DisplayImageTag   bool
	PDFConvertFunc    func(string, bool) (string, error)
}

func (t PptPptxToMarkdownTransformer) Transform() (string, error) {

	base64PDF, err := ConvertToPDF(t.Base64EncodedText, t.FileExtension)

	if err != nil {
		return "", fmt.Errorf("failed to encode file to base64: %w", err)
	}

	return t.PDFConvertFunc(base64PDF, t.DisplayImageTag)
}

type HTMLToMarkdownTransformer struct {
	Base64EncodedText string
	FileExtension     string
	DisplayImageTag   bool
}

func (t HTMLToMarkdownTransformer) Transform() (string, error) {

	data, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(t.Base64EncodedText))
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 to file: %w", err)
	}

	converter := md.NewConverter("", true, nil)

	html := string(data)
	markdown, err := converter.ConvertString(html)
	if err != nil {
		return "", fmt.Errorf("failed to convert HTML to markdown: %w", err)
	}

	return markdown, nil
}

type XlsxToMarkdownTransformer struct {
	Base64EncodedText string
}

func (t XlsxToMarkdownTransformer) Transform() (string, error) {

	base64String := strings.Split(t.Base64EncodedText, ",")[1]
	fileContent, err := base64.StdEncoding.DecodeString(base64String)

	if err != nil {
		return "", fmt.Errorf("failed to decode base64 to file: %w", err)
	}

	reader := bytes.NewReader(fileContent)

	f, err := excelize.OpenReader(reader)
	if err != nil {
		return "", fmt.Errorf("failed to open reader: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()

	var result string
	for _, sheet := range sheets {
		rows, err := f.GetRows(sheet)

		if err != nil {
			return "", fmt.Errorf("failed to get rows: %w", err)
		}

		if len(rows) == 0 {
			result += fmt.Sprintf("# %s\n", sheet)
			result += "No data found\n\n"
			continue
		}

		result += fmt.Sprintf("# %s\n", sheet)
		result += util.ConvertDataFrameToMarkdownTable(rows)
		result += "\n\n"
	}

	return result, nil
}

type pythonRunnerOutput struct {
	Body string `json:"body"`
}

func writeDecodeToFile(base64Str string, file *os.File) error {
	data, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(base64Str))
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	return err
}

func encodeFileToBase64(inputPath string) (string, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func ConvertToPDF(base64Encoded, fileExtension string) (string, error) {
	tempPpt, err := os.CreateTemp("", "temp_document.*."+fileExtension)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary document: %w", err)
	}
	inputFileName := tempPpt.Name()
	defer os.Remove(inputFileName)

	err = writeDecodeToFile(base64Encoded, tempPpt)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 to file: %w", err)
	}

	tempDir, err := os.MkdirTemp("", "libreoffice")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: " + err.Error())
	}
	defer os.RemoveAll(tempDir)

	cmd := exec.Command("libreoffice", "--headless", "--convert-to", "pdf", inputFileName)
	cmd.Env = append(os.Environ(), "HOME="+tempDir)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to execute LibreOffice command: " + err.Error())
	}

	// LibreOffice is not executed in temp directory like inputFileName.
	// The generated PDF is not in temp directory.
	// So, we need to remove the path and keep only the file name.
	noPathFileName := filepath.Base(inputFileName)
	tempPDFName := strings.TrimSuffix(noPathFileName, filepath.Ext(inputFileName)) + ".pdf"
	defer os.Remove(tempPDFName)

	base64PDF, err := encodeFileToBase64(tempPDFName)

	if err != nil {
		return "", fmt.Errorf("failed to encode file to base64: %w", err)
	}
	return base64PDF, nil
}
