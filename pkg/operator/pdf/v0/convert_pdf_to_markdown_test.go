package pdf

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/gojuno/minimock/v3"
	"github.com/instill-ai/component/pkg/mock"
)

func TestConvertPDFToText(t *testing.T) {
	c := qt.New(t)
	test := struct {
		name     string
		filepath string
	}{
		name:     "Convert PDF file",
		filepath: "testdata/test.pdf",
	}

	c.Run(test.name, func(c *qt.C) {
		fakePDF := "# Test\n\nThis is a test document.\n\n"
		b, err := json.Marshal(fakePDF)
		c.Assert(err, qt.IsNil)

		encoded := base64.StdEncoding.EncodeToString(b)
		input := convertPDFToMarkdownInput{
			PDF: encoded,
		}

		mc := minimock.NewController(t)
		mockRunner := mock.NewCommandRunnerMock(mc)
		mockIoWriteCloser := mock.NewWriteCloserMock(mc)
		mockIoWriteCloser.CloseMock.Expect().Return(nil)

		mockIoWriteCloser.WriteMock.Expect(b).Return(len(fakePDF), nil)
		mockRunner.StdinPipeMock.Expect().Return(mockIoWriteCloser, nil)

		mockOutput := convertPDFToMarkdownOutput{
			Body: "# Test\n\nThis is a test document.\n\n",
			// TODO: revert it when target the bug.
			// https://linear.app/instill-ai/issue/INS-3098/[component][pdf-operator]-add-task-convert-to-markdown#comment-dc17f0f3
			// Metadata: map[string]string{
			// 	"title": "Test",
			// },
		}
		mockOutputBytes, err := json.Marshal(mockOutput)
		c.Assert(err, qt.IsNil)

		mockRunner.CombinedOutputMock.Expect().Return(mockOutputBytes, nil)

		output, err := convertPDFToMarkdown(input, mockRunner)
		c.Assert(err, qt.IsNil)

		c.Assert(output.Body, qt.Equals, "# Test\n\nThis is a test document.\n\n")

	})

}
