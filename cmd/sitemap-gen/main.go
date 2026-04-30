package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"sitemap-gen/internal/dedup"
	"sitemap-gen/internal/metrics"
	"sitemap-gen/internal/ratelimiter"
	"sitemap-gen/internal/report"
	"sitemap-gen/internal/scheduler"
	"sitemap-gen/internal/sitemap"
	"sitemap-gen/internal/worker"
)

func main() {
	targetURL := flag.String("url", "https://go.dev", "seed URL to crawl")
	output := flag.String("out", "sitemap.xml", "output file path")
	workers := flag.Int("workers", 10, "number of concurrent workers")
	flag.Parse()

	// pprof
	go func() { log.Println(http.ListenAndServe("localhost:6060", nil)) }()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// extract hostname for domain scoping
	parsed, err := url.Parse(*targetURL)
	if err != nil {
		log.Fatalf("invalid URL: %v", err)
	}
	allowedHost := parsed.Hostname()

	queue := scheduler.NewTaskQueue(1000)
	visited := dedup.NewSet()
	metrics := metrics.NewMetrics()
	sm := sitemap.NewMap()

	var wg sync.WaitGroup
	limiter := ratelimiter.NewDomainLimiter(5, 10)
	worker.StartWorkers(ctx, *workers, queue, metrics, visited, sm, allowedHost, limiter, &wg)

	start := time.Now()

	if visited.Add(*targetURL) {
		queue.AddTask(scheduler.Task{URL: *targetURL, Depth: 0})
	}

	// why we doing this in a separate goroutine?
	// so we can block on <-ctx.Done() in main and still
	// allow workers to finish their current tasks before shutting down.
	go func() {
		wg.Wait()
		log.Println("all work done, shutting down...")
		stop() // cancels ctx, unblocks <-ctx.Done()
	}()

	<-ctx.Done()
	log.Println("shutdown signal received")

	queue.Stop()

	log.Println("waiting for workers...")
	wg.Wait()

	if err := sm.Write(*output); err != nil {
		log.Fatalf("failed to write sitemap: %v", err)
	}

	report.PrintShutdown(report.Stats{
		TargetURL:   *targetURL,
		OutputFile:  *output,
		Total:       metrics.GetTotalCount(),
		Success:     metrics.GetSuccessCount(),
		Errors:      metrics.GetErrorCount(),
		UniqueURLs:  visited.Count(),
		SitemapURLs: sm.Count(),
		Elapsed:     time.Since(start),
		Workers:     *workers,
	})
}
