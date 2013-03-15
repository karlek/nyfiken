// Package page contains functions which checks if a page has been updated.
package page

import "io/ioutil"
import "net/http"
import "net/url"
import "os"
import "regexp"
import "strings"
import "time"

import "code.google.com/p/cascadia"
import "code.google.com/p/go.net/html"
import "code.google.com/p/mahonia"
import "github.com/karlek/nyfiken/filename"
import "github.com/karlek/nyfiken/mail"
import "github.com/karlek/nyfiken/settings"
import "github.com/karlek/nyfiken/strip"
import "github.com/karlek/nyfiken/strmetr"
import "github.com/mewkiz/pkg/errorsutil"
import "github.com/mewkiz/pkg/htmlutil"

// Pages are checked if they have changed and based on user settings
// special selections are made to eliminate false-positives.
type Page struct {
	ReqUrl   *url.URL
	Settings settings.Page
}

// Result from Page.download() method, wrapped because of timeout implementation.
type result struct {
	node *html.Node
	err  error
}

// Check downloads the page, extracts the wanted selection and makes a comparison
// with a previous check to determine if the page has been updated. Check takes
// an error channel to send errors to be printed.
func (p *Page) Check(ch chan<- error) {
	// A buffered channel with size 1.
	downloadOrTimeoutChan := make(chan *result, 1)

	// Returned result from download.
	var r *result

	// Download page.
	go func() { downloadOrTimeoutChan <- p.errWrapDownload() }()

	// Retrieve result from download or return timeout error.
	select {
	case r = <-downloadOrTimeoutChan:
		if r.err != nil {
			ch <- r.err
			return
		}
	case <-time.After(settings.TimeoutDuration):
		ch <- errorsutil.ErrorfColor("timeout: %s", p.ReqUrl.String())
		return
	}

	// Extract selection from downloaded source.
	selection, err := p.makeSelection(r.node)
	if err != nil {
		ch <- err
		return
	}

	// File name is a escaped URL in a cache folder.
	cachePathName := settings.CacheRoot + filename.Linux(p.ReqUrl.String()) + ".htm"

	// Read in comparison.
	buf, err := ioutil.ReadFile(cachePathName)
	if err != nil {
		// If the page hasn't been checked before, create a new comparison file.
		if os.IsNotExist(err) {
			err = ioutil.WriteFile(
				cachePathName,
				[]byte(selection),
				settings.Global.FilePerms,
			)
			if err != nil {
				ch <- errorsutil.ErrorfColor("%s", err)
				return
			}
			ch <- nil
			return
		} else {
			ch <- errorsutil.ErrorfColor("%s", err)
			return
		}
	}

	// The distance between to strings in percentage.
	dist := strmetr.Approx(string(buf), selection)
	// If the distance is within the threshold level, i.e if the check was a
	// match.
	if dist > p.Settings.Threshold {
		u := settings.Update{p.ReqUrl.String()}
		settings.Updates[u] = true

		// If the page has a mail and all compulsory global mail settings are
		// set, send a mail to notify the user about an update.
		if p.Settings.RecvMail != "" &&
			settings.Global.SenderMail.AuthServer != "" &&
			settings.Global.SenderMail.OutServer != "" &&
			settings.Global.SenderMail.Address != "" {

			// Mail the selection without the stripping functions, since their
			// only purpose is to remove false-positives. It will make the
			// output look better.
			mailPage := Page{p.ReqUrl, p.Settings}
			mailPage.Settings.StripFuncs = nil
			sel, err := mailPage.makeSelection(r.node)
			if err != nil {
				ch <- err
				return
			}

			err = mail.Send(*p.ReqUrl, p.Settings.RecvMail, sel)
			if err != nil {
				ch <- err
				return
			}
		}

		// Update the comparison file.
		err = ioutil.WriteFile(cachePathName, []byte(selection), settings.Global.FilePerms)
		if err != nil {
			ch <- errorsutil.ErrorfColor("%s", err)
			return
		}
	}
	ch <- nil
}

// An error wrapping convenience function for p.download() used because of
// timeout implementation.
func (p *Page) errWrapDownload() *result {
	doc, err := p.download()
	return &result{
		doc,
		err,
	}
}

