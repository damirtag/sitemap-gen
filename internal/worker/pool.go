package worker

import (
	"context"
	"log"
	"runtime"
	"sync"
	"time"

	"sitemap-gen/internal/dedup"
	"sitemap-gen/internal/fetcher"
	"sitemap-gen/internal/metrics"
	"sitemap-gen/internal/parser"
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
	wg *sync.WaitGroup,
) {
	for i := 0; i < n; i++ {
		wg.Add(1)
		go workerLoop(ctx, i, q, visited, m, sm, allowedHost, wg)
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
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Printf("worker %d stopping...\n", id)
			return

		case task, ok := <-q.Chan():
			if !ok {
				log.Printf("worker %d queue closed\n", id)
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

			time.Sleep(200 * time.Millisecond)
		}
	}
}
