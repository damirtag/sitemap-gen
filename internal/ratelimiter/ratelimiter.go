package ratelimiter

import (
	"context"
	"net/url"
	"sync"

	"golang.org/x/time/rate"
)

type DomainLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	rps      float64
	burst    int
}

func NewDomainLimiter(rps float64, burst int) *DomainLimiter {
	return &DomainLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rps,
		burst:    burst,
	}
}

func (d *DomainLimiter) Wait(ctx context.Context, rawURL string) error {
	u, _ := url.Parse(rawURL)
	d.mu.Lock()
	l, ok := d.limiters[u.Hostname()]
	if !ok {
		l = rate.NewLimiter(rate.Limit(d.rps), d.burst)
		d.limiters[u.Hostname()] = l
	}
	d.mu.Unlock()
	return l.Wait(ctx)
}
