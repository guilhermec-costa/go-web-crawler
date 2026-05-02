# go-web-crawler

A concurrent web crawler built in Go, developed to study and practice goroutines, channels, worker pools, mutexes, context cancellation, rate limiting, and retry strategies.

Built while following the [Go by Example](https://gobyexample.com) guide.

---

## Features

- **Concurrent crawling** via a worker pool with configurable size
- **Depth-limited traversal** — controls how many levels deep the crawler follows links
- **Domain scoping** — only follows links within the same domain as the seed URL
- **Token bucket rate limiter** — prevents overwhelming the target server
- **Retry with exponential backoff and jitter** — handles transient failures gracefully
- **Per-request timeout** — workers never hang indefinitely on slow responses
- **Global timeout** — the entire crawl session is time-bounded via context
- **JSON Lines output** — results are streamed to disk as they arrive, avoiding memory accumulation
- **Live progress ticker** — periodic updates on nodes processed, errors, active workers, and rate limiter state

---

## Architecture

```
cmd/main.go                  # Entrypoint — parses flags and bootstraps the crawler
internal/
  cli/                       # CLI flag parsing and validation
  crawler/
    engine.go                # Orchestrates the crawl: worker pool, job queue, result consumer
    extractor.go             # HTTP requests, HTML parsing, DOM node extraction
    rate_limiter.go          # Token bucket rate limiter
    retry.go                 # WithRetry — exponential backoff with jitter
    url.go                   # URL parsing and validation
  perf/
    benchmark.go             # Time tracking utility
```

### Crawl Flow

```
seed URL
   ↓
enqueueJob (wg.Add)
   ↓
extractionJobsQueue (buffered channel)
   ↓
N workers (goroutines) → HTTP GET → HTML parse → extract nodes
   ↓
extractionResultsQueue (buffered channel)
   ↓
result consumer → write to .jsonl → enqueue new links (wg.Add) → wg.Done
   ↓
wg.Wait → done
```

The `WaitGroup` tracks jobs from the moment they are enqueued until their result is fully consumed — including any new jobs they generate. This ensures the crawl only terminates when all work, including dynamically discovered URLs, is complete.

---

## Rate Limiting

A token bucket is pre-filled with `N` tokens on startup. A background goroutine refills the bucket every second. Each worker consumes one token before making an HTTP request — if the bucket is empty, the worker blocks until a token is available.

---

## Retry Strategy

Failed requests are retried using exponential backoff with random jitter:

```
delay = baseDelay × multiplier^attempt + rand(jitterMaxMs)
```

Only retryable status codes trigger a retry: `429`, `500`, `504`. Non-retryable responses (e.g. `403`, `404`) fail immediately.

---

## Output

Results are written as [JSON Lines](https://jsonlines.org) (`.jsonl`) — one JSON object per line, streamed to disk as each URL is processed:

```json
{"url": "https://example.com", "parentUrl": "", "nodeCount": {"a": 42, "div": 18, "h1": 1, "h2": 3, "p": 7}}
{"url": "https://example.com/about", "parentUrl": "https://example.com", "nodeCount": {"a": 10, "div": 5, "h1": 1, "h2": 0, "p": 4}}
```

---

## Usage

```bash
go run cmd/main.go [options]
```

### Options

| Flag | Default | Description |
|---|---|---|
| `-url` | *(required)* | Seed URL to start crawling from |
| `-depth` | `1` | How many levels deep to follow links |
| `-workers` | `3` | Number of concurrent worker goroutines |
| `-timeout` | `30000` | Global crawl timeout in milliseconds |
| `-tickupdate` | `2000` | Interval between progress ticker updates in milliseconds |
| `-o` | `extractions-<timestamp>.jsonl` | Output file path |
| `-v` | `false` | Verbose mode |

### Examples

```bash
# Basic crawl with defaults
go run cmd/main.go -url="https://books.toscrape.com"

# Deeper crawl with more workers
go run cmd/main.go -url="https://books.toscrape.com" -depth=2 -workers=10

# Fast ticker updates, verbose output
go run cmd/main.go -url="https://books.toscrape.com" -depth=1 -workers=5 -tickupdate=500 -v

# Custom output file and longer timeout
go run cmd/main.go -url="https://books.toscrape.com" -depth=3 -workers=20 -timeout=60000 -o=results.jsonl
```

---

## Sample Output

```
time=2026-05-01T23:29:20.826-03:00 level=INFO msg="crawler config" url=http://crawler-test.com depth=1 workers=30
time=2026-05-01T23:29:21.426-03:00 level=INFO msg="ticker update" elapsed=1.000s processed_nodes=1532 errors=0 active_workers=30 rate_limite_buf_len=57
time=2026-05-01T23:29:25.374-03:00 level=INFO msg="request failed" status_code=429 message="rate limited by server" attempt=3
time=2026-05-01T23:29:28.444-03:00 level=INFO msg="ticker update" elapsed=9.018s processed_nodes=8313 errors=75 active_workers=30 rate_limite_buf_len=100
time=2026-05-01T23:29:28.444-03:00 level=INFO msg=Bootstrap took=9.018713165s
```

---

## What I Learned

This project was built as a hands-on exercise after studying the Go by Example guide. Key concepts practiced:

- Goroutines and channel-based communication
- Worker pool pattern with dynamic job generation
- WaitGroup for tracking variable-length workloads
- Context propagation for timeout and cancellation
- Token bucket rate limiting
- Exponential backoff with jitter
- Mutex usage for shared state (`visited`, `enqueued` maps)
- Atomic counters for lock-free metrics
- JSON Lines streaming to avoid memory accumulation
- Race condition detection with `-race`