package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
	"golang.org/x/net/html"
	"net/http"
	"os"
	"strings"
)

// TODO: faster serialization
// TODO: serialization should be done via browser engine itself? (without javascript)
var jsScriptSerializeDOM = `() => {
    function dynamicCssContent(el) {
        const cssRules = el && el.sheet && el.sheet.cssRules
        
        if (!cssRules) {
            return ""
        }

		let text = ""
		
		for (const rule of cssRules) {
			text += rule.cssText + " \n"
		}
        
        return text
    }
    
    function search(el, parent, i) {
    	if (!el) {
            return
    	}
        
        i++
        
        // TODO: computed styles
        const data = {
            i,
            nT: el.nodeType,
			tN: el.nodeName,
			a: {},
			tC: "",
			cN: [],
			r: el.getBoundingClientRect ? el.getBoundingClientRect() : undefined
        }
        
        if (el.attributes) {
			for (const a of el.attributes) {
				data.a[a.name] = a.value
			}	
        }

        switch (el.nodeType) {
            case Node.COMMENT_NODE:
            case Node.TEXT_NODE:
				data.tC = el.textContent
                
			   	if (!data.tC) {
			   		data.tC = dynamicCssContent(parent)
			   	}
               
               	break
        }
   
                
        if (!el.childNodes) {
            return
        }
        
        for (const e of el.childNodes) {
            const child = search(e, el, i)
            data.cN.push(child)
        }
        
        return data
	}
	
    return JSON.stringify(search(document, null, 0))
}
`

