package mongodb

import (
	"context"

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
	Fields []string       `json:"fields"`
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

type CreateSearchIndexInput struct {
	IndexName string         `json:"index-name"`
	IndexType string         `json:"index-type"`
	Syntax    map[string]any `json:"syntax"`
}

type CreateSearchIndexOutput struct {
	Status string `json:"status"`
}

type DropSearchIndexInput struct {
	IndexName string `json:"index-name"`
}

type DropSearchIndexOutput struct {
	Status string `json:"status"`
}

type VectorSearchInput struct {
	Exact         bool           `json:"exact"`
	Filter        map[string]any `json:"filter"`
	IndexName     string         `json:"index-name"`
	Limit         int            `json:"limit"`
	NumCandidates int            `json:"num-candidates"`
	Path          string         `json:"path"`
	QueryVector   []float64      `json:"query-vector"`
	Fields        []string       `json:"fields"`
}

type VectorSearchOutput struct {
	Status    string           `json:"status"`
	Documents []map[string]any `json:"documents"`
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
	fields := inputStruct.Fields

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
	if len(fields) > 0 {
		projection := bson.M{}
		projection["_id"] = 0
		for _, field := range fields {
			projection[field] = 1
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

	_, err = e.client.collectionClient.UpdateMany(ctx, filter, updateFields)
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

func (e *execution) createSearchIndex(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct CreateSearchIndexInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	syntax := inputStruct.Syntax

	searchIndexModel := mongo.SearchIndexModel{
		Definition: syntax,
		Options: &options.SearchIndexesOptions{
			Name: &inputStruct.IndexName,
			Type: &inputStruct.IndexType,
		},
	}

	_, err = e.client.searchIndexClient.CreateOne(ctx, searchIndexModel)
	if err != nil {
		return nil, err
	}

	outputStruct := CreateSearchIndexOutput{
		Status: "Successfully created search index",
	}

	// Convert the output structure to Structpb
	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (e *execution) dropSearchIndex(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct DropSearchIndexInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	indexName := inputStruct.IndexName

	err = e.client.searchIndexClient.DropOne(ctx, indexName)
	if err != nil {
		return nil, err
	}

	outputStruct := DropSearchIndexOutput{
		Status: "Successfully dropped search index",
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (e *execution) vectorSearch(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct VectorSearchInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	exact := inputStruct.Exact
	filter := inputStruct.Filter
	indexName := inputStruct.IndexName
	limit := inputStruct.Limit
	numCandidates := inputStruct.NumCandidates
	path := inputStruct.Path
	queryVector := inputStruct.QueryVector
	fields := inputStruct.Fields

	vectorSearch := bson.M{
		"exact":       exact,
		"index":       indexName,
		"path":        path,
		"queryVector": queryVector,
		"limit":       limit,
	}
	if filter != nil {
		vectorSearch["filter"] = filter
	}

	if !exact {
		if numCandidates > 0 {
			vectorSearch["numCandidates"] = numCandidates
		} else {
			vectorSearch["numCandidates"] = 3 * limit
		}
	}

	project := bson.M{"_id": 0}
	for _, field := range fields {
		project[field] = 1
	}

	query := bson.A{
		bson.M{
			"$vectorSearch": vectorSearch,
		},
		bson.M{
			"$addFields": bson.M{
				"score": bson.M{
					"$meta": "vectorSearchScore",
				},
			},
		},
	}

	if len(fields) > 0 {
		query = append(query, bson.M{
			"$project": project,
		})
	}

	cursor, err := e.client.collectionClient.Aggregate(ctx, query)
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

	outputStruct := VectorSearchOutput{
		Status:    "Successfully found documents",
		Documents: documents,
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}

	return output, nil
}
