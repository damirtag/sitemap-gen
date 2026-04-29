# sitemap-gen

![Tests](https://github.com/damirtag/sitemap-gen/actions/workflows/ci.yml/badge.svg)

A concurrent sitemap generator written in Go. Feed it a URL, it crawls the entire domain and outputs a valid `sitemap.xml` — respecting depth limits, deduplicating URLs, and shutting down gracefully on Ctrl+C

Built to explore goroutines, worker pools, context propagation, and graceful shutdown patterns

## Features

- **Concurrent crawling** — configurable worker pool, each worker runs in its own goroutine
- **Domain scoping** — stays within the seed domain, never drifts to external sites
- **Depth limiting** — won't crawl infinitely deep
- **URL normalization** — resolves relative links, strips fragments, deduplicates
- **Graceful shutdown** — `SIGINT`/`SIGTERM` drains the queue and waits for all workers before exiting
- **Valid sitemap output** — writes a spec-compliant `sitemap.xml` with `<loc>`, `<lastmod>`, `<changefreq>`, `<priority>`
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

## Tests

```bash
go test ./...
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
 ├── signal.NotifyContext     — cancels ctx on Ctrl+C / SIGTERM
 ├── TaskQueue (buffered ch)  — feeds work to workers
 ├── dedup.Set                — prevents revisiting URLs
 ├── sitemap.Map              — collects crawled URLs thread-safely
 └── worker pool (N goroutines)
      └── workerLoop
           ├── fetcher.Fetch  — HTTP GET with context
           ├── parser.ExtractLinks  — normalize + scope-filter hrefs
           ├── sitemap.Map.Add      — record the page
           └── TaskQueue.AddTask   — enqueue children (if depth < max)
```

## Profiling

A pprof endpoint runs on `localhost:6060` while the crawler is active:

```bash
go tool pprof http://localhost:6060/debug/pprof/goroutine
go tool pprof http://localhost:6060/debug/pprof/heap
```