//go:build !ingest

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/auriyoyo/cs-125-Context-Aware-Symptom-Search/services/api-gateway/pkg/config"
	"github.com/auriyoyo/cs-125-Context-Aware-Symptom-Search/services/api-gateway/pkg/database"
	"github.com/auriyoyo/cs-125-Context-Aware-Symptom-Search/services/api-gateway/pkg/datasources"
	"github.com/auriyoyo/cs-125-Context-Aware-Symptom-Search/services/api-gateway/pkg/datasources/localjson"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if config.MongoDBURI != "" {
	if err := database.Connect(config.MongoDBURI); err != nil {
		log.Printf("MongoDB not available, continuing without it: %v", err)
		} else {
			defer database.Disconnect()
		}
	} else {
		log.Println("MongoDB URI empty, continuing without MongoDB")
	}


	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var sources []datasources.DataSource

	// commented out until conditions source fixed
	/*
	conditionsSource, err := conditions.NewSource()
	if err != nil {
		log.Fatalf("Failed to initialize conditions source: %v", err)
	}
	sources = append(sources, conditionsSource)
	*/

	for _, source := range sources {
		if err := source.Start(ctx); err != nil {
			log.Fatalf("Failed to start data source %s: %v", source.Name(), err)
		}
		log.Printf("Started data source: %s (database: %s)", source.Name(), source.DatabaseName())
	}

	store, err := localjson.Load("../dataset-ingest/diseases.json")
	if err != nil {
		log.Fatalf("Failed to load diseases.json: %v", err)
	}
	go startHTTP(store)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Println("API Gateway running. Press Ctrl+C to stop...")
	<-sigChan

	log.Println("Shutting down...")
	for _, source := range sources {
		if err := source.Stop(); err != nil {
			log.Printf("Error stopping data source %s: %v", source.Name(), err)
		}
	}
}
