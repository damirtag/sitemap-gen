package worker

import (
	"context"
	"log"
	"runtime"
	"sync"
	"sync/atomic"

	"sitemap-gen/internal/dedup"
	"sitemap-gen/internal/fetcher"
	"sitemap-gen/internal/metrics"
	"sitemap-gen/internal/parser"
	"sitemap-gen/internal/ratelimiter"
	"sitemap-gen/internal/scheduler"
	"sitemap-gen/internal/sitemap"
)

const maxDepth = 3

func StartWorkers(
	ctx context.Context,
	n int,
	q *scheduler.TaskQueue,
	m *metrics.Metrics,
	visited *dedup.Set,
	sm *sitemap.Map,
	allowedHost string,
	limiter *ratelimiter.DomainLimiter,
	wg *sync.WaitGroup,
) {
	idle := atomic.Int32{}

	for i := 0; i < n; i++ {
		wg.Add(1)
		// run worker loop in separate goroutine
		go workerLoop(ctx, i, q, visited, m, sm, allowedHost, limiter, wg, &idle, int32(n))
	}
}

func workerLoop(
	ctx context.Context,
	id int,
	q *scheduler.TaskQueue,
	visited *dedup.Set,
	m *metrics.Metrics,
	sm *sitemap.Map,
	allowedHost string,
	limiter *ratelimiter.DomainLimiter,
	wg *sync.WaitGroup,
	idle *atomic.Int32,
	totalWorkers int32,
) {
	defer wg.Done()

	for {
		idle.Add(1)
		if idle.Load() == totalWorkers && len(q.Chan()) == 0 {
			idle.Add(-1)
			log.Printf("worker %d: all workers idle, shutting down...\n", id)
			q.Stop()
			q.Close()
			return
		}

		select {
		case <-ctx.Done():
			idle.Add(-1)
			log.Printf("worker %d stopping...\n", id)
			return

		case task, ok := <-q.Chan():
			idle.Add(-1)
			if !ok {
				log.Printf("worker %d queue closed\n", id)
				return
			}

			if err := limiter.Wait(ctx, task.URL); err != nil {
				return
			}

			resp, err := fetcher.Fetch(ctx, task.URL)
			if err != nil {
				m.IncErrors()
				continue
			}

			// record in sitemap
			sm.Add(resp.FinalURL)
			m.IncSuccess()

			log.Printf("[worker %d] depth=%d url=%s goroutines=%d\n",
				id, task.Depth, task.URL, runtime.NumGoroutine())

			// don't go deeper than the limit
			if task.Depth >= maxDepth {
				continue
			}

			links, err := parser.ExtractLinks(resp.Body, resp.FinalURL, allowedHost)
			if err != nil {
				m.IncErrors()
				continue
			}

			for _, link := range links {
				if visited.Add(link) {
					q.AddTask(scheduler.Task{URL: link, Depth: task.Depth + 1})
				}
			}
		}
	}
}
