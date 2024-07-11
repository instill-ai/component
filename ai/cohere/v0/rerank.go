package cohere

import (
	"fmt"

	cohereSDK "github.com/cohere-ai/cohere-go/v2"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

type rerankInput struct {
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
	ModelName string   `json:"model-name"`
}

type rerankOutput struct {
	Ranking   []string    `json:"ranking"`
	Usage     rerankUsage `json:"usage"`
	Relevance []float64   `json:"relevance"`
}

type rerankUsage struct {
	Search int `json:"search-counts"`
}

func (e *execution) taskRerank(in *structpb.Struct) (*structpb.Struct, error) {

	inputStruct := rerankInput{}
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, fmt.Errorf("error generating input struct: %v", err)
	}

	documents := []*cohereSDK.RerankRequestDocumentsItem{}
	for _, doc := range inputStruct.Documents {
		document := cohereSDK.RerankRequestDocumentsItem{
			String: doc,
		}
		documents = append(documents, &document)
	}

	returnDocument := true
	rankFields := []string{"text"}
	req := cohereSDK.RerankRequest{
		Model:           &inputStruct.ModelName,
		Query:           inputStruct.Query,
		Documents:       documents,
		RankFields:      rankFields,
		ReturnDocuments: &returnDocument,
	}
	resp, err := e.client.generateRerank(req)
	if err != nil {
		return nil, err
	}
	newRanking := []string{}
	relevance := []float64{}
	for _, rankResult := range resp.Results {
		relevance = append(relevance, rankResult.RelevanceScore)
		newRanking = append(newRanking, rankResult.Document.Text)
	}
	bills := resp.Meta.BilledUnits

	outputStruct := rerankOutput{
		Ranking:   newRanking,
		Usage:     rerankUsage{Search: int(*bills.SearchUnits)},
		Relevance: relevance,
	}

	output, err := base.ConvertToStructpb(outputStruct)

	if err != nil {
		return nil, err
	}

	return output, nil

}
