package mongodb

import (
	"context"
	"fmt"

	"github.com/instill-ai/component/base"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/protobuf/types/known/structpb"
)

type InsertInput struct {
	Data map[string]any `json:"data"`
}

type InsertOutput struct {
	Status string `json:"status"`
}

type FindInput struct {
	Filter map[string]any `json:"filter"`
	Limit  int            `json:"limit"`
}

type FindOutput struct {
	Status    string           `json:"status"`
	Documents []map[string]any `json:"documents"`
}

type UpdateInput struct {
	Filter     map[string]any `json:"filter"`
	UpdateData map[string]any `json:"update-data"`
}

type UpdateOutput struct {
	Status string `json:"status"`
}

type DeleteInput struct {
	Filter map[string]any `json:"filter"`
}

type DeleteOutput struct {
	Status string `json:"status"`
}

type DropCollectionInput struct {
	CollectionName string `json:"collection-name"`
}

type DropCollectionOutput struct {
	Status string `json:"status"`
}

type DropDatabaseInput struct {
	DatabaseName string `json:"database-name"`
}

type DropDatabaseOutput struct {
	Status string `json:"status"`
}

func (e *execution) insert(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct InsertInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	data := inputStruct.Data

	_, err = e.client.collectionClient.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}

	outputStruct := InsertOutput{
		Status: "Successfully inserted document",
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// Limit is optional (default is 0)
func (e *execution) find(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct FindInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	filter := inputStruct.Filter
	limit := inputStruct.Limit

	realFilter := bson.M{}
	for key, value := range filter {
		if value != nil {
			realFilter[key] = value
		}
	}

	findOptions := options.Find()

	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}

	var cursor *mongo.Cursor
	if len(filter) != 0 {
		projection := bson.M{}
		projection["_id"] = 0
		for key := range filter {
			projection[key] = 1
		}
		findOptions.SetProjection(projection)
	}
	cursor, err = e.client.collectionClient.Find(ctx, realFilter, findOptions)

	if err != nil {
		return nil, err
	}

	var documents []map[string]any
	for cursor.Next(ctx) {
		var document map[string]any
		err := cursor.Decode(&document)
		if err != nil {
			return nil, err
		}
		documents = append(documents, document)
	}

	outputStruct := FindOutput{
		Status:    "Successfully found documents",
		Documents: documents,
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (e *execution) update(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct UpdateInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	filter := inputStruct.Filter
	updateFields := inputStruct.UpdateData

	var result map[string]interface{}
	err = e.client.collectionClient.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return nil, err
	}

	setFields := bson.M{}
	unsetFields := bson.M{}

	for key := range result {
		if _, found := updateFields[key]; !found && key != "_id" {
			unsetFields[key] = ""
		}
	}

	for key, value := range updateFields {
		setFields[key] = value
	}

	updateDoc := bson.M{}
	if len(setFields) > 0 {
		updateDoc["$set"] = setFields
	}
	if len(unsetFields) > 0 {
		updateDoc["$unset"] = unsetFields
	}

	if len(updateDoc) == 0 {
		return nil, fmt.Errorf("no valid update operations found")
	}

	_, err = e.client.collectionClient.UpdateMany(ctx, filter, updateDoc)
	if err != nil {
		return nil, err
	}

	outputStruct := UpdateOutput{
		Status: "Successfully updated documents",
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (e *execution) delete(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct DeleteInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	filter := inputStruct.Filter

	_, err = e.client.collectionClient.DeleteMany(ctx, filter)
	if err != nil {
		return nil, err
	}

	outputStruct := DeleteOutput{
		Status: "Successfully deleted documents",
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (e *execution) dropCollection(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {

	err := e.client.collectionClient.Drop(ctx)
	if err != nil {
		return nil, err
	}

	outputStruct := DropCollectionOutput{
		Status: "Successfully dropped collection",
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (e *execution) dropDatabase(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {

	err := e.client.databaseClient.Drop(ctx)
	if err != nil {
		return nil, err
	}

	outputStruct := DropDatabaseOutput{
		Status: "Successfully dropped database",
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil
}
