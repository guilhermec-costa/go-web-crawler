package crawler

import (
	"context"
	"errors"
	"fmt"
	"guilhermec-costa/go-web-crawler/internal/cli"
	"guilhermec-costa/go-web-crawler/internal/perf"
	"log"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type NodesByTag map[string][]*html.Node

type ExtractionJob struct {
	Url       *url.URL
	ParentUrl *url.URL
	depth     int
}

func runCrawler(ctx context.Context, args cli.CrawlerFlags) error {
	log.Printf("Starting crawler for: %s", args.RootUrl)

	rootUrl, err := url.Parse(args.RootUrl)
	if err != nil {
		return fmt.Errorf("[ERROR] failed to parse url %s: %w", args.RootUrl, err)
	}

	if err := ValidateUrl(rootUrl); err != nil {
		return fmt.Errorf("[ERROR] url %s is not valid : %w", args.RootUrl, err)
	}

	extractions := []UrlExtractionResult{}
	visited := make(map[string]bool)

	var extractionMtx sync.Mutex
	var visitedMtx sync.Mutex

	extractionJobsQueue := make(chan ExtractionJob, args.Workers)
	extractionResultsQueue := make(chan UrlExtractionResult, args.Workers)

	for range args.Workers {
		go nodesExtractionWorker(ctx, extractionJobsQueue, extractionResultsQueue)
	}

	var wg sync.WaitGroup

	enqueueExtractionJob := func(curUrl *url.URL, parentUrl *url.URL, depth int) {
		wg.Add(1)
		go func() {
			extractionJobsQueue <- ExtractionJob{curUrl, parentUrl, depth}
		}()
	}

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

			extractionMtx.Lock()
			extractions = append(extractions, result)
			extractionMtx.Unlock()

			for _, node := range result.extractedNodes[NodeExtractorTypeA] {
				link := GetAttrValueFromNode(node, "href")
				if link == nil || *link == "" || *link == "#" {
					continue
				}

				parsedUrl, err := url.Parse(*link)
				if err != nil {
					log.Printf("[ERROR] Failed to parse link from node %v", node)
					continue
				}

				resolvedUrl := result.url.ResolveReference(parsedUrl)

				visitedMtx.Lock()
				if !visited[resolvedUrl.String()] && result.depth < args.Depth {
					visited[resolvedUrl.String()] = true
					enqueueExtractionJob(resolvedUrl, result.url, result.depth+1)
				}
				visitedMtx.Unlock()
			}
			wg.Done()
		}
	}()

	enqueueExtractionJob(rootUrl, nil, 0)

	wg.Wait()
	close(extractionJobsQueue)
	close(extractionResultsQueue)

	fmt.Println("Results: ", extractions)
	return nil
}

func Bootstrap(args cli.CrawlerFlags) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	defer perf.TimeTrack(time.Now(), "Bootstrap")

	if err := runCrawler(ctx, args); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Print("Timer exceeded timeout")
			return
		}
		log.Printf("[ERROR] %v", err)
	}
}

func nodesExtractionWorker(ctx context.Context, extractionJobQueue <-chan ExtractionJob, extractionResultQueue chan<- UrlExtractionResult) {
	for job := range extractionJobQueue {
		select {
		case <-ctx.Done():
			return
		default:
			log.Printf("Extracting nodes from %v", job.Url.String())

			nodes, err := extractPageNodes(ctx, job.Url)

			result := UrlExtractionResult{
				extractedNodes: nodes,
				url:            job.Url,
				parentUrl:      job.ParentUrl,
				error:          err,
				depth:          job.depth,
			}

			extractionResultQueue <- result
		}
	}
}
