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

/// Returns a number free string.
func Numbers(doc *html.Node) {
	var f func(node *html.Node)
	f = func(node *html.Node) {
		if node.Type == html.TextNode {
			text := strings.TrimSpace(node.Data)
			node.Data = ""
			for _, chr := range text {
				if !unicode.IsDigit(chr) {
					node.Data += string(chr)
				}
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
}

/// Returns a string with empty HTML attributes.
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

// Returns a HTML free string.
func HTML(doc *html.Node) (newSel string) {
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

	return newSel
}
