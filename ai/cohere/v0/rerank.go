package cohere

import (
	"encoding/json"

	cohereSDK "github.com/cohere-ai/cohere-go/v2"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

type rerankInput struct {
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
	ModelName string   `json:"model-name"`
}

type rerankOutput struct {
	Ranking []string    `json:"ranking"`
	Usage   cohereUsage `json:"usage"`
}

func (e *execution) taskRerank(in *structpb.Struct) (*structpb.Struct, error) {

	inputStruct := rerankInput{}
	err := base.ConvertFromStructpb(in, &inputStruct)
	if err != nil {
		return nil, err
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
	for _, rankResult := range resp.Results {
		newRanking = append(newRanking, rankResult.Document.Text)
	}
	outputStruct := rerankOutput{
		Ranking: newRanking,
		Usage: cohereUsage{
			InputTokens:  int(*resp.Meta.BilledUnits.InputTokens),
			OutputTokens: int(*resp.Meta.BilledUnits.OutputTokens),
		},
	}

	outputJSON, err := json.Marshal(outputStruct)
	if err != nil {
		return nil, err
	}
	output := structpb.Struct{}
	err = protojson.Unmarshal(outputJSON, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil

}
