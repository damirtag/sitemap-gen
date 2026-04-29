package dedup

import (
	"sync"
	"testing"
)

func TestAdd_NewURL(t *testing.T) {
	s := NewSet()
	if !s.Add("https://example.com") {
		t.Error("expected true for new URL")
	}
}

func TestAdd_DuplicateURL(t *testing.T) {
	s := NewSet()
	s.Add("https://example.com")
	if s.Add("https://example.com") {
		t.Error("expected false for duplicate URL")
	}
}

func TestCount(t *testing.T) {
	s := NewSet()
	s.Add("https://a.com")
	s.Add("https://b.com")
	s.Add("https://a.com") // duplicate
	if s.Count() != 2 {
		t.Errorf("expected 2, got %d", s.Count())
	}
}

func TestAdd_Concurrent(t *testing.T) {
	s := NewSet()
	var wg sync.WaitGroup

	// 100 goroutines all adding the same URL — only one should win
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Add("https://example.com")
		}()
	}
	wg.Wait()

	if s.Count() != 1 {
		t.Errorf("expected 1 under concurrency, got %d", s.Count())
	}
}
