package crawler

import (
	"fmt"
	"log"
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
