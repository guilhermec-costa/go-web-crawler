package crawler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"guilhermec-costa/go-web-crawler/crawler/perf"
	val "guilhermec-costa/go-web-crawler/crawler/validation"
	"log"
	"log/slog"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type UrlExtractionResultJSON struct {
	Url       string         `json:"url"`
	ParentUrl string         `json:"parentUrl"`
	NodeCount map[string]int `json:"nodeCount"`
}

func (r UrlExtractionResult) ToJSON() UrlExtractionResultJSON {
	counts := map[string]int{}
	for tag, nodes := range r.ExtractedNodes {
		counts[tag] = len(nodes.Nodes)
	}

	parentUrl := ""
	if r.parentUrl != nil {
		parentUrl = r.parentUrl.String()
	}
	return UrlExtractionResultJSON{Url: r.url.String(), ParentUrl: parentUrl, NodeCount: counts}
}

type TickerMetadata struct {
	start           time.Time
	processedNodes  atomic.Int64
	errors          atomic.Int64
	activeWorkers   atomic.Int32
	rateLimitBufLen int
}

func showTickerData(tickerMetadata *TickerMetadata) {
	slog.Info("ticker update", "elapsed", time.Since(tickerMetadata.start),
		"processed_nodes", tickerMetadata.processedNodes.Load(),
		"errors", tickerMetadata.errors.Load(),
		"active_workers", tickerMetadata.activeWorkers.Load(),
		"rate_limite_buf_len", tickerMetadata.rateLimitBufLen)
}

const RateLimiterBurstRate = 100

func runCrawler(ctx context.Context, args val.CrawlerParams) error {
	start := time.Now()
	slog.Info("Starting crawler", "url", args.RootUrl)

	rateLimiter := NewRateLimiter(ctx, RateLimiterBurstRate, 1000*time.Millisecond)

	tickerMetadata := TickerMetadata{start: start}
	ticker := NewTickerProgress(ctx, args.TickUpdateMs, func() {
		tickerMetadata.rateLimitBufLen = rateLimiter.TokenCount()
		showTickerData(&tickerMetadata)
	})

	ticker.startTicking()

	rootUrl, err := ParseAndValidateURL(args.RootUrl)
	if err != nil {
		return err
	}

	visited := make(map[string]bool)
	enqueued := make(map[string]bool)

	var visitedMtx sync.Mutex

	extractionJobsQueue := make(chan ExtractionJob, args.Workers*10)
	extractionResultsQueue := make(chan UrlExtractionResult, args.Workers*10)

	for range args.Workers {
		go nodesExtractionWorker(ctx, extractionJobsQueue, extractionResultsQueue, WorkerDeps{
			verbose:       args.Verbose,
			activeWorkers: &tickerMetadata.activeWorkers,
			rateLimiter:   rateLimiter,
		})
	}

	var wg sync.WaitGroup

	enqueueExtractionJob := func(curUrl *url.URL, parentUrl *url.URL, depth int) {
		wg.Add(1)
		go func() {
			// allow quick enqueue, so the main goroutine don't get stucked
			extractionJobsQueue <- ExtractionJob{curUrl, parentUrl, depth}
		}()
	}

	file, err := os.OpenFile(args.OutputPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	go func() {
		for result := range extractionResultsQueue {
			visitedMtx.Lock()
			if visited[result.url.String()] {
				visitedMtx.Unlock()
				wg.Done()
				continue
			}
			visited[result.url.String()] = true
			visitedMtx.Unlock()

			encoder.Encode(result.ToJSON())

			for _, data := range result.ExtractedNodes {
				tickerMetadata.processedNodes.Add(int64(data.Count))
			}

			if result.error != nil {
				tickerMetadata.errors.Add(1)
			}

			for _, node := range result.ExtractedNodes[NodeExtractorTypeA].Nodes {
				link := GetAttrValueFromNode(node, "href")
				if link == nil || *link == "" || *link == "#" {
					continue
				}

				parsedUrl, err := url.Parse(*link)
				if err != nil {
					slog.Error("failed to parse link from node", "node", node)
					continue
				}

				resolvedUrl := result.url.ResolveReference(parsedUrl)

				if resolvedUrl.Host == result.url.Host &&
					!enqueued[resolvedUrl.String()] &&
					result.Depth < args.Depth {
					enqueued[resolvedUrl.String()] = true
					enqueueExtractionJob(resolvedUrl, result.url, result.Depth+1)
				}
			}
			wg.Done()
		}
	}()

	enqueueExtractionJob(rootUrl, nil, 0)

	nodeExtractionDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(nodeExtractionDone)
	}()

	// checks which channel closes first
	select {
	case <-nodeExtractionDone:
		slog.Info("final ticker update after extraction completion")
		ticker.Display()
		ticker.Stop()
	case <-ctx.Done():
		return ctx.Err()
	}

	close(extractionJobsQueue)
	close(extractionResultsQueue)
	return nil

}

func Bootstrap(args val.CrawlerParams) error {
	ctx, cancel := context.WithTimeout(context.Background(), (time.Duration(args.TimeoutMs))*time.Millisecond)
	defer cancel()
	defer perf.TimeTrack(time.Now(), "Bootstrap")

	if err := runCrawler(ctx, args); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Print("Timer exceeded timeout")
			return err
		}
		slog.Error("could not complete crawling", "err", err)
	}

	return nil
}
