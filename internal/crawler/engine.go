package crawler

import (
	"context"
	"encoding/json"
	"errors"
	"guilhermec-costa/go-web-crawler/internal/cli"
	"guilhermec-costa/go-web-crawler/internal/perf"
	"log"
	"log/slog"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/html"
)

type NodesByTag map[string]struct {
	Nodes []*html.Node
	Count int
}

type ExtractionJob struct {
	Url       *url.URL
	ParentUrl *url.URL
	depth     int
}

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

const verboseKey = "verbose"

type TickerMetadata struct {
	processedNodes atomic.Int64
}

func runCrawler(ctx context.Context, args cli.CrawlerFlags) error {
	start := time.Now()
	slog.Info("Starting crawler", "url", args.RootUrl)

	procUpdtTicker := time.NewTicker(2 * time.Second)
	tickerDone := make(chan struct{})
	tickerMetadata := TickerMetadata{}

	go func() {
		for {
			select {
			case <-tickerDone:
				return
			case <-procUpdtTicker.C:
				slog.Info("ticker update", "elapsed", time.Since(start),
					"processed_nodes", tickerMetadata.processedNodes.Load())
			}
		}
	}()

	rootUrl, err := ParseAndValidateURL(args.RootUrl)
	if err != nil {
		return err
	}

	extractions := []UrlExtractionResult{}
	visited := make(map[string]bool)
	enqueued := make(map[string]bool)

	var extractionMtx sync.Mutex
	var visitedMtx sync.Mutex

	extractionJobsQueue := make(chan ExtractionJob, args.Workers)
	extractionResultsQueue := make(chan UrlExtractionResult, args.Workers)

	ctxWithVerboseFlag := context.WithValue(ctx, verboseKey, args.Verbose)
	for range args.Workers {
		go nodesExtractionWorker(ctxWithVerboseFlag, extractionJobsQueue, extractionResultsQueue)
	}

	var wg sync.WaitGroup

	enqueueExtractionJob := func(curUrl *url.URL, parentUrl *url.URL, depth int) {
		wg.Add(1)
		go func() {
			extractionJobsQueue <- ExtractionJob{curUrl, parentUrl, depth}
		}()
	}

	file, _ := os.OpenFile(args.OutputPath, os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	go func() {
		for result := range extractionResultsQueue {
			encoder.Encode(result.ToJSON())

			visitedMtx.Lock()
			if visited[result.url.String()] {
				visitedMtx.Unlock()
				wg.Done()
				continue
			}
			visited[result.url.String()] = true
			visitedMtx.Unlock()

			extractionMtx.Lock()
			extractions = append(extractions, result)
			extractionMtx.Unlock()

			for _, data := range result.ExtractedNodes {
				tickerMetadata.processedNodes.Add(int64(data.Count))
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
		tickerDone <- struct{}{}
	case <-ctx.Done():
		return ctx.Err()
	}

	close(extractionJobsQueue)
	close(extractionResultsQueue)

	slog.Info("final result", "node_count", tickerMetadata.processedNodes.Load())
	return nil

}

func setupOutputDir() {
	if err := os.MkdirAll("../output", 0o755); err != nil {
		slog.Error("failed to create output directory",
			"path", "../output",
			"err", err,
		)
		panic(err)
	}

	slog.Info("output directory ready",
		"path", "../output",
	)
}

func Bootstrap(args cli.CrawlerFlags) {
	setupOutputDir()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	defer perf.TimeTrack(time.Now(), "Bootstrap")

	if err := runCrawler(ctx, args); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Print("Timer exceeded timeout")
			return
		}
		slog.Error("could not complete crawling", "err", err)
	}
}
