package googlecloudstorage

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
)

type CreateBucketInput struct {
	BucketName string `json:"bucket-name"`
	ProjectID  string `json:"project-id"`
	Location   string `json:"location"`
}

type CreateBucketOutput struct {
	Result   string `json:"result"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

func createBucket(input CreateBucketInput, client *storage.Client, ctx context.Context) (CreateBucketOutput, error) {

	output := CreateBucketOutput{}
	bkt := client.Bucket(input.BucketName)

	attr := storage.BucketAttrs{
		Location: input.Location,
	}

	fmt.Println("Creating bucket with name: ", input.BucketName)
	if err := bkt.Create(ctx, input.ProjectID, &attr); err != nil {
		fmt.Println("Error creating bucket: ", err)
		return output, err
	}


	newBktAttrs, err := bkt.Attrs(ctx)
	if err != nil {
		return output, err
	}

	output.Result = "Success"
	output.Name = newBktAttrs.Name
	output.Location = newBktAttrs.Location

	return output, nil
}
