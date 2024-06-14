package document

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/gojuno/minimock/v3"
	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/mock"
)

func TestConvertPDFToText(t *testing.T) {
	c := qt.New(t)

	c.Run("Convert PDF file", func(c *qt.C) {

		fakePDF := "# Test\n\nThis is a test document.\n\n"
		b, err := json.Marshal(fakePDF)
		c.Assert(err, qt.IsNil)

		encoded := base64.StdEncoding.EncodeToString(b)

		input := convertPDFToMarkdownInput{
			PDF:             encoded,
			DisplayImageTag: false,
		}

		mc := minimock.NewController(t)
		mockRunner := mock.NewCommandRunnerMock(mc)
		mockIoWriteCloser := mock.NewWriteCloserMock(mc)
		mockIoWriteCloser.CloseMock.Expect().Return(nil)

		fakeParams, err := json.Marshal(map[string]interface{}{
			"PDF":               base.TrimBase64Mime(encoded),
			"display-image-tag": false,
		})

		c.Assert(err, qt.IsNil)

		mockIoWriteCloser.WriteMock.Expect(fakeParams).Return(len(fakePDF), nil)
		mockRunner.StdinPipeMock.Expect().Return(mockIoWriteCloser, nil)

		mockOutput := convertPDFToMarkdownOutput{
			Body: "# Test\n\nThis is a test document.\n\n",
		}
		mockOutputBytes, err := json.Marshal(mockOutput)
		c.Assert(err, qt.IsNil)

		mockRunner.CombinedOutputMock.Expect().Return(mockOutputBytes, nil)

		output, err := convertPDFToMarkdown(input, mockRunner)
		c.Assert(err, qt.IsNil)

		c.Assert(output.Body, qt.Equals, "# Test\n\nThis is a test document.\n\n")

	})

}
