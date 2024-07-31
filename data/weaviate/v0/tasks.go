package weaviate

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/instill-ai/component/base"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/fault"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/filters"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
	"google.golang.org/protobuf/types/known/structpb"
)

type InsertInput struct {
	CollectionName string         `json:"collection-name"`
	Vector         []float32      `json:"vector"`
	Metadata       map[string]any `json:"metadata"`
}

type InsertOutput struct {
	Status string `json:"status"`
}

type VectorSearchInput struct {
	CollectionName string         `json:"collection-name"`
	Vector         []float32      `json:"vector"`
	Filter         map[string]any `json:"filter"`
	Limit          int            `json:"limit"`
	Fields         []string       `json:"fields"`
	WithAdditional bool           `json:"with-additional"`
	Tenant         string         `json:"tenant"`
}

type Result struct {
	Objects  []map[string]any `json:"objects"`
	Vectors  [][]float32      `json:"vectors"`
	Metadata []map[string]any `json:"metadata"`
}

type VectorSearchOutput struct {
	Status string `json:"status"`
	Result Result `json:"result"`
}

type DeleteInput struct {
	CollectionName string         `json:"collection-name"`
	Filter         map[string]any `json:"filter"`
}

type DeleteOutput struct {
	Status string `json:"status"`
}

type BatchInsertInput struct {
	CollectionName string           `json:"collection-name"`
	ArrayMetadata  []map[string]any `json:"array-metadata"`
	ArrayVector    [][]float32      `json:"array-vector"`
}

type BatchInsertOutput struct {
	Status string `json:"status"`
}

type DeleteCollectionInput struct {
	CollectionName string `json:"collection-name"`
}

type DeleteCollectionOutput struct {
	Status string `json:"status"`
}

func jsonToWhereBuilder(jsonWhere *map[string]any) (*filters.WhereBuilder, error) {
	where := filters.Where()

	for key, value := range *jsonWhere {
		if key == "operands" {
			values := value.([]map[string]any)
			var operands []*filters.WhereBuilder
			for _, nestedJSONWhere := range values {
				operand, err := jsonToWhereBuilder(&nestedJSONWhere)
				operands = append(operands, operand)

				if err != nil {
					return nil, err
				}
			}
			where.WithOperands(operands)
		} else {
			switch key {
			case "path":
				path := value.(string)
				where.WithPath([]string{path})
			case "operator":
				operator := value.(string)
				switch operator {
				case "And":
					where.WithOperator(filters.And)
				case "Or":
					where.WithOperator(filters.Or)
				case "Equal":
					where.WithOperator(filters.Equal)
				case "NotEqual":
					where.WithOperator(filters.NotEqual)
				case "GreaterThan":
					where.WithOperator(filters.GreaterThan)
				case "GreaterThanEqual":
					where.WithOperator(filters.GreaterThanEqual)
				case "LessThan":
					where.WithOperator(filters.LessThan)
				case "LessThanEqual":
					where.WithOperator(filters.LessThanEqual)
				case "Like":
					where.WithOperator(filters.Like)
				case "WithinGeoRange":
					where.WithOperator(filters.WithinGeoRange)
				case "IsNull":
					where.WithOperator(filters.IsNull)
				case "ContainsAny":
					where.WithOperator(filters.ContainsAny)
				case "ContainsAll":
					where.WithOperator(filters.ContainsAll)
				default:
					return nil, fmt.Errorf("unknown operator: %s", operator)
				}
			case "valueInt":
				val := int64(value.(float64))
				where.WithValueInt(val)
			case "valueBoolean":
				val := value.(bool)
				where.WithValueBoolean(val)
			case "valueString":
				val := value.(string)
				where.WithValueString(val)
			case "valueText":
				val := value.(string)
				where.WithValueText(val)
			case "valueNumber":
				val := value.(float64)
				where.WithValueNumber(val)
			case "valueDate":
				val := value.(time.Time)
				where.WithValueDate(val)
			default:
				return nil, fmt.Errorf("unknown key: %s", key)
			}
		}
	}

	return where, nil
}

