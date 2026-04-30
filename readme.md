# sitemap-gen

![Tests](https://github.com/damirtag/sitemap-gen/actions/workflows/ci.yml/badge.svg)

A concurrent sitemap generator written in Go. Feed it a URL, it crawls the entire domain and outputs a valid `sitemap.xml` — respecting depth limits, deduplicating URLs, and shutting down gracefully

Built to explore goroutines, worker pools, context propagation, and graceful shutdown patterns

### What is a Sitemaps?
<blockquote><i>A sitemap is a file that lists the pages of a website to help search engines and users navigate the site more effectively. It can be in various formats, such as XML for search engines or HTML for users, and it provides information about the structure and content of the website.</i></blockquote>

## Features

- **Concurrent crawling** — configurable worker pool, each worker runs in its own goroutine
- **Auto shutdown** — detects when all workers are idle and the queue is empty, just exists
- **Domain scoping** — stays within the seed domain, never drifts to external sites
- **Depth limiting** — won't crawl infinitely deep (default: 3 levels)
- **URL normalization** — resolves relative links, strips fragments, deduplicates
- **Per-domain rate limiting** — configurable requests/sec with burst support, prevents bans
- **Graceful shutdown** — `SIGINT`/`SIGTERM` drains in-flight tasks and waits for all workers
- **Sharded dedup** — 32-shard concurrent URL set, minimizes lock contention at scale
- **Valid sitemap output** — writes a spec-compliant `sitemap.xml` with `<loc>`, `<lastmod>`, `<changefreq>`, `<priority>`
- **Shutdown report** — ASCII summary of crawl stats printed on exit
- **pprof** — profiling endpoint exposed at `localhost:6060`

## Installation

```bash
git clone https://github.com/damirtag/sitemap-gen
cd sitemap-gen
go build -o sitemap-gen ./cmd/sitemap-gen/
```

Requires Go 1.21+.

## Usage

```bash
./sitemap-gen -url https://damir.top
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `-url` | `https://go.dev` | Seed URL to start crawling from |
| `-out` | `sitemap.xml` | Output file path |
| `-workers` | `10` | Number of concurrent workers |

### Examples

```bash
# crawl a site with 20 workers, save to a custom path
./sitemap-gen -url https://example.com -workers 20 -out ./output/sitemap.xml

# run directly without building
go run ./cmd/sitemap-gen/ -url https://example.com
```

### Output

On completion, a report is printed to the terminal:

```
┌────────────────────────────────────────────────────┐
│  sitemap-gen — crawl report                        │
├────────────────────────────────────────────────────┤
│  target   https://example.com                      │
│  output   sitemap.xml                              │
│  elapsed  4.201s  (10 workers)                     │
├────────────────────────────────────────────────────┤
│  pages                                             │
├────────────────────────────────────────────────────┤
│  ✓ crawled    142                                  │
│  ✗ errors     3                                    │
│  # unique     145                                  │
│  ↗ sitemap    142                                  │
├────────────────────────────────────────────────────┤
│  ████████████████████████████░░░░░░░░░░░░░░░░░░  │
│  success rate  97.9%          33.8 urls/sec        │
└────────────────────────────────────────────────────┘
```

And a `sitemap.xml` is written to disk:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://example.com/</loc>
    <lastmod>2024-01-15</lastmod>
    <changefreq>weekly</changefreq>
    <priority>0.5</priority>
  </url>
  ...
</urlset>
```

## Architecture

```
main
 ├── signal.NotifyContext      — cancels ctx on shutdown / SIGTERM
 ├── TaskQueue (buffered ch)   — feeds work to workers
 ├── dedup.Set (32 shards)     — prevents revisiting URLs
 ├── sitemap.Map               — collects crawled URLs thread-safely
 ├── DomainLimiter             — per-domain rate limiter with burst
 └── worker pool (N goroutines)
      └── workerLoop
           ├── idle detection       — all idle + queue empty → close queue → exit
           ├── DomainLimiter.Wait   — rate limit before fetching
           ├── fetcher.Fetch        — HTTP GET with context
           ├── parser.ExtractLinks  — normalize + scope-filter hrefs
           ├── sitemap.Map.Add      — record the page
           └── TaskQueue.AddTask    — enqueue children (if depth < max)
```

## Tests

```bash
# run all tests
go test ./...

# with race detector
go test -race ./...
```

## Profiling

A pprof endpoint runs on `localhost:6060` while the crawler is active:

```bash
go tool pprof http://localhost:6060/debug/pprof/goroutine
go tool pprof http://localhost:6060/debug/pprof/heap
```