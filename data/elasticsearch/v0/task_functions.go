package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

type IndexInput struct {
	Data      map[string]any `json:"data"`
	IndexName string         `json:"index-name"`
}

type IndexOutput struct {
	Status string `json:"status"`
}

type UpdateInput struct {
	Update    map[string]any `json:"update"`
	Filter    map[string]any `json:"filter"`
	FilterSQL string         `json:"filter-sql"`
	Query     string         `json:"query"`
	IndexName string         `json:"index-name"`
}

type UpdateOutput struct {
	Status string `json:"status"`
}

type SearchInput struct {
	Fields    []string       `json:"fields"`
	Mode      string         `json:"mode"`
	MinScore  float64        `json:"min-score"`
	Filter    map[string]any `json:"filter"`
	FilterSQL string         `json:"filter-sql"`
	Query     string         `json:"query"`
	IndexName string         `json:"index-name"`
	Size      int            `json:"size"`
}

type SearchOutput struct {
	Documents []map[string]any `json:"documents"`
	Status    string           `json:"status"`
}

type VectorSearchInput struct {
	Mode          string         `json:"mode"`
	Filter        map[string]any `json:"filter"`
	FilterSQL     string         `json:"filter-sql"`
	IndexName     string         `json:"index-name"`
	Field         string         `json:"field"`
	Fields        []string       `json:"fields"`
	QueryVector   []float64      `json:"query-vector"`
	K             int            `json:"k"`
	NumCandidates int            `json:"num-candidates"`
	MinScore      float64        `json:"min-score"`
}

type SearchResponse struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		MaxScore float64 `json:"max_score"`
		Hits     []Hit   `json:"hits"`
	} `json:"hits"`
}

type Hit struct {
	Index  string         `json:"_index"`
	ID     string         `json:"_id"`
	Score  float64        `json:"_score"`
	Source map[string]any `json:"_source"`
}

type DeleteInput struct {
	Filter    map[string]any `json:"filter"`
	FilterSQL string         `json:"filter-sql"`
	Query     string         `json:"query"`
	IndexName string         `json:"index-name"`
}

type DeleteOutput struct {
	Status string `json:"status"`
}

type CreateIndexInput struct {
	IndexName string         `json:"index-name"`
	Mappings  map[string]any `json:"mappings"`
}

type CreateIndexOutput struct {
	Status string `json:"status"`
}

type DeleteIndexInput struct {
	IndexName string `json:"index-name"`
}

type DeleteIndexOutput struct {
	Status string `json:"status"`
}

type CountResponse struct {
	Count int `json:"count"`
}

func translateSQLQuery(es *esapi.SQLTranslate, query string, indexName string) (map[string]any, error) {
	sqlTranslateClient := ESSQLTranslate(*es)
	sqlQuery := fmt.Sprintf("SELECT * FROM %s WHERE %s", indexName, query)
	queryJSON := map[string]any{"query": sqlQuery}
	translateJSON, err := json.Marshal(queryJSON)
	if err != nil {
		return nil, err
	}

	res, err := sqlTranslateClient(bytes.NewReader(translateJSON))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error translating SQL: %s", res.Status())
	}

	var response map[string]any
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	translatedQuery, exists := response["query"]
	if !exists {
		return nil, fmt.Errorf("query not found in response")
	}

	return translatedQuery.(map[string]any), nil
}

func IndexDocument(es *esapi.Index, indexName string, data map[string]any) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	esClient := ESIndex(*es)

	res, err := esClient(indexName, bytes.NewReader(dataJSON), func(r *esapi.IndexRequest) {
		r.Refresh = "true"
	})

	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing document: %s", res.Status())
	}

	return nil
}

// size is optional, empty means all documents
func SearchDocument(es *esapi.Search, esSQLTranslate *esapi.SQLTranslate, inputStruct SearchInput) ([]Hit, error) {
	indexName := inputStruct.IndexName
	query := inputStruct.Query
	minScore := inputStruct.MinScore
	filter := inputStruct.Filter
	filterSQL := inputStruct.FilterSQL
	size := inputStruct.Size
	fields := inputStruct.Fields

	queryJSON := map[string]any{}

	if minScore > 0 {
		queryJSON["min_score"] = minScore
	}
	if filterSQL != "" && filter == nil {
		translatedQuery, err := translateSQLQuery(esSQLTranslate, filterSQL, indexName)
		if err != nil {
			return nil, err
		}
		queryJSON["query"] = translatedQuery
	} else if filter != nil {
		queryJSON["query"] = filter
	}
	if len(fields) > 0 {
		queryJSON["_source"] = fields
	}

	filterJSON, err := json.Marshal(queryJSON)
	if err != nil {
		return nil, err
	}

	body := strings.NewReader(string(filterJSON))

	esClient := ESSearch(*es)
	res, err := esClient(func(r *esapi.SearchRequest) {
		r.Index = []string{indexName}
		r.Body = body
		r.Query = query
		r.TrackTotalHits = true
		if size > 0 {
			r.Size = esapi.IntPtr(size)
		}
	})

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching document: %s", res.Status())
	}

	var response SearchResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Hits.Hits, nil
}

