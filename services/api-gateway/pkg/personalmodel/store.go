package personalmodel

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Model struct {
	UserID         string    `bson:"user_id"`
	RecentSymptoms []string  `bson:"recent_symptoms"`
	UpdatedAt      time.Time `bson:"updated_at"`
}

type Store struct {
	coll *mongo.Collection
}

func NewStore(db *mongo.Database) *Store {
	return &Store{coll: db.Collection("personal_models")}
}

func (s *Store) Get(ctx context.Context, userID string) (*Model, error) {
	var m Model
	err := s.coll.FindOne(ctx, bson.D{{Key: "user_id", Value: userID}}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return &Model{UserID: userID, RecentSymptoms: []string{}}, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *Store) AppendSymptoms(ctx context.Context, userID string, symptoms []string) error {
	if userID == "" || len(symptoms) == 0 {
		return nil
	}
	_, err := s.coll.UpdateOne(
		ctx,
		bson.D{{Key: "user_id", Value: userID}},
		bson.D{
			{Key: "$setOnInsert", Value: bson.D{{Key: "user_id", Value: userID}}},
			{Key: "$set", Value: bson.D{{Key: "updated_at", Value: time.Now()}}},
			{Key: "$push", Value: bson.D{{Key: "recent_symptoms", Value: bson.D{
				{Key: "$each", Value: symptoms},
				{Key: "$slice", Value: -20},
			}}}},
		},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}