// Download the page with or without user specified headers.
func (p *Page) download() (doc *html.Node, err error) {

	// Client to preform the requests.
	var client = http.DefaultClient

	// Construct the request.
	req, err := http.NewRequest("GET", p.ReqUrl.String(), nil)
	if err != nil {
		return nil, errorsutil.ErrorfColor("%s", err)
	}

	// If special headers were specified, add them to the request.
	if p.Settings.Header != nil {
		for key, val := range p.Settings.Header {
			req.Header.Add(key, val)
		}
	}

	// Do request and read response.
	resp, err := client.Do(req)
	if err != nil {
		return nil, errorsutil.ErrorfColor("%s", err)
	}
	defer resp.Body.Close()

	// If response contained a client or server error, fail with that error.
	if resp.StatusCode >= 400 {
		return nil, errorsutil.ErrorfColor("%s: (%d) - %s", p.ReqUrl.String(), resp.StatusCode, resp.Status)
	}

	// Read the response body to []byte.
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errorsutil.ErrorfColor("%s", err)
	}

	// Fix charset problems with servers that doesn't use utf-8
	charset := "utf-8"
	s := string(buf)

	types := strings.Split(resp.Header.Get("Content-Type"), ` `)
	for _, typ := range types {
		if strings.Contains(typ, "charset") {
			keyval := strings.Split(typ, `=`)
			if len(keyval) == 2 {
				charset = keyval[1]
			}
		}
	}
	if charset != "utf-8" {
		s = mahonia.NewDecoder(charset).ConvertString(s)
	}
	// Parse response into html.Node.
	return html.Parse(strings.NewReader(s))
}

// Select from the retrived page source the CSS selection defined in c4c.ini.
func (p *Page) makeSelection(htmlNode *html.Node) (selection string, err error) {

	// --- [ CSS selection ] --------------------------------------------------/

	// Write results into an array of nodes.
	var result []*html.Node

	// Append the whole page (htmlNode) to results if no selector where chosen.
	if p.Settings.Selection == "" {
		result = append(result, htmlNode)
	} else {

		// Make a selector from the user specified string.
		s, err := cascadia.Compile(p.Settings.Selection)
		if err != nil {
			return "", errorsutil.ErrorfColor("%s", err)
		}

		// Find all nodes that matches selection s.
		result = s.MatchAll(htmlNode)
	}

	// Loop through all the hits and render them to string.
	for _, hit := range result {
		selection += htmlutil.RenderToString(hit)
	}
	doc, err := html.Parse(strings.NewReader(selection))
	if err != nil {
		return "", errorsutil.ErrorfColor("%s", err)
	}

	// --- [ /CSS selection ] -------------------------------------------------/

	// --- [ Strip funcs ] ----------------------------------------------------/

	for _, stripFunc := range p.Settings.StripFuncs {
		stripFunc = strings.ToLower(stripFunc)
		switch stripFunc {
		case "numbers":
			selection = strip.Numbers(doc)
		case "attrs":
			selection = strip.Attrs(doc)
		case "html":
			selection = strip.HTML(doc)
		}
		doc, err = html.Parse(strings.NewReader(selection))
		if err != nil {
			return "", errorsutil.ErrorfColor("%s", err)
		}
	}

	// --- [ /Strip funcs ] ---------------------------------------------------/

	// --- [ Regexp ] ---------------------------------------------------------/

	if p.Settings.Regexp != "" {
		re, err := regexp.Compile(p.Settings.Regexp)
		if err != nil {
			return "", errorsutil.ErrorfColor("%s", err)
		}

		// -1 means to find all.
		result := re.FindAllString(selection, -1)

		selection = ""
		for _, res := range result {
			selection += res + settings.Global.Newline
		}
	}

	// --- [ /Regexp ] --------------------------------------------------------/

	// --- [ Negexp ] ---------------------------------------------------------/

	if p.Settings.Negexp != "" {
		ne, err := regexp.Compile(p.Settings.Negexp)
		if err != nil {
			return "", errorsutil.ErrorfColor("%s", err)
		}

		// Remove all that matches the regular expression ne
		selection = ne.ReplaceAllString(selection, "")
	}

	// --- [ /Negexp ] --------------------------------------------------------/

	return selection, nil
}