// Only support vector search for now, for semantic search, we can use external model on other component combined with vector search
// Can support up to more than 1 vector similarity fields, overall result will then be combined
func VectorSearchDocument(es *esapi.Search, esSQLTranslate *esapi.SQLTranslate, inputStruct VectorSearchInput) ([]Hit, error) {
	indexName := inputStruct.IndexName
	field := inputStruct.Field
	queryVector := inputStruct.QueryVector
	k := inputStruct.K
	numCandidates := inputStruct.NumCandidates
	minScore := inputStruct.MinScore
	filter := inputStruct.Filter
	fields := inputStruct.Fields
	filterSQL := inputStruct.FilterSQL

	var body io.Reader = nil
	query := make(map[string]any)

	knnQuery := map[string]any{
		"field":          field,
		"query_vector":   queryVector,
		"k":              k,
		"num_candidates": 2 * k,
	}

	if filterSQL != "" && filter == nil {
		translatedQuery, err := translateSQLQuery(esSQLTranslate, filterSQL, indexName)
		if err != nil {
			return nil, err
		}
		knnQuery["filter"] = translatedQuery
	} else if filter != nil {
		knnQuery["filter"] = filter
	}
	if numCandidates > 0 {
		knnQuery["num_candidates"] = numCandidates
	}

	query["knn"] = knnQuery
	query["size"] = k
	if minScore > 0 {
		query["min_score"] = minScore
	}
	if len(fields) > 0 {
		query["_source"] = fields
	}

	filterJSON, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	body = strings.NewReader(string(filterJSON))

	esClient := ESSearch(*es)
	res, err := esClient(func(r *esapi.SearchRequest) {
		r.Index = []string{indexName}
		r.Body = body
		r.TrackTotalHits = true
	})

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error vector searching document: %s", res.Status())
	}

	var response SearchResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Hits.Hits, nil
}

func UpdateDocument(es *esapi.UpdateByQuery, esSQLTranslate *esapi.SQLTranslate, inputStruct UpdateInput) error {
	indexName := inputStruct.IndexName
	query := inputStruct.Query
	filter := inputStruct.Filter
	filterSQL := inputStruct.FilterSQL
	update := inputStruct.Update

	if filterSQL != "" && filter == nil {
		translatedQuery, err := translateSQLQuery(esSQLTranslate, filterSQL, indexName)
		if err != nil {
			return err
		}
		filter = translatedQuery
	}

	updateByQueryReq := map[string]any{
		"query": filter,
		"script": map[string]any{
			"source": "for (entry in params.entry.entrySet()) { ctx._source[entry.getKey()] = entry.getValue() }",
			"lang":   "painless",
			"params": map[string]any{
				"entry": update,
			},
		},
	}

	var body io.Reader = nil
	if filter != nil {
		updateJSON, err := json.Marshal(updateByQueryReq)
		if err != nil {
			return err
		}

		body = strings.NewReader(string(updateJSON))
	}

	esClient := ESUpdate(*es)

	res, err := esClient([]string{indexName}, func(r *esapi.UpdateByQueryRequest) {
		r.Body = body
		r.Query = query
		r.Refresh = esapi.BoolPtr(true)
	})

	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error updating documents: %s", res.Status())
	}

	return nil
}

func DeleteDocument(es *esapi.DeleteByQuery, esSQLTranslate *esapi.SQLTranslate, inputStruct DeleteInput) error {
	indexName := inputStruct.IndexName
	query := inputStruct.Query
	filter := inputStruct.Filter
	filterSQL := inputStruct.FilterSQL

	if filterSQL != "" && filter == nil {
		translatedQuery, err := translateSQLQuery(esSQLTranslate, filterSQL, indexName)
		if err != nil {
			return err
		}
		filter = translatedQuery
	}
	deleteByQueryReq := map[string]any{
		"query": filter,
	}

	var body io.Reader = nil
	if filter != nil {
		filterJSON, err := json.Marshal(deleteByQueryReq)
		if err != nil {
			return err
		}

		body = bytes.NewReader(filterJSON)
	}

	esClient := ESDelete(*es)

	res, err := esClient([]string{indexName}, body, func(r *esapi.DeleteByQueryRequest) {
		r.Query = query
		r.Refresh = esapi.BoolPtr(true)
	})
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error deleting documents: %s", res.Status())
	}

	return nil
}