func getAllFields(ctx context.Context, client WeaviateSchemaAPIClassGetterClient, collectionName string) ([]string, error) {
	res, err := client.WithClassName(collectionName).Do(ctx)
	if err != nil {
		return nil, err
	}

	mapRes := make(map[string]any)
	byteRes, err := res.MarshalBinary()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(byteRes, &mapRes)
	if err != nil {
		return nil, err
	}

	properties, ok := mapRes["properties"].([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected type for properties")
	}

	var fields []string
	for _, property := range properties {
		if propertyMap, ok := property.(map[string]any); ok {
			fieldName, ok := propertyMap["name"].(string)
			if !ok {
				return nil, fmt.Errorf("unexpected type for field name")
			}
			fields = append(fields, fieldName)
		} else {
			return nil, fmt.Errorf("unexpected property format")
		}
	}

	return fields, nil
}

func VectorSearch(ctx context.Context, client WeaviateClient, inputStruct VectorSearchInput) ([]map[string]any, error) {
	collectionName := inputStruct.CollectionName
	filter := inputStruct.Filter
	limit := inputStruct.Limit
	rawFields := inputStruct.Fields
	vector := inputStruct.Vector
	tenant := inputStruct.Tenant

	nearVector := client.graphQLNearVectorArgumentBuilder.
		WithVector(vector)

	withBuilder := client.graphQLAPIGetClient.
		WithClassName(collectionName).
		WithNearVector(nearVector)

	if filter != nil {
		where, err := jsonToWhereBuilder(&filter)
		if err != nil {
			return nil, err
		}
		withBuilder.WithWhere(where)
	}
	if limit > 0 {
		withBuilder.WithLimit(limit)
	}
	fields := []graphql.Field{{Name: "_additional", Fields: []graphql.Field{
		{Name: "id"},
		{Name: "distance"},
		{Name: "vector"},
	}}}
	if len(rawFields) == 0 || rawFields == nil {
		allFields, err := getAllFields(ctx, client.schemaAPIClassGetterClient, collectionName)
		if err != nil {
			return nil, err
		}
		rawFields = allFields
	}
	for _, field := range rawFields {
		fields = append(fields, graphql.Field{Name: field})
	}
	withBuilder.WithFields(fields...)
	if tenant != "" {
		withBuilder.WithTenant(tenant)
	}

	res, err := withBuilder.Do(ctx)
	if err != nil || res.Errors != nil {
		return nil, err
	}

	mapRes := make(map[string]any)
	byteRes, err := res.MarshalBinary()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(byteRes, &mapRes)

	if err != nil {
		return nil, err
	}

	data, ok := mapRes["data"].(map[string]any)["Get"].(map[string]any)[collectionName].([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected type for data")
	}

	var results []map[string]any
	for _, item := range data {
		if itemMap, ok := item.(map[string]any); ok {
			results = append(results, itemMap)
		} else {
			return nil, fmt.Errorf("unexpected item format")
		}
	}

	return results, nil
}

func (e *execution) insert(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct InsertInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	_, err = e.client.dataAPICreatorClient.
		WithClassName(inputStruct.CollectionName).
		WithVector(inputStruct.Vector).
		WithProperties(inputStruct.Metadata).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	outputStruct := InsertOutput{
		Status: "Successfully inserted 1 object",
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

	res, err := VectorSearch(ctx, *e.client, inputStruct)
	if err != nil {
		return nil, err
	}

	var objects []map[string]any
	var vectors [][]float32
	var metadata []map[string]any

	for _, item := range res {
		vector, ok := item["_additional"].(map[string]any)["vector"].([]any)
		if !ok {
			return nil, fmt.Errorf("unexpected type for vector")
		}
		vectorFloat := make([]float32, len(vector))
		for i, v := range vector {
			vectorFloat[i] = float32(v.(float64))
		}

		vectors = append(vectors, vectorFloat)
		tempProperties := make(map[string]any)
		if !inputStruct.WithAdditional {
			delete(item, "_additional")
			metadata = append(metadata, item)
		} else {
			for key, value := range item {
				if key != "_additional" {
					tempProperties[key] = value
				}
			}
			metadata = append(metadata, tempProperties)
		}
		objects = append(objects, item)
	}

	outputStruct := VectorSearchOutput{
		Status: fmt.Sprintf("Successfully found %d objects", len(objects)),
		Result: Result{
			Objects:  objects,
			Vectors:  vectors,
			Metadata: metadata,
		},
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
	collectionName := inputStruct.CollectionName

	where, err := jsonToWhereBuilder(&filter)
	if err != nil {
		return nil, err
	}

	res, err := e.client.batchAPIDeleterClient.
		WithClassName(collectionName).
		WithWhere(where).
		Do(ctx)

	if err != nil {
		return nil, err
	}

	outputStruct := DeleteOutput{
		Status: fmt.Sprintf("Successfully deleted %d objects", res.Results.Successful),
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (e *execution) batchInsert(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct BatchInsertInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	collectionName := inputStruct.CollectionName
	arrayProperties := inputStruct.ArrayMetadata
	arrayVector := inputStruct.ArrayVector

	batcher := e.client.batchAPIBatcherClient
	for i, properties := range arrayProperties {
		batcher.WithObjects(&models.Object{
			Class:      collectionName,
			Properties: properties,
			Vector:     arrayVector[i],
		})
	}
	_, err = batcher.Do(ctx)

	if err != nil {
		return nil, err
	}

	outputStruct := BatchInsertOutput{
		Status: fmt.Sprintf("Successfully batch inserted %d objects", len(arrayProperties)),
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (e *execution) deleteCollection(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct DeleteCollectionInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	collectionName := inputStruct.CollectionName

	if err := e.client.schemaAPIDeleterClient.
		WithClassName(collectionName).
		Do(context.Background()); err != nil {
		if status, ok := err.(*fault.WeaviateClientError); ok && status.StatusCode != http.StatusBadRequest {
			panic(err)
		}
	}

	outputStruct := DeleteCollectionOutput{
		Status: "Successfully deleted 1 collection",
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// // Limit is optional (default is 0)
// func (e *execution) find(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
// 	var inputStruct FindInput
// 	err := base.ConvertFromStructpb(in, &inputStruct)
// 	if err != nil {
// 		return nil, err
// 	}

// 	filter := inputStruct.Filter
// 	limit := inputStruct.Limit
// 	fields := inputStruct.Fields

// 	realFilter := bson.M{}
// 	for key, value := range filter {
// 		if value != nil {
// 			realFilter[key] = value
// 		}
// 	}

// 	findOptions := options.Find()

// 	if limit > 0 {
// 		findOptions.SetLimit(int64(limit))
// 	}

// 	var cursor *mongo.Cursor
// 	if len(fields) > 0 {
// 		projection := bson.M{}
// 		projection["_id"] = 0
// 		for _, field := range fields {
// 			projection[field] = 1
// 		}
// 		findOptions.SetProjection(projection)
// 	}
// 	cursor, err = e.client.collectionClient.Find(ctx, realFilter, findOptions)

// 	if err != nil {
// 		return nil, err
// 	}

// 	var objects []map[string]any
// 	for cursor.Next(ctx) {
// 		var object map[string]any
// 		err := cursor.Decode(&object)
// 		if err != nil {
// 			return nil, err
// 		}
// 		objects = append(objects, object)
// 	}

// 	outputStruct := FindOutput{
// 		Status:    "Successfully found objects",
// 		objects: objects,
// 	}

// 	output, err := base.ConvertToStructpb(outputStruct)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return output, nil
// }

// func (e *execution) update(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
// 	var inputStruct UpdateInput
// 	err := base.ConvertFromStructpb(in, &inputStruct)
// 	if err != nil {
// 		return nil, err
// 	}

// 	filter := inputStruct.Filter
// 	updateFields := inputStruct.UpdateData

// 	if err != nil {
// 		return nil, err
// 	}

// 	setFields := bson.M{}

// 	for key, value := range updateFields {
// 		setFields[key] = value
// 	}

// 	updateDoc := bson.M{}
// 	if len(setFields) > 0 {
// 		updateDoc["$set"] = setFields
// 	}

// 	if len(updateDoc) == 0 {
// 		return nil, fmt.Errorf("no valid update operations found")
// 	}

// 	_, err = e.client.collectionClient.UpdateMany(ctx, filter, updateDoc)
// 	if err != nil {
// 		return nil, err
// 	}

// 	outputStruct := UpdateOutput{
// 		Status: "Successfully updated objects",
// 	}

// 	output, err := base.ConvertToStructpb(outputStruct)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return output, nil
// }

// func (e *execution) delete(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
// 	var inputStruct DeleteInput
// 	err := base.ConvertFromStructpb(in, &inputStruct)
// 	if err != nil {
// 		return nil, err
// 	}

// 	filter := inputStruct.Filter

// 	_, err = e.client.collectionClient.DeleteMany(ctx, filter)
// 	if err != nil {
// 		return nil, err
// 	}

// 	outputStruct := DeleteOutput{
// 		Status: "Successfully deleted objects",
// 	}

// 	output, err := base.ConvertToStructpb(outputStruct)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return output, nil
// }
