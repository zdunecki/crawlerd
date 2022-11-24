package analyze

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod/lib/devices"
	"golang.org/x/net/html"
	"strings"
)

func getAttribute(attrName string, n *html.Node) *html.Attribute {
	if n == nil {
		return nil
	}

	for i, a := range n.Attr {
		if a.Key == attrName {
			return &n.Attr[i]
		}
	}
	return nil
}

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

// TODO: get dom only visible on screen
func buildHTMLTree(el *DOM, node *html.Node) *html.Node {
	newNode := convertJSDOMToHTMLNode(el, node)

	for _, c := range el.ChildNodes {
		newNode.AppendChild(buildHTMLTree(c, node))
	}

	return newNode
}

func visibleOnScreen(device devices.Device, el *DOM) bool {
	if el == nil {
		return false
	}
	if el.Rect == nil {
		return true
	}
	return el.Rect.Top <= float64(device.Screen.Horizontal.Height)
}

func buildVisibleHTMLTree(device devices.Device, el *DOM, node *html.Node) *html.Node {
	newNode := convertJSDOMToHTMLNode(el, node)

	for _, c := range el.ChildNodes {
		if visibleOnScreen(device, c) {
			newNode.AppendChild(buildVisibleHTMLTree(device, c, node))
		}
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

			txt := strings.TrimSpace(el.Data)
			if txt == "" {
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

func intersect(rect1 *Rect, rect2 *Rect) bool {
	return (rect1.Left >= rect2.Left && rect1.Left <= rect2.Right || rect1.Right >= rect2.Left && rect1.Right <= rect2.Right) &&
		(rect1.Top >= rect2.Top && rect1.Top <= rect2.Bottom || rect1.Bottom >= rect2.Top && rect1.Bottom <= rect2.Bottom)
}

func getTextNodesFromLink(link *goquery.Selection) map[*html.Node]*Rect {
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

	return textNodeRects
}

func findFirstInNeighbourhood(el *goquery.Selection, selector string) (found *goquery.Selection) {
	var f func(*goquery.Selection)

	f = func(el *goquery.Selection) {
		if el == nil || len(el.Nodes) <= 0 {
			return
		}

		search := el.NextAll().Find(selector)

		if search.Size() == 0 {
			f(el.Parent())
		} else {
			found = search
		}
	}

	f(el)

	return
}

// TODO: tweak this function
func isNavbar(device devices.Device, rect *Rect) bool {
	minTop := float64(device.Screen.Horizontal.Height) * 0.15

	if rect.Top <= minTop {
		return true
	}

	return false
}

func searchArticleByLinkIntersect() {}

func searchArticleByLinkParent(device devices.Device, link *goquery.Selection, images *goquery.Selection) (bool, float64) {
	minTop := 0.0
	foundArticle := false
	textNodeRects := getTextNodesFromLink(link)

	// search in siblings
	if len(textNodeRects) == 0 {
		if searchLinks := findFirstInNeighbourhood(link, "a"); searchLinks != nil {
			textNodeRects = getTextNodesFromLink(searchLinks)
		}

		// TODO: remove below comments
		//var s func(selection *goquery.Selection)
		//
		//s = func(el *goquery.Selection) {
		//	searchLinks := el.NextAll().Find("a")
		//
		//	if searchLinks.Size() == 0 {
		//		s(el.Parent())
		//	} else {
		//		textNodeRects = getTextNodesFromLink(searchLinks)
		//	}
		//}
		//
		//s(link)
	}

	//fmt.Println("ok")
	images.Each(func(i int, image *goquery.Selection) {
		imageRect := getRect(image)
		if imageRect == nil || imageRect.Bottom == 0 || imageRect.Right == 0 {
			// TODO: log err
			return
		}

		if len(textNodeRects) <= 0 {
			return
		}

		// TODO: all is not needed just greater than half should be enough or only the largest text?
		textNearImage := false

		for _, t := range textNodeRects {
			textIsBelowImage := t.Top >= imageRect.Bottom && imageRect.Left <= t.Left && imageRect.Right >= t.Right
			if textIsBelowImage {
				textNearImage = true
				continue
			}

			textIsRightFromImage := t.Left >= imageRect.Right
			if textIsRightFromImage {
				textNearImage = true
				continue
			}

			textIsLeftFromImage := t.Right <= imageRect.Left
			if textIsLeftFromImage {
				textNearImage = true
				continue
			}

			textWithinImage := intersect(t, imageRect)
			if textWithinImage {
				textNearImage = true
				continue
			}

			textNearImage = false
			break
		}

		if textNearImage {
			// TODO: better algorithm for getting top of image/text (avg ?)
			if minTop == 0 || imageRect.Top < minTop {
				minTop = imageRect.Top
			}

			// it's for debug
			//{
			//	for v := range textNodeRects {
			//		fmt.Println(v.Data)
			//	}
			//	fmt.Println(isNavbar(device, imageRect))
			//}

			if !isNavbar(device, imageRect) {
				foundArticle = true
			}
		}
	})

	return foundArticle, minTop
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
// TODO: if two potential blogs has a similar cardinality of blogs then check few contents of them to check if it's engineering blog
// TODO: if lowest top value is higher than window height than we can assume it's not a blog
func IsArticle(device devices.Device, doc *goquery.Document) (bool, int, error) {
	links := doc.Find("a")

	foundArticleCount := 0
	minTop := 0.0

	var globlaErr error

	// TODO: mark node as checked
	links.Each(func(i int, link *goquery.Selection) {
		images := link.Find("img")

		// TODO: if link is children/sibling of image
		if images == nil || images.Nodes == nil {
			images = link.Find("[crawlerd-is-image]") // TODO: issues with this technique
		}

		if images == nil || images.Nodes == nil {
			// find first image in neaighourhood and text but all must be within link box
			// TODO: maybe only part of box?
			//abc := findFirstInNeighbourhood(link, "[crawlerd-is-image]")
			//
			//if abc != nil {
			//	fmt.Println(abc)
			//}

			fmt.Println(link.Nodes[0].Attr)
			return
		}

		if found, localMinTop := searchArticleByLinkParent(device, link, images); found {
			if minTop == 0 || localMinTop < minTop {
				minTop = localMinTop
			}
			foundArticleCount++
		}
	})

	firstArticleHigherThanDevice := minTop != 0.0 && minTop > float64(device.Screen.Horizontal.Height)

	if firstArticleHigherThanDevice {
		foundArticleCount = 0
	}

	if minTop == 0.0 && foundArticleCount > 0 {
		return false, foundArticleCount, errors.New("found articles but min top is zero")
	}

	return minTop != 0.0 && foundArticleCount > 0, foundArticleCount, globlaErr
}
