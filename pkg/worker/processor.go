package worker

import (
	"bytes"

	"golang.org/x/net/html"
)

type processor struct{}

func NewProcessor() *processor {
	return &processor{}
}

// TODO: process html based on transformers
func (p *processor) processHTML(body []byte) error {
	var walk func(*html.Node)

	walk = func(node *html.Node) {
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	node, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return err
	}

	walk(node)

	return nil
}
