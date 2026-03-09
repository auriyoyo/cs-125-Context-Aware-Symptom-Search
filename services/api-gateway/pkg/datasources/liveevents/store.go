package liveevents

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Store struct {
	cfg       *Config
	events    *mongo.Collection
	userCtx   *mongo.Collection
	areaRisks *mongo.Collection
}

func NewStore(db *mongo.Database, cfg *Config) *Store {
	return &Store{
		cfg:       cfg,
		events:    db.Collection(cfg.CollectionName),
		userCtx:   db.Collection("live_event_user_context"),
		areaRisks: db.Collection("live_event_area_risks"),
	}
}

func (s *Store) Events() *mongo.Collection { return s.events }
func (s *Store) Cfg() *Config              { return s.cfg }

func (s *Store) CreateIndexes(ctx context.Context) error {
	eventIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
		{Keys: bson.D{{Key: "event_type", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "zip_code", Value: 1}, {Key: "created_at", Value: -1}}},
	}
	if _, err := s.events.Indexes().CreateMany(ctx, eventIndexes); err != nil {
		return err
	}
	return s.createProjectionIndexes(ctx)
}

func (s *Store) createProjectionIndexes(ctx context.Context) error {
	userCtxIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
	}
	if _, err := s.userCtx.Indexes().CreateMany(ctx, userCtxIndexes); err != nil {
		return err
	}

	areaRiskIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "zip_code", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
	}
	_, err := s.areaRisks.Indexes().CreateMany(ctx, areaRiskIndexes)
	return err
}

func (s *Store) AppendLocationUpdate(ctx context.Context, userID string, loc LocationPayload, source string) error {
	now := time.Now()
	event := LiveEvent{
		EventType: EventTypeLocationUpdate,
		UserID:    userID,
		ZipCode:   loc.ZipCode,
		AreaInfo: AreaInfo{
			Country: loc.Country,
			State:   loc.State,
			City:    loc.City,
		},
		Data: bson.M{
			"source": source,
		},
		CreatedAt: now,
		ExpiresAt: now.Add(s.cfg.LocationRetention()),
	}
	_, err := s.events.InsertOne(ctx, event)
	return err
}

func (s *Store) AppendPublicAlert(ctx context.Context, zipCode string, areaInfo AreaInfo, alert AlertPayload, source string) error {
	now := time.Now()
	event := LiveEvent{
		EventType: EventTypePublicAlert,
		ZipCode:   zipCode,
		AreaInfo:  areaInfo,
		Data: bson.M{
			"provider":      alert.Provider,
			"hazard":        alert.Hazard,
			"severity":      alert.Severity,
			"active_window": bson.M{"start": alert.ActiveWindow.Start, "end": alert.ActiveWindow.End},
			"symptom_tags":  alert.SymptomTags,
			"guidance_url":  alert.GuidanceURL,
			"source":        source,
		},
		CreatedAt: now,
		ExpiresAt: alert.ActiveWindow.End.Add(s.cfg.AlertBuffer()),
	}
	_, err := s.events.InsertOne(ctx, event)
	return err
}
