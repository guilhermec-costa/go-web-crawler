package crawler

import (
	"fmt"
	"guilhermec-costa/go-web-crawler/internal/cli"
	"log"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
)

type NodeExtractionByTagResult map[string][]*html.Node

func StartCrawlerEngine(args cli.CrawlerFlags) {
	log.Printf("Starting crawler for: %s", args.RootUrl)

	rootUrl, err := url.Parse(args.RootUrl)
	if err != nil {
		log.Fatalf("[ERROR] url %s is not valid: %v", args.RootUrl, err)
	}

	if err := ValidateUrl(rootUrl); err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	var recursiveExtractPageNodes func(*url.URL, *url.URL, int)

	extractions := []UrlExtractionResult{}
	visited := make(map[string]bool)

	recursiveExtractPageNodes = func(url *url.URL, parentUrl *url.URL, depth int) {
		if depth > args.Depth {
			return
		}

		if visited[url.String()] {
			return
		}

		visited[url.String()] = true

		log.Printf("Extracting nodes from %v", url.String())

		var result UrlExtractionResult
		nodes, err := ExtractPageNodes(url)
		if err != nil {
			log.Printf("[ERROR] Failed to extract page nodes for %v: %s", url.String(), err.Error())
			result = UrlExtractionResult{extractions: nil, url: url, parentUrl: parentUrl, error: err}
			extractions = append(extractions, result)
			return
		}

		result = UrlExtractionResult{extractions: nodes, url: url, parentUrl: parentUrl, error: nil}
		extractions = append(extractions, result)

		for _, node := range nodes[NodeExtractorTypeA] {
			link := GetAttrValueFromNode(node, "href")
			if link == nil || *link == "" || *link == "#" {
				log.Printf("[ERROR] Failed to get link from node %v", node)
				continue
			}

			u, err := url.Parse(*link)
			if err != nil {
				log.Printf("[ERROR] Failed to parse link from node %v", node)
				continue
			}

			u = url.ResolveReference(u)
			recursiveExtractPageNodes(u, url, depth+1)
		}
	}
	recursiveExtractPageNodes(rootUrl, nil, 0)

	fmt.Println("Results: ", extractions)
}

func ExtractPageNodes(pageUrl *url.URL) (NodeExtractionByTagResult, error) {
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
		// NewDOMNodesExtractor(NodeExtractorTypeH1),
		// NewDOMNodesExtractor(NodeExtractorTypeDiv),
		// NewDOMNodesExtractor(NodeExtractorTypeH2),
		// NewDOMNodesExtractor(NodeExtractorTypeP),
	}

	type result struct {
		key   string
		nodes []*html.Node
	}

	extractions := make(chan result, len(extractors))

	for _, e := range extractors {
		go func(e DOMExtractor) {
			log.Printf("Extracting dom nodes for %v type element", e.Type)
			extractions <- result{
				key:   e.Type,
				nodes: e.ExtractNodes(doc),
			}
		}(e)
	}

	nodesByExtractionType := make(NodeExtractionByTagResult)
	for range extractors {
		r := <-extractions
		nodesByExtractionType[r.key] = r.nodes
	}

	return nodesByExtractionType, nil
}
