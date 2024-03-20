package text

import (
	"fmt"
	"os"
	"testing"

	"encoding/base64"

	"code.sajari.com/docconv"
	"google.golang.org/protobuf/types/known/structpb"
)

// TestConvertToText tests the convert to text task
func TestConvertToText(t *testing.T) {

	tests := []struct {
		name     string
		filepath string
	}{
		{
			name:     "Convert pdf file",
			filepath: "testdata/test.pdf",
		},
		{
			name:     "Convert docx file",
			filepath: "testdata/test.docx",
		},
		{
			name:     "Convert html file",
			filepath: "testdata/test.html",
		},
		{
			name:     "Convert odt file",
			filepath: "testdata/test.odt",
		},
		{
			name:     "Convert rtf file",
			filepath: "testdata/test.rtf",
		},
		{
			name:     "Convert png file",
			filepath: "testdata/test.png",
		},
		{
			name:     "Convert jpg file",
			filepath: "testdata/test.jpg",
		},
		{
			name:     "Convert tiff file",
			filepath: "testdata/test.tif",
		},
		{
			name:     "Convert txt file",
			filepath: "testdata/test.txt",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Read the fileContent content
			fileContent, err := os.ReadFile(test.filepath)
			if err != nil {
				t.Fatalf("error reading the file: %s\n", err)
				return
			}

			base64DataURI := fmt.Sprintf("data:%s;base64,%s", docconv.MimeTypeByExtension(test.filepath), base64.StdEncoding.EncodeToString(fileContent))

			input := &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"doc": {Kind: &structpb.Value_StringValue{StringValue: base64DataURI}},
				},
			}
			inputs := []*structpb.Struct{
				input,
			}

			e := &Execution{}
			e.Task = "TASK_CONVERT_TO_TEXT"

			if outputs, err := e.Execute(inputs); err != nil {
				t.Fatalf("convertToText returned an error: %v", err)
			} else if outputs[0].Fields["body"].GetStringValue() == "" {
				t.Fatal("convertToText returned an empty body")
			} else if outputs[0].Fields["meta"].GetStructValue() == nil {
				t.Fatal("convertToText returned a nil meta")
			}

		})
	}

}
