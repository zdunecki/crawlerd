package layer

import (
	"fmt"
	"io"

	"golang.org/x/net/html"
)

const (
	HTMLTextarea = "TEXTAREA"
	HTMLInput    = "INPUT"
	HTMLSelect   = "SELECT"
	HTMLScript   = "SCRIPT"
	HTMLNoScript = "NOSCRIPT"
)

const (
	HTMLAttributeValue = "value"
)

type HTML struct {
	NodeType    html.NodeType     `json:"nT"`
	ID          uint              `json:"i"`
	Name        string            `json:"n"`
	TagName     string            `json:"tN"`
	Attrs       map[string]string `json:"a"`
	TextContent string            `json:"tC"`
	ChildNodes  []*HTML           `json:"cN"`

	reader io.Reader
	idC    uint
}

func NewHTMLReader(reader io.Reader) *HTML {
	return &HTML{
		reader: reader,
	}
}

func (h *HTML) Encode() (*HTML, error) {
	htmlA, err := html.Parse(h.reader)
	if err != nil {
		return nil, err
	}

	var scan func(*html.Node) *HTML
	scan = func(node *html.Node) *HTML {
		data := &HTML{
			NodeType: htmlA.Type,
			ID:       h.incrID(),
		}

		switch node.Type {
		case html.DocumentNode:
			fmt.Println("TODO")
		case html.CommentNode, html.TextNode:
			if node.Parent != nil {
				if node.Parent.Data == HTMLTextarea {
					data.TextContent = node.Data
					break
				}
			}

			data.TextContent = node.Data
		case html.ElementNode:
			data.TagName = node.Data
			data.Attrs = make(map[string]string)

			for _, attr := range node.Attr {
				data.Attrs[attr.Key] = attr.Val
			}

			switch node.Data {
			case HTMLInput, HTMLSelect, HTMLTextarea:
				value := ""

				for _, attr := range node.Attr {
					if attr.Key == HTMLAttributeValue {
						value = attr.Val
					}
				}

				data.Attrs["value"] = value
			}

			if node.Data == HTMLScript || node.Data == HTMLNoScript {
				break
			}
		}

		if node.FirstChild != nil {
			data.ChildNodes = make([]*HTML, 0)
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			data.ChildNodes = append(data.ChildNodes, scan(c))
		}

		return data
	}

	return scan(htmlA), nil
}

func (h *HTML) incrID() uint {
	h.idC += 1
	return h.idC
}
