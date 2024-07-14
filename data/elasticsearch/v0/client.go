package elasticsearch

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"google.golang.org/protobuf/types/known/structpb"
)

func newClient(setup *structpb.Struct) (*esapi.Search, *esapi.Index, *esapi.UpdateByQuery, *esapi.DeleteByQuery, *esapi.IndicesCreate, *esapi.IndicesDelete) {
	cfg := elasticsearch.Config{
		CloudID: getCloudID(setup),
		APIKey:  getAPIKey(setup),
	}

	es, _ := elasticsearch.NewClient(cfg)
	return &es.Search, &es.Index, &es.UpdateByQuery, &es.DeleteByQuery, &es.Indices.Create, &es.Indices.Delete
}

// Need to confirm where the map is
func getAPIKey(setup *structpb.Struct) string {
	return setup.GetFields()["api-key"].GetStringValue()
}

func getCloudID(setup *structpb.Struct) string {
	return setup.GetFields()["cloud-id"].GetStringValue()
}
