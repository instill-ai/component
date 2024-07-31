package weaviate

import (
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/auth"
	"google.golang.org/protobuf/types/known/structpb"
)

func newClient(setup *structpb.Struct) *WeaviateClient {
	cfg := weaviate.Config{
		Host:       getURL(setup),
		Scheme:     "https",
		AuthConfig: auth.ApiKey{Value: getAPIKey(setup)},
		Headers:    nil,
	}

	client, err := weaviate.NewClient(cfg)
	if err != nil {
		return &WeaviateClient{}
	}

	return &WeaviateClient{
		dataAPICreatorClient:             client.Data().Creator(),
		graphQLAPIGetClient:              client.GraphQL().Get(),
		batchAPIDeleterClient:            client.Batch().ObjectsBatchDeleter(),
		batchAPIBatcherClient:            client.Batch().ObjectsBatcher(),
		schemaAPIDeleterClient:           client.Schema().ClassDeleter(),
		schemaAPIClassGetterClient:       client.Schema().ClassGetter(),
		graphQLNearVectorArgumentBuilder: client.GraphQL().NearVectorArgBuilder(),
	}
}

func getURL(setup *structpb.Struct) string {
	return setup.GetFields()["url"].GetStringValue()
}

func getAPIKey(setup *structpb.Struct) string {
	return setup.GetFields()["api-key"].GetStringValue()
}
