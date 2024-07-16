package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/protobuf/types/known/structpb"
)

func newClient(setup *structpb.Struct) (*mongo.Collection, *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, _ := mongo.Connect(ctx, options.Client().ApplyURI(getURI(setup)))

	db := client.Database(getDatabaseName(setup))
	collection := db.Collection(getCollectionName(setup))

	return collection, db
}

func getURI(setup *structpb.Struct) string {
	return setup.GetFields()["uri"].GetStringValue()
}

func getDatabaseName(setup *structpb.Struct) string {
	return setup.GetFields()["database-name"].GetStringValue()
}

func getCollectionName(setup *structpb.Struct) string {
	return setup.GetFields()["collection-name"].GetStringValue()
}
