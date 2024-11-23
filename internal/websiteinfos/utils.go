package websiteinfos

import (
	"bytes"

	"golang.org/x/net/html"
)

func nodeToString(node *html.Node) string {
	var b bytes.Buffer
	err := html.Render(&b, node)
	if err != nil {
		return "Rendering error: " + err.Error()
	}
	return b.String()
}

func getAttr(n *html.Node, key string) *html.Attribute {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return &attr
		}
	}
	return nil
}
