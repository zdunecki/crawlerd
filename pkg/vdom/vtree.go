package vdom

import "golang.org/x/net/html"

type VTree struct {
}

func (vt *VTree) Compare(node1, node2 *html.Node, compareAttrs bool) bool {
	node1Childrens := vt.findChildNodes(node1)
	node2Childrens := vt.findChildNodes(node2)

	if len(node1Childrens) != len(node2Childrens) {
		return false
	}

	for i, root := range node1Childrens {
		otherRoot := node2Childrens[i]
		if match := vt.compareNodesRecursive(root, otherRoot, compareAttrs); !match {
			return false
		}
	}
	return true
}

func (vt *VTree) findChildNodes(node *html.Node) []*html.Node {
	var childNodes []*html.Node

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		childNodes = append(childNodes, c)
	}

	return childNodes
}

func (vt *VTree) compareNodesRecursive(node1 *html.Node, node2 *html.Node, compareAttrs bool) bool {
	if match := vt.compareNodes(node1, node2, compareAttrs); !match {
		return false
	}

	children := vt.findChildNodes(node1)
	otherChildren := vt.findChildNodes(node2)

	if len(children) != len(otherChildren) {
		return false
	}

	for i, child := range children {
		otherChild := otherChildren[i]

		if match := vt.compareNodesRecursive(child, otherChild, compareAttrs); !match {
			return false
		}
	}

	return true
}

func (vt *VTree) compareNodes(node1, node2 *html.Node, attrs bool) bool {
	if !attrs {
		return node1.Data == node2.Data
	}

	if len(node1.Attr) != len(node2.Attr) {
		return false
	}

	// TODO: attributes can be in other order
	for i, attr := range node1.Attr {
		otherAttr := node2.Attr[i]
		if attr != otherAttr {
			return false
		}
	}

	return true
}
