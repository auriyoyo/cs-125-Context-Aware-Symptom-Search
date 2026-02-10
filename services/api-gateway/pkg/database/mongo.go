package database

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	Client   *mongo.Client
	Database *mongo.Database
)

func Connect(uri string) error {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	if err := client.Ping(context.TODO(), nil); err != nil {
		return err
	}

	Client = client
	log.Println("Connected to MongoDB")
	return nil
}

func Disconnect() error {
	if Client == nil {
		return nil
	}
	return Client.Disconnect(context.TODO())
}

func GetDatabase(name string) *mongo.Database {
	return Client.Database(name)
}
