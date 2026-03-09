package liveevents

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (s *Store) QueryUserEvents(ctx context.Context, userID string, window time.Duration) ([]LiveEvent, error) {
	now := time.Now()
	filter := bson.D{
		{Key: "user_id", Value: userID},
		{Key: "created_at", Value: bson.D{{Key: "$gte", Value: now.Add(-window)}}},
		{Key: "expires_at", Value: bson.D{{Key: "$gt", Value: now}}},
	}
	return s.findEvents(ctx, filter)
}

func (s *Store) QueryAreaAlerts(ctx context.Context, zipCode string) ([]LiveEvent, error) {
	now := time.Now()
	filter := bson.D{
		{Key: "zip_code", Value: zipCode},
		{Key: "event_type", Value: string(EventTypePublicAlert)},
		{Key: "expires_at", Value: bson.D{{Key: "$gt", Value: now}}},
	}
	return s.findEvents(ctx, filter)
}

func (s *Store) findEvents(ctx context.Context, filter bson.D) ([]LiveEvent, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := s.events.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	var events []LiveEvent
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}
	return events, nil
}

func (s *Store) GetUserContext(ctx context.Context, userID string) (*UserContextProjection, error) {
	var proj UserContextProjection
	err := s.userCtx.FindOne(ctx, bson.D{{Key: "user_id", Value: userID}}).Decode(&proj)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &proj, nil
}

func (s *Store) GetAreaRisks(ctx context.Context, zipCode string) (*AreaRiskProjection, error) {
	var proj AreaRiskProjection
	err := s.areaRisks.FindOne(ctx, bson.D{{Key: "zip_code", Value: zipCode}}).Decode(&proj)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &proj, nil
}
