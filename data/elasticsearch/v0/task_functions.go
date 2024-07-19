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
	Query     string         `json:"query"`
	IndexName string         `json:"index-name"`
}

type UpdateOutput struct {
	Status string `json:"status"`
}

type SearchInput struct {
	Mode      string         `json:"mode"`
	Filter    map[string]any `json:"filter"`
	Query     string         `json:"query"`
	IndexName string         `json:"index-name"`
	Size      int            `json:"size"`
}

type SearchOutput struct {
	Documents []map[string]any `json:"documents"`
	Status    string           `json:"status"`
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
	Query     string         `json:"query"`
	IndexName string         `json:"index-name"`
}

type DeleteOutput struct {
	Status string `json:"status"`
}

type DeleteIndexInput struct {
	IndexName string `json:"index-name"`
}

type DeleteIndexOutput struct {
	Status string `json:"status"`
}

func indexDocument(es *esapi.Index, indexName string, data map[string]any) error {
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

func searchDocument(es *esapi.Search, indexName string, query string, rawFilter map[string]any, size int) ([]Hit, error) {
	var body io.Reader = nil
	if rawFilter != nil {
		filter := make(map[string]any)
		var source []string
		for key, value := range rawFilter {
			if value != nil {
				filter[key] = value
			}
			source = append(source, key)
		}

		filterQuery := make(map[string]any)
		if len(source) > 0 {
			filterQuery["_source"] = source
		}
		if len(filter) > 0 {
			filterQuery["query"] = map[string]any{
				"match": filter,
			}
		}

		filterJSON, err := json.Marshal(filterQuery)
		if err != nil {
			return nil, err
		}

		body = strings.NewReader(string(filterJSON))
	}

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

func updateDocument(es *esapi.UpdateByQuery, indexName string, query string, filter map[string]any, update map[string]any) error {
	updateByQueryReq := map[string]any{
		"query": map[string]any{
			"match": filter,
		},
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

func deleteDocument(es *esapi.DeleteByQuery, indexName string, query string, filter map[string]any) error {
	deleteByQueryReq := map[string]any{
		"query": map[string]any{
			"match": filter,
		},
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

func deleteIndex(es *esapi.IndicesDelete, indexName string) error {
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

func (e *execution) index(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct IndexInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	err = indexDocument(&e.indexClient, inputStruct.IndexName, inputStruct.Data)
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

	err = updateDocument(&e.updateClient, inputStruct.IndexName, inputStruct.Query, inputStruct.Filter, inputStruct.Update)
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

	resultTemp, err := searchDocument(&e.searchClient, inputStruct.IndexName, inputStruct.Query, inputStruct.Filter, inputStruct.Size)
	if err != nil {
		return nil, err
	}

	var result []map[string]any
	if inputStruct.Mode == "Source Only" {
		for _, hit := range resultTemp {
			result = append(result, hit.Source)
		}
	} else if inputStruct.Mode == "Hits" {
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

func (e *execution) delete(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct DeleteInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	err = deleteDocument(&e.deleteClient, inputStruct.IndexName, inputStruct.Query, inputStruct.Filter)
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

func (e *execution) deleteIndex(in *structpb.Struct) (*structpb.Struct, error) {
	var inputStruct DeleteIndexInput
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
	}

	err = deleteIndex(&e.deleteIndexClient, inputStruct.IndexName)
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
