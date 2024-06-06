package googlecloudstorage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

type ReadInput struct {
	BucketName               string
	Delimiter                string
	Prefix                   string
	Versions                 bool
	StartOffset              string
	EndOffset                string
	IncludeTrailingDelimiter bool
	MatchGlob                string
	IncludeFoldersAsPrefixes bool
}

type ReadOutput struct {
	TextObjects     []TextObject
	ImageObjects    []ImageObject
	DocumentObjects []DocumentObject
}

type TextObject struct {
	Data       string
	Attributes Attributes
}

type ImageObject struct {
	Data       string
	Attributes Attributes
}

type DocumentObject struct {
	Data       string
	Attributes Attributes
}

type Attributes struct {
	Name               string
	ContentType        string
	ContentLanguage    string
	Owner              string
	Size               int64
	ContentEncoding    string
	ContentDisposition string
	MD5                []byte
	MediaLink          string
	Metadata           map[string]string
	StorageClass       string
}

func readObjects(input ReadInput, client *storage.Client) (ReadOutput, error) {
	bucketName := input.BucketName
	ctx := context.Background()
	query := &storage.Query{
		Delimiter:                input.Delimiter,
		Prefix:                   input.Prefix,
		Versions:                 input.Versions,
		StartOffset:              input.StartOffset,
		EndOffset:                input.EndOffset,
		IncludeTrailingDelimiter: input.IncludeTrailingDelimiter,
		MatchGlob:                input.MatchGlob,
		IncludeFoldersAsPrefixes: input.IncludeFoldersAsPrefixes,
	}

	it := client.Bucket(bucketName).Objects(ctx, query)

	output := ReadOutput{
		TextObjects:     []TextObject{},
		ImageObjects:    []ImageObject{},
		DocumentObjects: []DocumentObject{},
	}
	for {
		attrs, err := it.Next()

		if err == iterator.Done {
			break
		}

		rc, err := client.Bucket(bucketName).Object(attrs.Name).NewReader(ctx)

		if err != nil {
			return output, fmt.Errorf("readObjects: %v", err)
		}
		defer rc.Close()

		b, err := io.ReadAll(rc)
		if err != nil {
			return output, fmt.Errorf("readObjects: %v", err)
		}

		attribute := Attributes{
			Name:               attrs.Name,
			ContentType:        attrs.ContentType,
			ContentLanguage:    attrs.ContentLanguage,
			Owner:              attrs.Owner,
			Size:               attrs.Size,
			ContentEncoding:    attrs.ContentEncoding,
			ContentDisposition: attrs.ContentDisposition,
			MD5:                attrs.MD5,
			MediaLink:          attrs.MediaLink,
			Metadata:           attrs.Metadata,
			StorageClass:       attrs.StorageClass,
		}

		if strings.Contains(attrs.ContentType, "text") {
			textObject := TextObject{
				Data:       string(b),
				Attributes: attribute,
			}
			output.TextObjects = append(output.TextObjects, textObject)
		} else if strings.Contains(attrs.ContentType, "image") {
			imageObject := ImageObject{
				Data:       string(b),
				Attributes: attribute,
			}
			output.ImageObjects = append(output.ImageObjects, imageObject)
		} else { // TODO chuang8511: discuss with reviewer what types should we specify here?
			documentObject := DocumentObject{
				Data:       string(b),
				Attributes: attribute,
			}
			output.DocumentObjects = append(output.DocumentObjects, documentObject)
		}
	}

	return output, nil
}
