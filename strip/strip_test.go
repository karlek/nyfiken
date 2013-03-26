package strip

import (
	"strings"
	"testing"

	"code.google.com/p/go.net/html"
	"github.com/karlek/nyfiken/settings"
)

func TestNumber(t *testing.T) {
	node1, err := html.Parse(strings.NewReader("<html><head><title>Number test 12345</title></head><body><b>I am a number 2!</b></body></html>"))
	if err != nil {
		t.Errorf("error: %s", err)
	}

	var testTable = []struct {
		input  *html.Node
		output string
	}{
		{node1, "<html><head><title>Number test </title></head><body><b>I am a number !</b></body></html>"},
	}

	for _, test := range testTable {
		numFree := Numbers(test.input)
		if numFree != test.output {
			t.Errorf("output `%v` != expected `%v`", numFree, test.output)
		}
	}
}

func TestAttrs(t *testing.T) {
	node1, err := html.Parse(strings.NewReader(`<html><head><title>Attr test</title></head><body><b style="color: #f00;">I am red!</b></body></html>`))
	if err != nil {
		t.Errorf("error: %s", err)
	}

	var testTable = []struct {
		input  *html.Node
		output string
	}{
		{node1, `<html><head><title>Attr test</title></head><body><b>I am red!</b></body></html>`},
	}

	for _, test := range testTable {
		numFree := Attrs(test.input)
		if numFree != test.output {
			t.Errorf("output `%v` != expected `%v`", numFree, test.output)
		}
	}
}

func TestHTML(t *testing.T) {
	node1, err := html.Parse(strings.NewReader(`<html><head><title>HTML test</title></head><body><b style="color: #f00;">I am red!</b></body></html>`))
	if err != nil {
		t.Errorf("error: %s", err)
	}

	var testTable = []struct {
		input  *html.Node
		output string
	}{
		{node1, `HTML test` + settings.Newline + `I am red!` + settings.Newline},
	}

	for _, test := range testTable {
		numFree := HTML(test.input)
		if numFree != test.output {
			t.Errorf("output `%v` != expected `%v`", numFree, test.output)
		}
	}
}
