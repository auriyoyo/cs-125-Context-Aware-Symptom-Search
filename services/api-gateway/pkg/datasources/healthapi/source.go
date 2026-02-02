package healthapi

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/auriyoyo/cs-125-Context-Aware-Symptom-Search/services/api-gateway/pkg/database"
	"github.com/auriyoyo/cs-125-Context-Aware-Symptom-Search/services/api-gateway/pkg/datasources"
)

type Source struct {
	config *Config
	client *Client
	store  *Store
	db     *mongo.Database
	ctx    context.Context
	cancel context.CancelFunc
}

func NewSource() (*Source, error) {
	config := LoadConfig()
	db := database.GetDatabase(config.DatabaseName)

	source := &Source{
		config: config,
		client: NewClient(config.APIBaseURL, config.APIKey),
		store:  NewStore(db),
		db:     db,
	}

	ctx := context.Background()
	if err := source.store.CreateIndexes(ctx); err != nil {
		log.Printf("Warning: Failed to create indexes for %s: %v", source.Name(), err)
	}

	return source, nil
}

func (s *Source) Name() string {
	return "health-api"
}

func (s *Source) DatabaseName() string {
	return s.config.DatabaseName
}

func (s *Source) Query(query string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cached, err := s.store.GetCachedResponse(ctx, query)
	if err != nil {
		return err
	}
	if cached != nil {
		log.Printf("[%s] Using cached response for query: %s", s.Name(), query)
		return nil
	}

	log.Printf("[%s] Fetching from API for query: %s", s.Name(), query)
	data, err := s.client.Query(query)
	if err != nil {
		if logErr := s.store.LogError(ctx, query, err); logErr != nil {
			log.Printf("Failed to log error: %v", logErr)
		}
		return err
	}

	ttl := 24 * time.Hour
	if err := s.store.StoreResponse(ctx, query, data, ttl); err != nil {
		return err
	}

	if err := s.store.UpdateSyncMetadata(ctx, time.Now()); err != nil {
		log.Printf("Warning: Failed to update sync metadata: %v", err)
	}

	log.Printf("[%s] Successfully stored response for query: %s", s.Name(), query)
	return nil
}

func (s *Source) Start(ctx context.Context) error {
	s.ctx, s.cancel = context.WithCancel(ctx)

	log.Printf("[%s] Starting data source", s.Name())

	go s.runPeriodicSync()

	return nil
}

func (s *Source) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}
	log.Printf("[%s] Stopped data source", s.Name())
	return nil
}

func (s *Source) runPeriodicSync() {
	ticker := time.NewTicker(s.config.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			log.Printf("[%s] Running periodic sync", s.Name())
		}
	}
}

var _ datasources.DataSource = (*Source)(nil)
