package liveevents

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type changeDoc struct {
	OperationType string    `bson:"operationType"`
	FullDocument  LiveEvent `bson:"fullDocument"`
}

func (s *Store) RunProjectionWorker(ctx context.Context) error {
	if !s.cfg.EnableProjections {
		log.Println("[liveevents] projections disabled, worker not started")
		return nil
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "operationType", Value: "insert"}}}},
	}
	cs, err := s.events.Watch(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cs.Close(ctx)

	log.Println("[liveevents] projection worker started")

	for cs.Next(ctx) {
		var doc changeDoc
		if err := cs.Decode(&doc); err != nil {
			log.Printf("[liveevents] decode change event: %v", err)
			continue
		}
		s.processEvent(ctx, &doc.FullDocument)
	}

	if err := cs.Err(); err != nil && ctx.Err() == nil {
		return err
	}
	return nil
}

func (s *Store) processEvent(ctx context.Context, event *LiveEvent) {
	switch event.EventType {
	case EventTypeLocationUpdate:
		if err := s.UpsertUserContext(ctx, event.UserID, event.ZipCode, event.ExpiresAt); err != nil {
			log.Printf("[liveevents] upsert user context: %v", err)
		}
	case EventTypePublicAlert:
		if err := s.RefreshAreaRisks(ctx, event.ZipCode); err != nil {
			log.Printf("[liveevents] refresh area risks: %v", err)
		}
	}
}
