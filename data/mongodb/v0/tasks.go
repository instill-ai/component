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
	Criteria map[string]any `json:"criteria"`
}

type FindOutput struct {
	Status    string           `json:"status"`
	Documents []map[string]any `json:"documents"`
}

type UpdateInput struct {
	Criteria map[string]any `json:"criteria"`
	Update   map[string]any `json:"update"`
}

type UpdateOutput struct {
	Status string `json:"status"`
}

type DeleteInput struct {
	Criteria map[string]any `json:"criteria"`
}

type DeleteOutput struct {
	Status string `json:"status"`
}

func (e *execution) insert(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct InsertInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	data := inputStruct.Data
	ctx := context.Background()

	_, err = e.client.InsertOne(ctx, data)
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

func (e *execution) find(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct FindInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	criteria := inputStruct.Criteria
	ctx := context.Background()

	realCriteria := bson.M{}
	for key, value := range criteria {
		if value != nil {
			realCriteria[key] = value
		}
	}

	var cursor *mongo.Cursor
	if len(criteria) == 0 {
		// Find all documents with the specified projection
		cursor, err = e.client.Find(ctx, realCriteria)
	} else {
		// Define the projection to include only the email field
		projection := bson.M{}

		projection["_id"] = 0
		for key := range criteria {
			projection[key] = 1
		}

		// Find all documents with the specified projection
		cursor, err = e.client.Find(ctx, realCriteria, options.Find().SetProjection(projection))
	}

	if err != nil {
		return nil, err
	}

	var documents []map[string]any
	for cursor.Next(ctx) {
		var row map[string]any
		err := cursor.Decode(&row)
		if err != nil {
			return nil, err
		}
		documents = append(documents, row)
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

func (e *execution) update(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct UpdateInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	criteria := inputStruct.Criteria
	updateFields := inputStruct.Update
	ctx := context.Background()

	// Get the document to identify which fields need to be removed
	var result map[string]interface{}
	err = e.client.FindOne(ctx, criteria).Decode(&result)
	if err != nil {
		return nil, err
	}

	setFields := bson.M{}
	unsetFields := bson.M{}

	for key := range result {
		// If the field in the existing document is not in the updateFields, it should be removed
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

	_, err = e.client.UpdateMany(ctx, criteria, updateDoc)
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

func (e *execution) delete(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct DeleteInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	criteria := inputStruct.Criteria
	ctx := context.Background()

	_, err = e.client.DeleteMany(ctx, criteria)
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
