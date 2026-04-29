package crawler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"golang.org/x/net/html"
)

func extractNodesOfType(rootNode *html.Node, nodeType string) []*html.Node {
	nodes := []*html.Node{}

	for n := range rootNode.Descendants() {
		if n.Type == html.ElementNode && n.Data == nodeType {
			nodes = append(nodes, n)
		}
	}
	return nodes
}

const (
	NodeExtractorTypeA   string = "a"
	NodeExtractorTypeH1  string = "h1"
	NodeExtractorTypeH2  string = "h2"
	NodeExtractorTypeDiv string = "div"
	NodeExtractorTypeP   string = "p"
)

type DOMExtractor struct {
	Type         string
	ExtractNodes func(node *html.Node) []*html.Node
}

func NewDOMNodesExtractor(extType string) DOMExtractor {
	return DOMExtractor{
		ExtractNodes: func(n *html.Node) []*html.Node {
			return extractNodesOfType(n, extType)
		},
		Type: extType,
	}
}

func GetAttrValueFromNode(aElem *html.Node, attrKey string) *string {
	for _, attr := range aElem.Attr {
		if attr.Key == attrKey {
			return &attr.Val
		}
	}
	return nil
}

func ReportExtractionsByTag(tags map[string][]*html.Node) {
	log.Println("[REPORT] Extraction summary")
	log.Println("--------------------------------")

	for tag, nodes := range tags {
		log.Printf("• %-5s → %3d nodes", tag, len(nodes))
	}
}

type UrlExtractionResult struct {
	extractedNodes NodesByTag
	url            *url.URL
	parentUrl      *url.URL
	error          error
	depth          int
}

func (r UrlExtractionResult) String() string {
	urlStr := "<nil>"
	if r.url != nil {
		urlStr = r.url.String()
	}

	parentStr := "<nil>"
	if r.parentUrl != nil {
		parentStr = r.parentUrl.String()
	}

	errStr := "nil"
	if r.error != nil {
		errStr = r.error.Error()
	}

	total := 0
	for _, nodes := range r.extractedNodes {
		total += len(nodes)
	}

	return fmt.Sprintf(
		"URL=%s | Parent=%s | Nodes=%d | Error=%s\n",
		urlStr,
		parentStr,
		total,
		errStr,
	)
}

func extractPageNodes(ctx context.Context, pageUrl *url.URL) (NodesByTag, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", pageUrl.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for url %s: %w", pageUrl.String(), err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("timeout requesting url %s: %w", pageUrl.String(), err)
		}
		if errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("request canceled for url %s: %w", pageUrl.String(), err)
		}
		return nil, fmt.Errorf("failed to request url %s: %w", pageUrl.String(), err)
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
