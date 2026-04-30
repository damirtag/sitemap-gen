package dedup

import (
	"hash/fnv"
	"sync"
)

const shards = 32

type Set struct {
	shards [shards]shard
}

type shard struct {
	mu    sync.RWMutex
	items map[string]struct{}
}

func (s *Set) getShard(url string) *shard {
	h := fnv.New32a()
	h.Write([]byte(url))
	return &s.shards[h.Sum32()%shards]
}

func NewSet() *Set {
	s := &Set{}
	for i := range s.shards {
		s.shards[i].items = make(map[string]struct{})
	}
	return s
}

func (s *Set) Add(url string) bool {
	shard := s.getShard(url)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	if _, ok := shard.items[url]; ok {
		return false
	}

	shard.items[url] = struct{}{}
	return true
}

func (s *Set) Count() int {
	total := 0
	for i := 0; i < shards; i++ {
		s.shards[i].mu.RLock()
		total += len(s.shards[i].items)
		s.shards[i].mu.RUnlock()
	}
	return total
}
