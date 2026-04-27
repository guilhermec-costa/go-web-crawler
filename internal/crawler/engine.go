package crawler

import (
	"golang.org/x/net/html"
	"guilhermec-costa/go-web-crawler/internal/cli"
	"log"
	"net/http"
	"net/url"
)

func StartCrawlerEngine(args cli.CrawlerFlags) {
	log.Printf("Starting crawler for: %s", args.Url)

	_url, err := url.Parse(args.Url)
	if err != nil {
		log.Fatalf("[ERROR] url %s is not valid: %v", args.Url, err)
	}

	if err := ValidateUrl(_url); err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	resp, err := http.Get(_url.String())
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if handler, ok := statusHandlers[resp.StatusCode]; ok {
		if err := handler(_url); err != nil {
			log.Fatal(err)
		}
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatal("[ERROR] Failed to parse html body. Exiting program")
	}

	extractors := []DOMExtractor{
		NewDOMNodesExtractor(NodeExtractorTypeHref),
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
		go func(e DOMExtractor) {
			log.Printf("Extracting dom nodes for %v type element", e.Type)
			extractions <- result{
				key:   e.Type,
				nodes: e.ExtractNodes(doc),
			}
		}(e)
	}

	nodesByExtractionType := map[string][]*html.Node{}
	for range extractors {
		r := <-extractions
		nodesByExtractionType[r.key] = r.nodes
	}

	ReportExtractionsByTag(nodesByExtractionType)
}
