//go:build ingest

package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/auriyoyo/cs-125-Context-Aware-Symptom-Search/services/api-gateway/pkg/config"
	"github.com/auriyoyo/cs-125-Context-Aware-Symptom-Search/services/api-gateway/pkg/database"
	"github.com/auriyoyo/cs-125-Context-Aware-Symptom-Search/services/api-gateway/pkg/datasources/healthapi"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Dummy ingest script. Runs API call to Clinical Tables conditions database with "Cold"
// query and adds resulting JSON to clinical_tables/conditions in MongoDB.
func main() {
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if config.MongoDBURI == "" {
		log.Fatal("MONGODB_URI is required for ingest")
	}
	if err := database.Connect(config.MongoDBURI); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer database.Disconnect()

	cfg := healthapi.LoadConfig()
	client := healthapi.NewClient(cfg.APIBaseURL, cfg.APIKey)

	req := &healthapi.QueryRequest{
		Terms: []string{"Cold"},
	}

	resp, err := client.QueryWithRequest(req)
	if err != nil {
		log.Fatalf("Health API query failed: %v", err)
	}

	out, _ := json.MarshalIndent(resp, "", "  ")
	log.Println("Response:")
	log.Println(string(out))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	doc := bson.M{
		"terms":      req.Terms,
		"response":   resp,
		"fetched_at": time.Now(),
	}

	coll := database.GetDatabase("clinical_tables").Collection("conditions")
	if _, err := coll.InsertOne(ctx, doc); err != nil {
		log.Fatalf("Failed to store results: %v", err)
	}
	log.Println("Stored results in clinical_tables.conditions")
}
