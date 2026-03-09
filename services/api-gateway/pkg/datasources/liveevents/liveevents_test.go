package liveevents

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// BuildAreaKey removed

func TestConfigDefaults(t *testing.T) {
	cfg := LoadConfig()
	if cfg.DatabaseName != "live_events_db" {
		t.Errorf("default DatabaseName = %q, want live_events_db", cfg.DatabaseName)
	}
	if cfg.CollectionName != "live_events" {
		t.Errorf("default CollectionName = %q, want live_events", cfg.CollectionName)
	}
	if cfg.LocationRetentionDays != 30 {
		t.Errorf("default LocationRetentionDays = %d, want 30", cfg.LocationRetentionDays)
	}
	if cfg.AlertBufferHours != 24 {
		t.Errorf("default AlertBufferHours = %d, want 24", cfg.AlertBufferHours)
	}
	if cfg.EnableSearchEnrichment {
		t.Error("default EnableSearchEnrichment should be false")
	}
	if cfg.EnableProjections {
		t.Error("default EnableProjections should be false")
	}
}

func TestConfigRetentionDurations(t *testing.T) {
	cfg := &Config{LocationRetentionDays: 7, AlertBufferHours: 12}
	if cfg.LocationRetention() != 7*24*time.Hour {
		t.Errorf("LocationRetention() = %v, want %v", cfg.LocationRetention(), 7*24*time.Hour)
	}
	if cfg.AlertBuffer() != 12*time.Hour {
		t.Errorf("AlertBuffer() = %v, want %v", cfg.AlertBuffer(), 12*time.Hour)
	}
}

func TestBuildActiveRisks(t *testing.T) {
	now := time.Now()
	events := []LiveEvent{
		{
			ExpiresAt: now.Add(2 * time.Hour),
			Data: bson.M{
				"hazard":       "air_quality",
				"severity":     "unhealthy",
				"symptom_tags": bson.A{"cough", "headache"},
				"guidance_url": "https://example.com",
			},
		},
		{
			ExpiresAt: now.Add(6 * time.Hour),
			Data: bson.M{
				"hazard":       "pollen",
				"severity":     "high",
				"symptom_tags": bson.A{"sneezing"},
				"guidance_url": "https://example.com/pollen",
			},
		},
	}

	risks, latest := buildActiveRisks(events)
	if len(risks) != 2 {
		t.Fatalf("got %d risks, want 2", len(risks))
	}
	if risks[0].Hazard != "air_quality" {
		t.Errorf("risks[0].Hazard = %q, want air_quality", risks[0].Hazard)
	}
	if len(risks[0].SymptomTags) != 2 {
		t.Errorf("risks[0] symptom tags count = %d, want 2", len(risks[0].SymptomTags))
	}
	if !latest.Equal(now.Add(6 * time.Hour)) {
		t.Errorf("latest expiry mismatch")
	}
}

func TestBuildWarnings_SymptomOverlap(t *testing.T) {
	risks := []ActiveRisk{
		{Hazard: "air_quality", Severity: "unhealthy", SymptomTags: []string{"cough", "headache"}},
		{Hazard: "pollen", Severity: "high", SymptomTags: []string{"sneezing", "runny nose"}},
	}
	warnings := buildWarnings(risks, []string{"cough", "sneezing"})
	if len(warnings) != 2 {
		t.Fatalf("got %d warnings, want 2", len(warnings))
	}
	if warnings[0].MatchedSymptoms[0] != "cough" {
		t.Errorf("expected cough in matched symptoms")
	}
}

func TestBuildWarnings_NoOverlap(t *testing.T) {
	risks := []ActiveRisk{
		{Hazard: "air_quality", Severity: "moderate", SymptomTags: []string{"cough"}},
	}
	warnings := buildWarnings(risks, []string{"fever"})
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings for no overlap, got %d", len(warnings))
	}
}

func TestFindOverlap(t *testing.T) {
	set := map[string]bool{"cough": true, "fever": true}
	got := findOverlap([]string{"Cough", "headache", "FEVER"}, set)
	if len(got) != 2 {
		t.Errorf("findOverlap returned %d matches, want 2", len(got))
	}
}

func TestGetString(t *testing.T) {
	m := bson.M{"key": "value", "num": 42}
	if getString(m, "key") != "value" {
		t.Error("getString failed for existing key")
	}
	if getString(m, "num") != "" {
		t.Error("getString should return empty for non-string value")
	}
	if getString(m, "missing") != "" {
		t.Error("getString should return empty for missing key")
	}
}

func TestGetStringSlice(t *testing.T) {
	m := bson.M{
		"tags":  bson.A{"a", "b"},
		"plain": []string{"x", "y"},
		"num":   42,
	}
	if got := getStringSlice(m, "tags"); len(got) != 2 {
		t.Errorf("getStringSlice bson.A: got %d, want 2", len(got))
	}
	if got := getStringSlice(m, "num"); got != nil {
		t.Errorf("getStringSlice non-slice: expected nil, got %v", got)
	}
	if got := getStringSlice(m, "missing"); got != nil {
		t.Errorf("getStringSlice missing: expected nil, got %v", got)
	}
}
