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
	Criteria  map[string]any `json:"criteria"`
	Query     string         `json:"query"`
	IndexName string         `json:"index-name"`
}

type UpdateOutput struct {
	Status string `json:"status"`
}

type SearchInput struct {
	Criteria  map[string]any `json:"criteria"`
	Query     string         `json:"query"`
	IndexName string         `json:"index-name"`
}

type SearchOutput struct {
	Documents []Hit  `json:"documents"`
	Status    string `json:"status"`
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
	Criteria  map[string]any `json:"criteria"`
	Query     string         `json:"query"`
	IndexName string         `json:"index-name"`
}

type DeleteOutput struct {
	Status string `json:"status"`
}

// Index document into Elasticsearch
func indexDocument(es *esapi.Index, indexName string, data map[string]interface{}) error {
	// Serialize data to JSON
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	esClient := ESIndex(*es)

	// Index document using elasticsearch.Client.Index method
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

// Search document from Elasticsearch
func searchDocument(es *esapi.Search, indexName string, query string, criteria map[string]any) ([]Hit, error) {
	criteriaQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"match": criteria,
		},
	}

	var body io.Reader
	if criteria != nil {
		// Serialize data to JSON
		criteriaJSON, err := json.Marshal(criteriaQuery)
		if err != nil {
			return nil, err
		}

		body = strings.NewReader(string(criteriaJSON))
	} else {
		body = nil
	}

	esClient := ESSearch(*es)

	res, err := esClient(func(r *esapi.SearchRequest) {
		r.Index = []string{indexName}
		r.Body = body
		r.Query = query
		r.TrackTotalHits = true
	})

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching document: %s", res.Status())
	}

	var response SearchResponse
	// Deserialize response body
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Hits.Hits, nil
}

// Update document in Elasticsearch
func updateDocument(es *esapi.UpdateByQuery, indexName string, query string, criteria map[string]interface{}, update map[string]interface{}) error {
	// Create the update by query request body
	updateByQueryReq := map[string]interface{}{
		"query": map[string]interface{}{
			"match": criteria,
		},
		"script": map[string]interface{}{
			"source": "for (entry in params.entry.entrySet()) { ctx._source[entry.getKey()] = entry.getValue() }",
			"lang":   "painless",
			"params": map[string]interface{}{
				"entry": update,
			},
		},
	}

	var body io.Reader
	if criteria != nil {
		// Serialize data to JSON
		updateJSON, err := json.Marshal(updateByQueryReq)
		if err != nil {
			return err
		}

		body = strings.NewReader(string(updateJSON))
	} else {
		body = nil
	}

	esClient := ESUpdate(*es)

	// Update documents using elasticsearch.Client.UpdateByQuery method
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

// Delete document from Elasticsearch
func deleteDocument(es *esapi.DeleteByQuery, indexName string, query string, criteria map[string]interface{}) error {
	// Create the delete by query request body
	deleteByQueryReq := map[string]interface{}{
		"query": map[string]interface{}{
			"match": criteria,
		},
	}

	var body io.Reader
	if criteria != nil {
		// Serialize data to JSON
		criteriaJSON, err := json.Marshal(deleteByQueryReq)
		if err != nil {
			return err
		}

		body = bytes.NewReader(criteriaJSON)
	} else {
		body = nil
	}

	esClient := ESDelete(*es)

	// Delete documents using elasticsearch.Client.DeleteByQuery method
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

	err = updateDocument(&e.updateClient, inputStruct.IndexName, inputStruct.Query, inputStruct.Criteria, inputStruct.Update)
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

	result, err := searchDocument(&e.searchClient, inputStruct.IndexName, inputStruct.Query, inputStruct.Criteria)
	if err != nil {
		return nil, err
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

	err = deleteDocument(&e.deleteClient, inputStruct.IndexName, inputStruct.Query, inputStruct.Criteria)
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
