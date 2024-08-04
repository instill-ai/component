package document

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
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
	Converter         string
}

func (t PDFToMarkdownTransformer) Transform() (string, error) {
	if false {
		return extractPDFTextInMarkdownFormatDeprecate(t.Base64EncodedText, t.DisplayImageTag)
	} else {
		return extractPDFTextInMarkdownFormat(t.Base64EncodedText, t.DisplayImageTag)
	}
}

type DocxDocToMarkdownTransformer struct {
	Base64EncodedText string
	FileExtension     string
	DisplayImageTag   bool
	Converter         string
}

func (t DocxDocToMarkdownTransformer) Transform() (string, error) {

	tempDoc, err := os.CreateTemp("", "temp_document.*."+t.FileExtension)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary document: %w", err)
	}
	inputTempDecodeFileName := tempDoc.Name()
	defer os.Remove(inputTempDecodeFileName)

	err = writeDecodeToFile(t.Base64EncodedText, tempDoc)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 to file: %w", err)
	}

	tempPDFName, err := convertToPDF(inputTempDecodeFileName)
	if err != nil {
		return "", fmt.Errorf("failed to convert file to PDF: %w", err)
	}
	defer os.Remove(tempPDFName)

	base64PDF, err := encodeFileToBase64(tempPDFName)

	if err != nil {
		return "", fmt.Errorf("failed to encode file to base64: %w", err)
	}

	if false {
		return extractPDFTextInMarkdownFormatDeprecate(base64PDF, t.DisplayImageTag)
	} else {
		return extractPDFTextInMarkdownFormat(base64PDF, t.DisplayImageTag)
	}
}

type PptPptxToMarkdownTransformer struct {
	Base64EncodedText string
	FileExtension     string
	DisplayImageTag   bool
	Converter         string
}

func (t PptPptxToMarkdownTransformer) Transform() (string, error) {
	tempPpt, err := os.CreateTemp("", "temp_document.*."+t.FileExtension)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary document: %w", err)
	}
	inputTempDecodeFileName := tempPpt.Name()
	defer os.Remove(inputTempDecodeFileName)

	err = writeDecodeToFile(t.Base64EncodedText, tempPpt)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 to file: %w", err)
	}

	tempPDFName, err := convertToPDF(inputTempDecodeFileName)
	if err != nil {
		return "", fmt.Errorf("failed to convert file to PDF: %w", err)
	}
	defer os.Remove(tempPDFName)

	base64PDF, err := encodeFileToBase64(tempPDFName)

	if err != nil {
		return "", fmt.Errorf("failed to encode file to base64: %w", err)
	}

	if false {
		return extractPDFTextInMarkdownFormatDeprecate(base64PDF, t.DisplayImageTag)
	} else {
		return extractPDFTextInMarkdownFormat(base64PDF, t.DisplayImageTag)
	}
}

type HTMLToMarkdownTransformer struct {
	Base64EncodedText string
	FileExtension     string
	DisplayImageTag   bool
	Converter         string
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

func extractPDFTextInMarkdownFormatDeprecate(base64Text string, displayImageTag bool) (string, error) {

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

func extractPDFTextInMarkdownFormat(base64Text string, displayImageTag bool) (string, error) {
	inputDir, err := os.MkdirTemp(os.TempDir(), "pdf")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(inputDir)
	outputDir, err := os.MkdirTemp(os.TempDir(), "pdf")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(outputDir)

	inputBytes, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(base64Text))
	if err != nil {
		return "", err
	}
	err = os.WriteFile(path.Join(inputDir, "document.pdf"), inputBytes, 0644)
	if err != nil {
		return "", err
	}
	cmdRunner := exec.Command("pdf2md", fmt.Sprintf("--inputFolderPath=%s", inputDir), fmt.Sprintf("--outputFolderPath=%s", outputDir))
	err = cmdRunner.Run()
	if err != nil {
		return "", err
	}

	outputBytes, err := os.ReadFile(path.Join(outputDir, "document.md"))
	if err != nil {
		return "", err
	}

	return string(outputBytes), nil
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

func convertToPDF(inputFileName string) (fileName string, err error) {
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
	generatedPDF := strings.TrimSuffix(noPathFileName, filepath.Ext(inputFileName)) + ".pdf"
	return generatedPDF, nil
}
