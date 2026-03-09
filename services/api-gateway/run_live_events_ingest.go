//go:build live_events_ingest

package main

import (
	"context"
	"log"
	"time"

	"github.com/auriyoyo/cs-125-Context-Aware-Symptom-Search/services/api-gateway/pkg/config"
	"github.com/auriyoyo/cs-125-Context-Aware-Symptom-Search/services/api-gateway/pkg/database"
	"github.com/auriyoyo/cs-125-Context-Aware-Symptom-Search/services/api-gateway/pkg/datasources/liveevents"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	if err := database.Connect(config.MongoDBURI); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer database.Disconnect()

	cfg := liveevents.LoadConfig()
	db := database.GetDatabase(cfg.DatabaseName)
	store := liveevents.NewStore(db, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := store.CreateIndexes(ctx); err != nil {
		log.Fatalf("Failed to create indexes: %v", err)
	}

	// Replace these sample alerts with real provider API calls.
	sampleAlerts := []struct {
		zipCode  string
		areaInfo liveevents.AreaInfo
		alert    liveevents.AlertPayload
	}{
		{
			zipCode: "92617",
			areaInfo: liveevents.AreaInfo{
				Country: "US",
				State:   "CA",
				City:    "Irvine",
			},
			alert: liveevents.AlertPayload{
				Provider:     "aqi_api",
				Hazard:       "air_quality",
				Severity:     "unhealthy_for_sensitive",
				ActiveWindow: liveevents.ActiveWindow{Start: time.Now(), End: time.Now().Add(12 * time.Hour)},
				SymptomTags:  []string{"cough", "shortness of breath", "wheezing", "headache"},
				GuidanceURL:  "https://www.airnow.gov/",
			},
		},
		{
			zipCode: "92617",
			areaInfo: liveevents.AreaInfo{
				Country: "US",
				State:   "CA",
				City:    "Irvine",
			},
			alert: liveevents.AlertPayload{
				Provider:     "pollen_api",
				Hazard:       "pollen",
				Severity:     "high",
				ActiveWindow: liveevents.ActiveWindow{Start: time.Now(), End: time.Now().Add(24 * time.Hour)},
				SymptomTags:  []string{"sneezing", "runny nose", "itchy eyes", "congestion"},
				GuidanceURL:  "https://www.pollen.com/",
			},
		},
	}

	for _, sa := range sampleAlerts {
		if err := store.AppendPublicAlert(ctx, sa.zipCode, sa.areaInfo, sa.alert, sa.alert.Provider); err != nil {
			log.Printf("Failed to ingest alert %s/%s: %v", sa.zipCode, sa.alert.Hazard, err)
			continue
		}
		log.Printf("Ingested %s alert for %s", sa.alert.Hazard, sa.zipCode)
	}

	log.Println("Live events ingest complete")
}
