package dedup

import "sync"

type Set struct {
	mu    sync.Mutex
	items map[string]struct{}
}

func NewSet() *Set {
	return &Set{
		items: make(map[string]struct{}),
	}
}

func (s *Set) Add(url string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[url]; ok {
		return false
	}

	s.items[url] = struct{}{}
	return true
}

func (s *Set) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.items)
}
