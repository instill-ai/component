package artifact

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"code.sajari.com/docconv"
	"github.com/frankban/quicktest"
	"github.com/gojuno/minimock/v3"
	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/mock"
	artifactPB "github.com/instill-ai/protogen-go/artifact/artifact/v1alpha"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func Test_uploadFiles(t *testing.T) {
	c := quicktest.New(t)
	mc := minimock.NewController(t)

	testCases := []struct {
		name     string
		fileName string
		option   string
		expected string
	}{
		{
			name:     "upload file with new catalog",
			fileName: "testdata/test.pdf",
			option:   "create new catalog",
			expected: "testdata/test.pdf",
		},
		{
			name:     "upload file with existing catalog",
			fileName: "testdata/test.pdf",
			option:   "existing catalog",
			expected: "testdata/test.pdf",
		},
	}

	for _, tc := range testCases {
		c.Run(tc.name, func(c *quicktest.C) {
			component := Init(base.Component{})

			sysVar := map[string]interface{}{
				"__ARTIFACT_BACKEND":       "http://localhost:8082",
				"__PIPELINE_USER_UID":      "fakeUser",
				"__PIPELINE_REQUESTER_UID": "fakeRequester",
			}
			e := &execution{
				ComponentExecution: base.ComponentExecution{Component: component, SystemVariables: sysVar, Setup: nil, Task: taskUploadFile},
			}

			e.execute = e.uploadFiles

			fileContent, _ := os.ReadFile(tc.fileName)
			base64DataURI := fmt.Sprintf("data:%s;base64,%s", docconv.MimeTypeByExtension(tc.fileName), base64.StdEncoding.EncodeToString(fileContent))

			input := UploadFilesInput{
				Options: UploadData{
					Option:    tc.option,
					Namespace: "fakeNs",
					CatalogID: "fakeID",
					File:      base64DataURI,
					FileName:  tc.fileName,
				},
			}
			inputStruct, _ := base.ConvertToStructpb(input)

			clientMock := mock.NewArtifactPublicServiceClientMock(mc)
			if tc.option == "create new catalog" {
				clientMock.
					CreateCatalogMock.
					Times(1).
					Expect(minimock.AnyContext,
						&artifactPB.CreateCatalogRequest{
							NamespaceId: "fakeNs",
							Name:        "fakeID",
						}).
					Return(nil, nil)
			}
			clientMock.UploadCatalogFileMock.Times(1).
				Expect(minimock.AnyContext,
					&artifactPB.UploadCatalogFileRequest{
						NamespaceId: "fakeNs",
						CatalogId:   "fakeID",
						File: &artifactPB.File{
							Name:    tc.fileName,
							Type:    artifactPB.FileType_FILE_TYPE_PDF,
							Content: base64.StdEncoding.EncodeToString(fileContent),
						},
					},
				).
				Return(&artifactPB.UploadCatalogFileResponse{
					File: &artifactPB.File{
						FileUid: "fakeFileID",
						Name:    tc.fileName,
						Type:    artifactPB.FileType_FILE_TYPE_PDF,
						Size:    1,
						CreateTime: &timestamppb.Timestamp{
							Seconds: 1,
							Nanos:   1,
						},
						UpdateTime: &timestamppb.Timestamp{
							Seconds: 1,
							Nanos:   1,
						},
					},
				}, nil)

			clientMock.ProcessCatalogFilesMock.
				Expect(minimock.AnyContext, &artifactPB.ProcessCatalogFilesRequest{
					FileUids: []string{"fakeFileID"},
				}).
				Times(1).
				Return(nil, nil)

			e.client = clientMock
			e.connection = fakeConnection{}

			output, err := e.execute(inputStruct)

			c.Assert(err, quicktest.IsNil)

			var outputStruct UploadFilesOutput
			err = base.ConvertFromStructpb(output, &outputStruct)

			c.Assert(err, quicktest.IsNil)

			c.Assert(outputStruct.File.FileUID, quicktest.Equals, "fakeFileID")
			c.Assert(outputStruct.File.FileName, quicktest.Equals, tc.fileName)
			c.Assert(outputStruct.File.Size, quicktest.Equals, int64(1))
			c.Assert(outputStruct.File.CreateTime, quicktest.Equals, "1970-01-01T00:00:01Z")
			c.Assert(outputStruct.File.UpdateTime, quicktest.Equals, "1970-01-01T00:00:01Z")
			c.Assert(outputStruct.File.CatalogID, quicktest.Equals, "fakeID")

		})
	}

}

func Test_getFilesMetadata(t *testing.T) {
}

func Test_getChunksMetadata(t *testing.T) {

}

func Test_getFileInMarkdown(t *testing.T) {}

func Test_matchFileStatus(t *testing.T) {}

func Test_searchChunks(t *testing.T) {}

func Test_query(t *testing.T) {}

type fakeConnection struct{}

func (f fakeConnection) Close() error {
	return nil
}