func DeleteIndex(es *esapi.IndicesDelete, indexName string) error {
	esClient := ESDeleteIndex(*es)

	res, err := esClient([]string{indexName})

	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error deleting index: %s", res.Status())
	}

	return nil
}

// mappings refer to elasticsearch documentation for more information, use dense_vector type with similarity and dims fields
// pre-defined mappings is mandatory for vector search, if index isnt created with mappings, vector search will not work as dense_vector type doesnt explicitly defined
func CreateIndex(es *esapi.IndicesCreate, indexName string, mappings map[string]any) error {
	createIndexReq := map[string]map[string]any{
		"mappings": {
			"properties": mappings,
		},
	}

	createIndexJSON, err := json.Marshal(createIndexReq)
	if err != nil {
		return err
	}

	esClient := ESCreateIndex(*es)

	res, err := esClient(indexName, func(r *esapi.IndicesCreateRequest) {
		r.Body = strings.NewReader(string(createIndexJSON))
	})

	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error creating index: %s", res.Status())
	}

	return nil
}

func (e *execution) index(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct IndexInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	err = IndexDocument(&e.client.indexClient, inputStruct.IndexName, inputStruct.Data)
	if err != nil {
		return nil, err
	}

	outputStruct := IndexOutput{
		Status: "Successfully indexed document",
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

	err = UpdateDocument(&e.client.updateClient, &e.client.sqlTranslateClient, inputStruct)
	if err != nil {
		return nil, err
	}

	outputStruct := UpdateOutput{
		Status: "Successfully updated document",
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (e *execution) search(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct SearchInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	resultTemp, err := SearchDocument(&e.client.searchClient, &e.client.sqlTranslateClient, inputStruct)

	if err != nil {
		return nil, err
	}

	var result []map[string]any
	if inputStruct.Mode == "Source Only" {
		for _, hit := range resultTemp {
			result = append(result, hit.Source)
		}
	} else {
		for _, hit := range resultTemp {
			hitMap := make(map[string]any)
			hitMap["_index"] = hit.Index
			hitMap["_id"] = hit.ID
			hitMap["_score"] = hit.Score
			hitMap["_source"] = hit.Source
			result = append(result, hitMap)
		}
	}

	outputStruct := SearchOutput{
		Documents: result,
		Status:    "Successfully searched document",
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (e *execution) vectorSearch(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct VectorSearchInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	resultTemp, err := VectorSearchDocument(&e.client.searchClient, &e.client.sqlTranslateClient, inputStruct)

	if err != nil {
		return nil, err
	}

	var result []map[string]any
	if inputStruct.Mode == "Source Only" {
		for _, hit := range resultTemp {
			result = append(result, hit.Source)
		}
	} else {
		for _, hit := range resultTemp {
			hitMap := make(map[string]any)
			hitMap["_index"] = hit.Index
			hitMap["_id"] = hit.ID
			hitMap["_score"] = hit.Score
			hitMap["_source"] = hit.Source
			result = append(result, hitMap)
		}
	}

	outputStruct := SearchOutput{
		Documents: result,
		Status:    "Successfully vector searched document",
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

	err = DeleteDocument(&e.client.deleteClient, &e.client.sqlTranslateClient, inputStruct)
	if err != nil {
		return nil, err
	}

	outputStruct := DeleteOutput{
		Status: "Successfully deleted document",
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (e *execution) createIndex(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct CreateIndexInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	err = CreateIndex(&e.client.createIndexClient, inputStruct.IndexName, inputStruct.Mappings)
	if err != nil {
		return nil, err
	}

	outputStruct := CreateIndexOutput{
		Status: "Successfully created index",
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (e *execution) deleteIndex(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct DeleteIndexInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	err = DeleteIndex(&e.client.deleteIndexClient, inputStruct.IndexName)
	if err != nil {
		return nil, err
	}

	outputStruct := DeleteIndexOutput{
		Status: "Successfully deleted index",
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}
	return output, nil
}