type Rect struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`

	Width  float64 `json:"width"`
	Height float64 `json:"height"`

	Top    float64 `json:"top"`
	Bottom float64 `json:"bottom"`

	Left  float64 `json:"left"`
	Right float64 `json:"right"`
}

type DOM struct {
	ID          int                    `json:"i"`
	Type        int                    `json:"nT"`
	TagName     string                 `json:"tN"`
	Attrs       map[string]interface{} `json:"a"`
	TextContent string                 `json:"tC"`
	ChildNodes  []*DOM                 `json:"cN"`
	Rect        *Rect                  `json:"r"`
}

const (
	JSNodeElement      = 1
	JSNodeAttribute    = 2
	JSNodeText         = 3
	JSNodeCdata        = 4
	JSNodeProcessing   = 7
	JSNodeComment      = 8
	JSNodeDocument     = 9
	JSNodeDocumentType = 10
	JSNodeFragment     = 11
)

func convertJSDOMToHTMLNode(jsDOM *DOM, node *html.Node) *html.Node {
	attributes := make([]html.Attribute, 0)

	for k, v := range jsDOM.Attrs {
		var val string

		switch vv := v.(type) {
		case string:
			val = vv
		case uint, uint8, uint16, uint32, uint64, int, int8, int16, int32, int64:
			fmt.Println("ok")
			// TODO:
			//strconv.Itoa(int(vv))
			//val = vv
		}

		if val != "" {
			attributes = append(attributes, html.Attribute{
				Key: k,
				Val: val,
			})
		}
	}

	// TODO: find better solution?
	if jsDOM.Rect != nil {
		rectB, _ := json.Marshal(jsDOM.Rect)

		attributes = append(attributes, html.Attribute{
			Key: "crawlerd-rect",
			Val: string(rectB),
		})
	}

	data := ""
	var nodeType html.NodeType

	switch jsDOM.Type {
	case JSNodeElement:
		nodeType = html.ElementNode
	case JSNodeAttribute:
		fmt.Println("ok")
	// TODO:
	case JSNodeText:
		nodeType = html.TextNode
	case JSNodeComment:
		nodeType = html.CommentNode
	case JSNodeDocument:
		nodeType = html.DocumentNode
	case JSNodeDocumentType:
		nodeType = html.DoctypeNode
	case JSNodeFragment:
		nodeType = html.DoctypeNode
	}

	if nodeType == html.TextNode {
		data = jsDOM.TextContent
	} else {
		data = strings.ToLower(jsDOM.TagName)
	}

	return &html.Node{
		Parent:      node,
		FirstChild:  nil,
		LastChild:   nil,
		PrevSibling: nil,
		NextSibling: nil,
		Type:        nodeType,
		DataAtom:    0,
		Data:        data,
		Namespace:   "",
		Attr:        attributes,
	}

}

func buildHTMLTree(el *DOM, node *html.Node) *html.Node {
	newNode := convertJSDOMToHTMLNode(el, node)

	for _, c := range el.ChildNodes {
		newNode.AppendChild(buildHTMLTree(c, node))
	}

	return newNode
}

func getRect(el *goquery.Selection) (rect *Rect) {
	val, exists := el.Attr("crawlerd-rect")
	if !exists {
		rect = nil
		return
	}

	json.Unmarshal([]byte(val), &rect)

	return
}

func isWithin(rect1 *Rect, rect2 *Rect) {

}

var ignoredTextNode = map[string]bool{
	"script":   true,
	"noscript": true,
	"style":    true,
}

func getTextNodes(root *html.Node) []*html.Node {
	textNodes := make([]*html.Node, 0)
	var inner func(el *html.Node)

	inner = func(el *html.Node) {
		if el == nil {
			return
		}

		if el.Type == html.TextNode {
			if _, ignore := ignoredTextNode[el.Parent.Data]; ignore {
				return
			}

			textNodes = append(textNodes, el)
			return
		}

		for c := el.FirstChild; c != nil; c = c.NextSibling {
			inner(c)
		}
	}

	inner(root)

	return textNodes
}

// find closest relation between links, box, image and text in DOM

// if link has image - +1
// if link has image and closest text is very close from bottom or right - +1
// if document.body.innerText is engineering - +1

// link can be:
// * parent - image or text are children
// * child - image can be sibling but something is in root
// * sibling ?

// TODO: check if elements are in grid horizontal/vertical (mobile)
func isBlog(doc *goquery.Document) {
	links := doc.Find("a")

	// TODO: mark node as checked
	links.Each(func(i int, link *goquery.Selection) {
		images := link.Find("img")

		// TODO: if link is children/sibling of image
		if images == nil || images.Nodes == nil {
			return
		}

		textNodeRects := make(map[*html.Node]*Rect, 0)

		for _, n := range link.Nodes {
			textNodes := getTextNodes(n)
			if len(textNodes) > 0 {
				for _, t := range textNodes {
					p := t.Parent
					attr := getAttribute("crawlerd-rect", p)
					if attr == nil {
						continue
					}

					var rect *Rect

					json.Unmarshal([]byte(attr.Val), &rect)

					if rect != nil {
						if rect.Bottom == 0 && rect.Right == 0 {
							continue
						}

						textNodeRects[t] = rect
					}
				}
			}
		}

		images.Each(func(i int, image *goquery.Selection) {
			imageRect := getRect(image)
			if imageRect == nil {
				// TODO: log err
				return
			}

			if len(textNodeRects) <= 0 {
				return
			}

			// TODO: all is not needed just greater than half should be enough or only the largest text?
			allTextDownFromImage := false

			for _, t := range textNodeRects {
				textIsBelowImage := t.Top >= imageRect.Bottom && imageRect.Left <= t.Left && imageRect.Right >= t.Right
				if textIsBelowImage {
					allTextDownFromImage = true
					continue
				}

				textIsRightFromImage := t.Left >= imageRect.Right
				if textIsRightFromImage {
					allTextDownFromImage = true
					continue
				}

				allTextDownFromImage = false
				break
			}

			fmt.Println(allTextDownFromImage, imageRect, textNodeRects, link)
			//imageRect.Bottom

		})
	})
}

// TODO: performances (zero-copy etc.)
func main() {
	{
		resp, err := http.Get("https://livesession.io/blog/")
		if err != nil {
			panic(err)
		}

		tree, err := html.Parse(resp.Body)
		if err != nil {
			panic(err)
		}

		w, err := os.Create("./original.html")
		if err != nil {
			panic(err)
		}

		if err := html.Render(w, tree); err != nil {
			panic(err)
		}
	}

	{
		u := "https://www.canva.com/newsroom/news"
		u = "https://livesession.io/blog"
		u = "https://www.canva.com/newsroom/news"

		page := rod.New().MustConnect().MustPage(u).MustWaitLoad()

		jsDOMString := page.MustEval(jsScriptSerializeDOM).Str()

		var jsDOM *DOM
		if err := json.Unmarshal([]byte(jsDOMString), &jsDOM); err != nil {
			panic(err)
		}

		tree := buildHTMLTree(jsDOM, nil)

		w, err := os.Create("./converted.html")
		if err != nil {
			panic(err)
		}

		if err := html.Render(w, tree); err != nil {
			panic(err)
		}

		doc := goquery.NewDocumentFromNode(tree)

		isBlog(doc)
	}

}
