package crawler

import (
	"fmt"
	"guilhermec-costa/go-web-crawler/internal/cli"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type NodesByTag map[string][]*html.Node

type extractionJob struct {
	Url       *url.URL
	ParentUrl *url.URL
	depth     int
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func StartCrawlerEngine(args cli.CrawlerFlags) {
	defer timeTrack(time.Now(), "StartCrawlerEngine")
	log.Printf("Starting crawler for: %s", args.RootUrl)

	rootUrl, err := url.Parse(args.RootUrl)
	if err != nil {
		log.Fatalf("[ERROR] url %s is not valid: %v", args.RootUrl, err)
	}

	if err := ValidateUrl(rootUrl); err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	extractions := []UrlExtractionResult{}
	visited := make(map[string]bool)

	var extractionMtx sync.Mutex
	var visitedMtx sync.Mutex

	extractionJobs := make(chan extractionJob, args.Workers)
	extractionResults := make(chan UrlExtractionResult, args.Workers)

	for range args.Workers {
		go nodesExtractionWorker(extractionJobs, extractionResults)
	}

	var wg sync.WaitGroup

	enqueueJob := func(curUrl *url.URL, parentUrl *url.URL, depth int) {
		wg.Add(1)
		go func() {
			extractionJobs <- extractionJob{curUrl, parentUrl, depth}
		}()
	}

	go func() {
		for result := range extractionResults {
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
					enqueueJob(resolvedUrl, result.url, result.depth+1)
				}
				visitedMtx.Unlock()
			}
			wg.Done()
		}
	}()

	enqueueJob(rootUrl, nil, 0)

	wg.Wait()
	close(extractionJobs)
	close(extractionResults)
	fmt.Println("Results: ", extractions)
}

func extractPageNodes(pageUrl *url.URL) (NodesByTag, error) {
	resp, err := http.Get(pageUrl.String())
	if err != nil {
		return nil, fmt.Errorf("Failed to request html page for url %v: %w", pageUrl.String(), err)
	}
	defer resp.Body.Close()

	if handler, ok := statusHandlers[resp.StatusCode]; ok {
		if err := handler(pageUrl); err != nil {
			return nil, fmt.Errorf("Failed to parse html page for url %v: %w", pageUrl.String(), err)
		}
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatal("[ERROR] Failed to parse html body. Exiting program")
		return nil, fmt.Errorf("Failed to parse html page for url %v: %w", pageUrl.String(), err)
	}

	extractors := []DOMExtractor{
		NewDOMNodesExtractor(NodeExtractorTypeA),
		NewDOMNodesExtractor(NodeExtractorTypeH1),
		NewDOMNodesExtractor(NodeExtractorTypeDiv),
		NewDOMNodesExtractor(NodeExtractorTypeH2),
		NewDOMNodesExtractor(NodeExtractorTypeP),
	}

	type result struct {
		key   string
		nodes []*html.Node
	}

	extractions := make(chan result, len(extractors))

	for _, e := range extractors {
		// passing "e" as argument prevents race condition
		go func(e DOMExtractor) {
			log.Printf("Extracting dom nodes for <%v> element", e.Type)
			extractions <- result{
				key:   e.Type,
				nodes: e.ExtractNodes(doc),
			}
		}(e)
	}

	nodesByExtractionType := make(NodesByTag)
	for range extractors {
		r := <-extractions
		nodesByExtractionType[r.key] = r.nodes
	}

	return nodesByExtractionType, nil
}

func nodesExtractionWorker(extractionQueue <-chan extractionJob, extractionResult chan<- UrlExtractionResult) {
	for job := range extractionQueue {
		log.Printf("Extracting nodes from %v", job.Url.String())

		nodes, err := extractPageNodes(job.Url)

		result := UrlExtractionResult{
			extractedNodes: nodes,
			url:            job.Url,
			parentUrl:      job.ParentUrl,
			error:          err,
			depth:          job.depth,
		}

		extractionResult <- result
	}
}
