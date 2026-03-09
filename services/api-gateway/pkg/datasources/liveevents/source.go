package liveevents

import (
	"context"
	"log"
)

type Source struct {
	store  *Store
	cancel context.CancelFunc
}

func NewSource(store *Store) *Source {
	return &Source{store: store}
}

func (s *Source) Name() string         { return "live_events" }
func (s *Source) DatabaseName() string { return s.store.cfg.DatabaseName }
func (s *Source) Query(_ string) error { return nil }

func (s *Source) Start(ctx context.Context) error {
	if err := s.store.CreateIndexes(ctx); err != nil {
		return err
	}

	workerCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	go func() {
		if err := s.store.RunProjectionWorker(workerCtx); err != nil {
			log.Printf("[liveevents] projection worker stopped: %v", err)
		}
	}()

	return nil
}

func (s *Source) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}
