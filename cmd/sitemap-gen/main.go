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
	worker.StartWorkers(ctx, *workers, queue, metrics, visited, sm, allowedHost, &wg)

	if visited.Add(*targetURL) {
		queue.AddTask(scheduler.Task{URL: *targetURL, Depth: 0})
	}

	<-ctx.Done()
	log.Println("shutdown signal received")

	queue.Stop()
	time.Sleep(1 * time.Second)
	queue.Close()

	log.Println("waiting for workers...")
	wg.Wait()

	// write the sitemap
	if err := sm.Write(*output); err != nil {
		log.Printf("failed to write sitemap: %v\n", err)
	} else {
		log.Printf("sitemap written to %s (%d URLs)\n", *output, sm.Count())
	}

	log.Printf("total: %d | errors: %d | unique URLs: %d\n",
		metrics.GetTotalCount(), metrics.GetErrorCount(), visited.Count())
	log.Println("done")
}
