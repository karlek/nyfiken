// Package strip containts functions to remove false positives from comparisons
// of new and last scrape.
//
// Example: Number of posts or number of comments are very commonly changed.
// A solution is to compare the requests while ignoring numbers.
// This package seeks to solve these kind of problems.
package strip

import (
	"strings"
	"unicode"

	"code.google.com/p/go.net/html"
	"github.com/karlek/nyfiken/settings"
)

// Numbers removes numbers from all text nodes in an html.Node.
func Numbers(doc *html.Node) {
	var f func(node *html.Node)
	f = func(node *html.Node) {
		if node.Type == html.TextNode {
			newSel := ""
			for _, chr := range node.Data {
				if !unicode.IsDigit(chr) {
					newSel += string(chr)
				}
			}
			node.Data = newSel
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
}

// Attrs removes all HTML attributes from an html.Node.
func Attrs(doc *html.Node) {
	var f func(node *html.Node)
	f = func(node *html.Node) {
		if node.Type == html.ElementNode {
			node.Attr = nil
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
}

// Scripts removes all script elements from an html.Node.
func Scripts(doc *html.Node) {
	var f func(node *html.Node)
	f = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "script" {
			node = nil
			return
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
}

// HTML removes HTML tags from an html.Node and leaves the text.
func HTML(doc *html.Node) {
	var newSel string
	var f func(node *html.Node, newSel *string)
	f = func(node *html.Node, newSel *string) {
		if node.Type == html.TextNode {
			*newSel += strings.TrimSpace(node.Data) + settings.Newline
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			f(c, newSel)
		}
	}
	f(doc, &newSel)

	/// Check for errors
	stringNode, _ := html.Parse(strings.NewReader(newSel))
	*doc = *stringNode
}
