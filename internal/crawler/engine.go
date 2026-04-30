package crawler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"guilhermec-costa/go-web-crawler/internal/cli"
	"guilhermec-costa/go-web-crawler/internal/perf"
	"log"
	"net/url"
	"os"
	"sync"
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

func validateRawUrl(rawUrl string) (*url.URL, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] failed to parse url %s: %w", rawUrl, err)
	}

	if err := ValidateUrl(parsedUrl); err != nil {
		return nil, fmt.Errorf("[ERROR] url %s is not valid : %w", rawUrl, err)
	}

	return parsedUrl, nil
}

func runCrawler(ctx context.Context, args cli.CrawlerFlags) error {
	log.Printf("Starting crawler for: %s", args.RootUrl)

	rootUrl, err := validateRawUrl(args.RootUrl)
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

			for _, node := range result.ExtractedNodes[NodeExtractorTypeA].Nodes {
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

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// checks which channel closes first
	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	}

	close(extractionJobsQueue)
	close(extractionResultsQueue)

	return nil

}

func setupOutputDir() {
	createErr := os.Mkdir("../output", 0755)
	if createErr != nil {
		if errors.Is(createErr, os.ErrExist) {
			log.Printf("Directory already exist")
		} else {
			panic(createErr)
		}
	}
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
		log.Printf("[ERROR] %v", err)
	}
}
