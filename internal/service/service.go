// Package service provides the cache service layer.
package service

import (
	"context"
	"log"
	"time"

	"github.com/tharunn0/blitzdb/internal/clock"
	"github.com/tharunn0/blitzdb/internal/db"
	"github.com/tharunn0/blitzdb/internal/metrics"

	// "github.com/tharunn0/blitzdb/internal/shard"
	"github.com/tharunn0/blitzdb/internal/store"
	"github.com/tharunn0/blitzdb/pkg/types"
)

const compressionThreshold = 256 // bytes - only compress values larger than this

type Service struct {
	store   store.Store
	clock   *clock.Clock
	metrics *metrics.CacheMetrics
	config  *types.Config
}

func NewService(cfg *types.Config) *Service {
	clk := clock.NewClock()
	// store := shard.NewStore(clk)
	store := db.InitDB()
	m := &metrics.CacheMetrics{}

	s := &Service{
		store:   store,
		clock:   clk,
		metrics: m,
		config:  cfg,
	}
	return s
}

func (s *Service) Set(key string, value []byte, ttl uint64) error {
	log.Printf("Set req [ key : %s | value: %s ]\n", key, string(value))
	if ttl == 0 {
		ttl = s.config.DefaultTTL
	}

	s.store.Set(key, value, ttl)
	s.metrics.IncSet()
	return nil
}

func (s *Service) Get(key string) ([]byte, bool) {

	log.Printf("Get req [ key : %s ]\n", key)

	value, found := s.store.Get(key)
	if !found {
		s.metrics.IncMiss()
		return nil, false
	}
	s.metrics.IncHit()

	return value, true
}

func (s *Service) Delete(key string) {
	s.store.Delete(key)
	s.metrics.IncDel()
}

func (s *Service) Janitor(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.store.Janitor()
		}
	}
}

func (s *Service) Metrics() *metrics.CacheMetrics {
	return s.metrics.Get()
}

// Stop gracefully stops the service and cleans up resources.
func (s *Service) Stop() {
	s.clock.Stop()
}
