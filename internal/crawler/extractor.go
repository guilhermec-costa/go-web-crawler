package crawler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

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

type NodeData struct {
	Nodes []*html.Node
	Count int
}

type NodesByTag map[string]NodeData

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

func checkForCtxErr(err error, pageUrl *url.URL) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		slog.Error("request timeout", "url", pageUrl, "error", err)
		return fmt.Errorf("timeout requesting url %s: %w", pageUrl.String(), err)

	case errors.Is(err, context.Canceled):
		slog.Error("request canceled", "url", pageUrl, "error", err)
		return fmt.Errorf("request canceled for url %s: %w", pageUrl.String(), err)

	default:
		if urlErr, ok := errors.AsType[*url.Error](err); ok {
			slog.Error("url error", "op", urlErr.Op, "url", urlErr.URL, "err", urlErr.Err)
		} else {
			slog.Error("request failed", "url", pageUrl, "err", err)
		}
		return fmt.Errorf("request %q: %w", pageUrl, err)
	}
}

func is2xx(resp *http.Response) bool {
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

var retryConfig = RetryConfig{
	maxRetries:    3,
	baseDelay:     time.Duration(700 * time.Millisecond),
	jitterMaxRand: 100,
	multiplier:    2,
}

func extractPageNodes(ctx context.Context, pageUrl *url.URL, rateLimiter *RateLimiter) (NodesByTag, error) {
	resp, err := WithRetry(retryConfig, func() (*http.Response, error) {
		reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		req, reqErr := http.NewRequestWithContext(reqCtx, "GET", pageUrl.String(), nil)
		if reqErr != nil {
			slog.Error("failed to create request", "url", pageUrl, "error", reqErr)
			return nil, fmt.Errorf("failed to create request for url %s: %w", pageUrl.String(), reqErr)
		}

		rateLimiter.Wait()
		resp, respErr := http.DefaultClient.Do(req)

		if ctxErr := checkForCtxErr(respErr, pageUrl); ctxErr != nil {
			return nil, ctxErr
		}

		// cancel() clears the context, but resp.Body still needs to be read outside the closure
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}

		resp.Body = io.NopCloser(bytes.NewReader(body))
		return resp, respErr
	})

	if err != nil {
		slog.Error("failed to extract page nodes", "err", err)
		return nil, err
	}

	defer resp.Body.Close()

	doc, reqErr := html.Parse(resp.Body)
	if reqErr != nil {
		slog.Error("failed to parse html body", "url", pageUrl, "err", reqErr)
		return nil, fmt.Errorf("failed to parse html page for url %v: %w", pageUrl.String(), reqErr)
	}

	extractors := []DOMExtractor{
		NewDOMNodesExtractor(NodeExtractorTypeA),
		NewDOMNodesExtractor(NodeExtractorTypeH1),
		NewDOMNodesExtractor(NodeExtractorTypeDiv),
		NewDOMNodesExtractor(NodeExtractorTypeH2),
		NewDOMNodesExtractor(NodeExtractorTypeP),
	}

	nodesByExtractionType := make(NodesByTag)
	for _, e := range extractors {
		nodes := e.ExtractNodes(doc)
		nodesByExtractionType[e.Type] = NodeData{Nodes: nodes, Count: len(nodes)}
	}

	return nodesByExtractionType, nil
}

type WorkerDeps struct {
	verbose       bool
	activeWorkers *atomic.Int32
	rateLimiter   *RateLimiter
}

type ExtractionJob struct {
	Url       *url.URL
	ParentUrl *url.URL
	depth     int
}

func nodesExtractionWorker(ctx context.Context, extractionJobQueue <-chan ExtractionJob, extractionResultQueue chan<- UrlExtractionResult, deps WorkerDeps) {
	deps.activeWorkers.Add(1)
	defer deps.activeWorkers.Add(-1)

	for job := range extractionJobQueue {
		select {
		case <-ctx.Done():
			return
		default:
			if deps.verbose {
				slog.Info("extracting nodes from url", "url", job.Url.String())
			}

			nodes, err := extractPageNodes(ctx, job.Url, deps.rateLimiter)

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
