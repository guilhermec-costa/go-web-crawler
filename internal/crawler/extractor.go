package crawler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"sync"

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

type UrlExtractionResult struct {
	ExtractedNodes NodesByTag
	url            *url.URL
	parentUrl      *url.URL
	error          error
	Depth          int
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
	for _, tagMap := range r.ExtractedNodes {
		total += len(tagMap.Nodes)
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
		slog.Error("failed to create request", "url", pageUrl, "error", err)
		return nil, fmt.Errorf("failed to create request for url %s: %w", pageUrl.String(), err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			slog.Error("request timeout", "url", pageUrl, "error", err)
			return nil, fmt.Errorf("timeout requesting url %s: %w", pageUrl.String(), err)

		case errors.Is(err, context.Canceled):
			slog.Error("request canceled", "url", pageUrl, "error", err)
			return nil, fmt.Errorf("request canceled for url %s: %w", pageUrl.String(), err)

		default:
			if urlErr, ok := errors.AsType[*url.Error](err); ok {
				slog.Error("url error", "op", urlErr.Op, "url", urlErr.URL, "err", urlErr.Err)
			} else {
				slog.Error("request failed", "url", pageUrl, "err", err)
			}
			return nil, fmt.Errorf("request %q: %w", pageUrl, err)
		}
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
	var wg sync.WaitGroup

	for _, e := range extractors {
		select {
		case <-ctx.Done():
			continue

		default:
			wg.Add(1)
			// passing "e" as argument prevents race condition
			go func(e DOMExtractor) {
				defer wg.Done()
				if v, ok := ctx.Value(verboseKey).(bool); ok && v {
					slog.Info("extracing dom nodes", "element_type", e.Type)
				}
				extractions <- result{
					key:   e.Type,
					nodes: e.ExtractNodes(doc),
				}
			}(e)
		}
	}

	go func() {
		wg.Wait()
		close(extractions)
	}()

	nodesByExtractionType := make(NodesByTag)
	for r := range extractions {
		nodesByExtractionType[r.key] =
			struct {
				Nodes []*html.Node
				Count int
			}{Nodes: r.nodes, Count: len(r.nodes)}
	}

	return nodesByExtractionType, nil
}

func nodesExtractionWorker(ctx context.Context, extractionJobQueue <-chan ExtractionJob, extractionResultQueue chan<- UrlExtractionResult) {
	for job := range extractionJobQueue {
		select {
		case <-ctx.Done():
			return
		default:
			if v, ok := ctx.Value(verboseKey).(bool); ok && v {
				slog.Info("extracting nodes from url", "url", job.Url.String())
			}

			nodes, err := extractPageNodes(ctx, job.Url)

			result := UrlExtractionResult{
				ExtractedNodes: nodes,
				url:            job.Url,
				parentUrl:      job.ParentUrl,
				error:          err,
				Depth:          job.depth,
			}

			extractionResultQueue <- result
		}
	}
}
