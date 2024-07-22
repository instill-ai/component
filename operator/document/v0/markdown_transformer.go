package document

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/instill-ai/component/base"
)

type MarkdownTransformer interface {
	Transform() (string, error)
}

type PDFToMarkdownTransformer struct {
	Base64EncodedText string
	FileExtension     string
	DisplayImageTag   bool
}

func (t PDFToMarkdownTransformer) Transform() (string, error) {
	return extractPDFTextInMarkdownFormat(t.Base64EncodedText, t.DisplayImageTag)
}

type DocxDocToMarkdownTransformer struct {
	Base64EncodedText string
	FileExtension     string
	DisplayImageTag   bool
}

func (t DocxDocToMarkdownTransformer) Transform() (string, error) {

	inputTempDecodeFileName := "temp_document." + t.FileExtension
	tempPDFName := "temp_document.pdf"
	defer os.Remove(inputTempDecodeFileName)
	defer os.Remove(tempPDFName)

	err := decodeBase64ToFile(t.Base64EncodedText, inputTempDecodeFileName)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 to file: %w", err)
	}

	err = convertToPDF(inputTempDecodeFileName, tempPDFName)
	if err != nil {
		return "", fmt.Errorf("failed to convert file to PDF: %w", err)
	}

	base64PDF, err := encodeFileToBase64(tempPDFName)

	if err != nil {
		return "", fmt.Errorf("failed to encode file to base64: %w", err)
	}

	return extractPDFTextInMarkdownFormat(base64PDF, t.DisplayImageTag)
}

type PptPptxToMarkdownTransformer struct {
	Base64EncodedText string
	FileExtension     string
	DisplayImageTag   bool
}

func (t PptPptxToMarkdownTransformer) Transform() (string, error) {
	inputTempDecodeFileName := "temp_document." + t.FileExtension
	tempPDFName := "temp_document.pdf"
	defer os.Remove(inputTempDecodeFileName)
	defer os.Remove(tempPDFName)

	err := decodeBase64ToFile(t.Base64EncodedText, inputTempDecodeFileName)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 to file: %w", err)
	}

	err = convertToPDF(inputTempDecodeFileName, tempPDFName)
	if err != nil {
		return "", fmt.Errorf("failed to convert file to PDF: %w", err)
	}

	base64PDF, err := encodeFileToBase64(tempPDFName)

	if err != nil {
		return "", fmt.Errorf("failed to encode file to base64: %w", err)
	}

	return extractPDFTextInMarkdownFormat(base64PDF, t.DisplayImageTag)
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

type pythonRunnerOutput struct {
	Body string `json:"body"`
}

func extractPDFTextInMarkdownFormat(base64Text string, displayImageTag bool) (string, error) {

	paramsJSON, err := json.Marshal(map[string]interface{}{
		"PDF":               base.TrimBase64Mime(base64Text),
		"display-image-tag": displayImageTag,
	})

	if err != nil {
		return "", err
	}

	cmdRunner := exec.Command(pythonInterpreter, "-c", pythonCode)
	stdin, err := cmdRunner.StdinPipe()

	if err != nil {
		return "", err
	}
	errChan := make(chan error, 1)
	go func() {
		defer stdin.Close()
		_, err := stdin.Write(paramsJSON)
		if err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	outputBytes, err := cmdRunner.CombinedOutput()
	if err != nil {
		return "", err
	}

	writeErr := <-errChan
	if writeErr != nil {
		return "", writeErr
	}

	var output pythonRunnerOutput
	err = json.Unmarshal(outputBytes, &output)
	if err != nil {
		return "", err
	}
	return output.Body, nil
}

func decodeBase64ToFile(base64Str, tempPDFPath string) error {
	data, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(base64Str))
	if err != nil {
		return err
	}
	return os.WriteFile(tempPDFPath, data, 0644)
}

func encodeFileToBase64(inputPath string) (string, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func convertToPDF(inputPath, outputPath string) error {
	tempDir, err := os.MkdirTemp("", "libreoffice")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: " + err.Error())
	}
	defer os.RemoveAll(tempDir)

	cmd := exec.Command("libreoffice", "--headless", "--convert-to", "pdf", "--outdir", "./", inputPath)
	cmd.Env = append(os.Environ(), "HOME="+tempDir)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute LibreOffice command: " + err.Error())
	}

	generatedPDF := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".pdf"
	if err := os.Rename(generatedPDF, outputPath); err != nil {
		return fmt.Errorf("failed to rename output file: " + err.Error())
	}
	return nil
}
