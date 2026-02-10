package healthapi

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type APIResponse struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	Query     string       `bson:"query"`
	Data      bson.M       `bson:"data"`
	FetchedAt time.Time    `bson:"fetched_at"`
	ExpiresAt time.Time    `bson:"expires_at"`
}

type Store struct {
	responses *mongo.Collection
	metadata  *mongo.Collection
	errors    *mongo.Collection
}

func NewStore(db *mongo.Database) *Store {
	return &Store{
		responses: db.Collection("responses"),
		metadata:  db.Collection("metadata"),
		errors:    db.Collection("errors"),
	}
}

func (s *Store) CreateIndexes(ctx context.Context) error {
	_, err := s.responses.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "query", Value: 1},
			{Key: "fetched_at", Value: -1},
		},
	})
	if err != nil {
		return err
	}

	_, err = s.responses.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	})
	return err
}

func (s *Store) GetCachedResponse(ctx context.Context, query string) (*APIResponse, error) {
	var response APIResponse
	err := s.responses.FindOne(ctx, bson.D{
		{Key: "query", Value: query},
		{Key: "expires_at", Value: bson.D{{Key: "$gt", Value: time.Now()}}},
	}).Decode(&response)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (s *Store) StoreResponse(ctx context.Context, query string, data map[string]interface{}, ttl time.Duration) error {
	now := time.Now()
	response := APIResponse{
		Query:     query,
		Data:      bson.M(data),
		FetchedAt: now,
		ExpiresAt: now.Add(ttl),
	}

	_, err := s.responses.InsertOne(ctx, response)
	return err
}

func (s *Store) LogError(ctx context.Context, query string, err error) error {
	errorDoc := bson.M{
		"query":     query,
		"error":     err.Error(),
		"timestamp": time.Now(),
	}
	_, err = s.errors.InsertOne(ctx, errorDoc)
	return err
}

func (s *Store) UpdateSyncMetadata(ctx context.Context, lastSync time.Time) error {
	opts := options.UpdateOne().SetUpsert(true)
	_, err := s.metadata.UpdateOne(
		ctx,
		bson.D{{Key: "type", Value: "last_sync"}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "timestamp", Value: lastSync},
		}}},
		opts,
	)
	return err
}

