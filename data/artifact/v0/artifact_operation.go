package artifact

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/component/internal/util"
	artifactPB "github.com/instill-ai/protogen-go/artifact/artifact/v1alpha"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

type UploadFilesInput struct {
	Options UploadData `json:"options"`
}

type UploadData struct {
	Option    string `json:"option"`
	Namespace string `json:"namespace"`
	CatalogID string `json:"catalog-id"`
	// Base64 encoded file content
	File        string   `json:"file"`
	FileName    string   `json:"file-name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

func (input *UploadFilesInput) isNewCatalog() bool {
	return input.Options.Option == "create new catalog"
}

type UploadFilesOutput struct {
	File   FileOutput `json:"file"`
	Status bool       `json:"status"`
}

type FileOutput struct {
	FileUID    string `json:"file-uid"`
	FileName   string `json:"file-name"`
	FileType   string `json:"file-type"`
	CreateTime string `json:"create-time"`
	UpdateTime string `json:"update-time"`
	Size       int64  `json:"size"`
	CatalogID  string `json:"catalog-id"`
}

type Connection interface {
	Close() error
}

func (e *execution) uploadFiles(input *structpb.Struct) (*structpb.Struct, error) {

	inputStruct := UploadFilesInput{}

	err := base.ConvertFromStructpb(input, &inputStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to convert input to struct: %w", err)
	}

	client, connection, err := e.initClient(getArtifactServerURL(e.SystemVariables))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize artifact client: %w", err)
	}
	defer connection.(Connection).Close()

	artifactClient := client.(artifactPB.ArtifactPublicServiceClient)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	ctx = metadata.NewOutgoingContext(ctx, getRequestMetadata(e.SystemVariables))

	if inputStruct.isNewCatalog() {
		_, err = artifactClient.CreateCatalog(ctx, &artifactPB.CreateCatalogRequest{
			NamespaceId: inputStruct.Options.Namespace,
			Name:        inputStruct.Options.CatalogID,
			Description: inputStruct.Options.Description,
			Tags:        inputStruct.Options.Tags,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to create new catalog: %w", err)
		}
	}

	output := UploadFilesOutput{
		File: FileOutput{},
	}
	file := inputStruct.Options.File

	contentType, err := util.GetContentTypeFromBase64(file)
	if err != nil {
		return nil, fmt.Errorf("failed to get content type: %w", err)
	}

	typeString := "FILE_TYPE_" + strings.ToUpper(util.TransformContentTypeToFileExtension(contentType))
	content := util.GetFileBase64Content(file)
	typePB := artifactPB.FileType_value[typeString]
	filePB := &artifactPB.File{
		Name:    inputStruct.Options.FileName,
		Type:    artifactPB.FileType(typePB),
		Content: content,
	}
	uploadRes, err := artifactClient.UploadCatalogFile(ctx, &artifactPB.UploadCatalogFileRequest{
		NamespaceId: inputStruct.Options.Namespace,
		CatalogId:   inputStruct.Options.CatalogID,
		File:        filePB,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	uploadedFilePB := uploadRes.File

	output.File = FileOutput{
		FileUID:    uploadedFilePB.FileUid,
		FileName:   uploadedFilePB.Name,
		FileType:   artifactPB.FileType_name[int32(uploadedFilePB.Type)],
		CreateTime: uploadedFilePB.CreateTime.AsTime().Format(time.RFC3339),
		UpdateTime: uploadedFilePB.UpdateTime.AsTime().Format(time.RFC3339),
		Size:       uploadedFilePB.Size,
		CatalogID:  inputStruct.Options.CatalogID,
	}

	// TODO: chuang, will need to process again in another task.
	_, err = artifactClient.ProcessCatalogFiles(ctx, &artifactPB.ProcessCatalogFilesRequest{
		FileUids: []string{uploadedFilePB.FileUid},
	})

	if err == nil {
		output.Status = true
	}

	return base.ConvertToStructpb(output)
}

func (e *execution) getFilesMetadata(input *structpb.Struct) (*structpb.Struct, error) {
	return nil, nil
}

func (e *execution) getChunksMetadata(input *structpb.Struct) (*structpb.Struct, error) {
	return nil, nil
}

func (e *execution) getFileInMarkdown(input *structpb.Struct) (*structpb.Struct, error) {
	return nil, nil
}

func (e *execution) matchFileStatus(input *structpb.Struct) (*structpb.Struct, error) {
	return nil, nil
}

func (e *execution) searchChunks(input *structpb.Struct) (*structpb.Struct, error) {
	return nil, nil
}

func (e *execution) query(input *structpb.Struct) (*structpb.Struct, error) {
	return nil, nil
}
