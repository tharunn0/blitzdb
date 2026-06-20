// Package sharded implements a sharded in-memory store with lock striping.
package shard

import (
	"log"
	"sync"

	"github.com/tharunn0/blitzdb/internal/clock"
	"github.com/tharunn0/blitzdb/pkg/types"

	"github.com/cespare/xxhash/v2"
)

const NumShards = 256

// Compile-time check that NumShards is a power of 2
// If NumShards is a power of 2, (NumShards & (NumShards - 1)) == 0
// This creates a zero-sized array type, which is valid only if the expression is 0
type _ [NumShards & (NumShards - 1)]struct{}

type Shard struct {
	mu    sync.RWMutex
	items map[string]*types.Entry
}

func NewShard() *Shard {
	return &Shard{
		items: make(map[string]*types.Entry, 1024), // preallocate
	}
}

func (s *Shard) Set(key string, value []byte, expireTick uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := &types.Entry{Value: value, ExpireTick: expireTick}
	s.items[key] = entry
}

func (s *Shard) Get(key string, nowTick uint64) ([]byte, bool) {
	s.mu.RLock()
	entry, exists := s.items[key]
	if !exists {
		s.mu.RUnlock()
		return nil, false
	}

	if entry.ExpireTick < nowTick {
		s.mu.RUnlock()
		s.mu.Lock()
		if entry, exists := s.items[key]; exists && entry.ExpireTick < nowTick {
			delete(s.items, key)
		}
		s.mu.Unlock()
		return nil, false
	}

	value := entry.Value
	s.mu.RUnlock()
	return value, true
}

func (s *Shard) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.items, key)
}

func (s *Shard) Janitor(nowTick uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for k, e := range s.items {
		if e.ExpireTick < nowTick { // Fixed: use < instead of <=
			delete(s.items, k)
		}
	}
}

type Store struct {
	shards [NumShards]*Shard
	clock  *clock.Clock
}

// NewStore creates a new sharded store.
func NewStore(clock *clock.Clock) *Store {
	s := &Store{clock: clock}

	for i := 0; i < NumShards; i++ {
		s.shards[i] = NewShard()
	}
	return s
}

func (st *Store) shardIndex(key string) int {
	h := xxhash.Sum64String(key)
	return int(h & (NumShards - 1))
}

func (st *Store) Set(key string, value []byte, ttlMinutes uint64) {
	now := st.clock.Now()
	expire := now + ttlMinutes
	shard := st.shards[st.shardIndex(key)]
	shard.Set(key, value, expire)
}

func (st *Store) Get(key string) ([]byte, bool) {
	now := st.clock.Now()
	shard := st.shards[st.shardIndex(key)]
	return shard.Get(key, now)
}

func (st *Store) Delete(key string) {
	shard := st.shards[st.shardIndex(key)]
	shard.Delete(key)
}

// Janitor performs cleanup across all shards with proper synchronization.
func (st *Store) Janitor() {
	now := st.clock.Now()
	var wg sync.WaitGroup

	for _, shard := range st.shards {
		wg.Add(1)
		go func(s *Shard) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Janitor panic in shard: %v", r)
				}
			}()
			s.Janitor(now)
		}(shard)
	}

	wg.Wait()
}
