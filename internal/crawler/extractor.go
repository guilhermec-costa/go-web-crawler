package crawler

import (
	"golang.org/x/net/html"
	"log"
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
	NodeExtractorTypeHref string = "a"
	NodeExtractorTypeH1   string = "h1"
	NodeExtractorTypeH2   string = "h2"
	NodeExtractorTypeDiv  string = "div"
	NodeExtractorTypeP    string = "p"
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

func ReportExtractionsByTag(tags map[string][]*html.Node) {
	log.Println("[REPORT] Extraction summary")
	log.Println("--------------------------------")

	for tag, nodes := range tags {
		log.Printf("• %-5s → %3d nodes", tag, len(nodes))
	}
}
