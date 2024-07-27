package weaviate

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/instill-ai/component/base"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/filters"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"google.golang.org/protobuf/types/known/structpb"
)

type UpsertInput struct {
	CollectionName string         `json:"collection-name"`
	Vector         []float32      `json:"vector"`
	Properties     map[string]any `json:"properties"`
}

type UpsertOutput struct {
	Status string `json:"status"`
}

type VectorSearchInput struct {
	CollectionName string         `json:"collection-name"`
	Vector         []float32      `json:"vector"`
	Filter         map[string]any `json:"filter"`
	Limit          int            `json:"limit"`
	Fields         []string       `json:"fields"`
}

type VectorSearchOutput struct {
	Status  string           `json:"status"`
	Objects []map[string]any `json:"objects"`
}

type UpdateInput struct {
	Filter     map[string]any `json:"filter"`
	UpdateData map[string]any `json:"update-data"`
}

type UpdateOutput struct {
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
				path := value.([]string)
				where.WithPath(path)
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
				val := value.(int64)
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

func VectorSearch(ctx context.Context, client *weaviate.Client, inputStruct VectorSearchInput) (map[string]any, error) {
	collectionName := inputStruct.CollectionName
	filter := inputStruct.Filter
	limit := inputStruct.Limit
	rawFields := inputStruct.Fields
	vector := inputStruct.Vector

	nearVector := client.GraphQL().
		NearVectorArgBuilder().
		WithVector(vector)

	withBuilder := client.GraphQL().Get().
		WithClassName(collectionName).
		WithNearVector(nearVector)

	if filter != nil {
		where, err := jsonToWhereBuilder(&filter)
		if err != nil {
			return nil, err
		}
		withBuilder.WithWhere(where)
	}

	withBuilder.WithLimit(limit)

	fields := []graphql.Field{
		{Name: "_additional", Fields: []graphql.Field{
			{Name: "distance"},
		}},
	}

	if len(rawFields) > 0 {
		for _, field := range rawFields {
			fields = append(fields, graphql.Field{Name: field})
		}

		withBuilder.WithFields(fields...)
	}

	res, err := withBuilder.Do(ctx)

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

	return mapRes, nil
}

func (e *execution) upsert(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct UpsertInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	_, err = e.client.Data().Creator().
		WithClassName(inputStruct.CollectionName).
		WithVector(inputStruct.Vector).
		WithProperties(inputStruct.Properties).Do(ctx)
	if err != nil {
		return nil, err
	}

	outputStruct := UpsertOutput{
		Status: "Successfully upserted document",
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

	res, err := VectorSearch(ctx, e.client, inputStruct)
	if err != nil {
		return nil, err
	}

	outputStruct := VectorSearchOutput{
		Status:  "Successfully upserted document",
		Objects: []map[string]any{res},
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

// 	var documents []map[string]any
// 	for cursor.Next(ctx) {
// 		var document map[string]any
// 		err := cursor.Decode(&document)
// 		if err != nil {
// 			return nil, err
// 		}
// 		documents = append(documents, document)
// 	}

// 	outputStruct := FindOutput{
// 		Status:    "Successfully found documents",
// 		Documents: documents,
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
// 		Status: "Successfully updated documents",
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
// 		Status: "Successfully deleted documents",
// 	}

// 	output, err := base.ConvertToStructpb(outputStruct)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return output, nil
// }
