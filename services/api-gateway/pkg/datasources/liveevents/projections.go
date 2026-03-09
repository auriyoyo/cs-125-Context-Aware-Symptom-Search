package liveevents

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (s *Store) UpsertUserContext(ctx context.Context, userID, zipCode string, expiresAt time.Time) error {
	now := time.Now()
	_, err := s.userCtx.UpdateOne(
		ctx,
		bson.D{{Key: "user_id", Value: userID}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "user_id", Value: userID},
			{Key: "last_known_zip", Value: zipCode},
			{Key: "last_updated_at", Value: now},
			{Key: "expires_at", Value: expiresAt},
		}}},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func (s *Store) RefreshAreaRisks(ctx context.Context, zipCode string) error {
	events, err := s.QueryAreaAlerts(ctx, zipCode)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		_, _ = s.areaRisks.DeleteOne(ctx, bson.D{{Key: "zip_code", Value: zipCode}})
		return nil
	}

	risks, latestExpiry := buildActiveRisks(events)
	proj := AreaRiskProjection{
		ZipCode:     zipCode,
		ActiveRisks: risks,
		UpdatedAt:   time.Now(),
		ExpiresAt:   latestExpiry,
	}
	_, err = s.areaRisks.ReplaceOne(
		ctx,
		bson.D{{Key: "zip_code", Value: zipCode}},
		proj,
		options.Replace().SetUpsert(true),
	)
	return err
}

func buildActiveRisks(events []LiveEvent) ([]ActiveRisk, time.Time) {
	risks := make([]ActiveRisk, 0, len(events))
	var latest time.Time
	for _, e := range events {
		risk := ActiveRisk{
			Hazard:      getString(e.Data, "hazard"),
			Severity:    getString(e.Data, "severity"),
			SymptomTags: getStringSlice(e.Data, "symptom_tags"),
			GuidanceURL: getString(e.Data, "guidance_url"),
			ExpiresAt:   e.ExpiresAt,
		}
		risks = append(risks, risk)
		if e.ExpiresAt.After(latest) {
			latest = e.ExpiresAt
		}
	}
	return risks, latest
}

func getString(m bson.M, key string) string {
	v, _ := m[key].(string)
	return v
}

func getStringSlice(m bson.M, key string) []string {
	raw, ok := m[key]
	if !ok {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		return v
	case bson.A:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}
