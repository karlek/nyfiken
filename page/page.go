// Package page contains functions which checks if a page has been updated.
package page

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"code.google.com/p/cascadia"
	"code.google.com/p/go.net/html"
	"code.google.com/p/mahonia"
	"github.com/karlek/nyfiken/filename"
	"github.com/karlek/nyfiken/mail"
	"github.com/karlek/nyfiken/settings"
	"github.com/karlek/nyfiken/strip"
	"github.com/karlek/nyfiken/strmetr"
	"github.com/mewkiz/pkg/errutil"
	"github.com/mewkiz/pkg/htmlutil"
)

// Page is a site which is checked for changes. It has specialized settings to
// eliminate false-positives.
type Page struct {
	ReqUrl   *url.URL
	Settings settings.Page
}

// Check downloads and makes a specialized comparison with a previous check
// saved on disk to determine if the page has been updated. Check takes
// an error channel to concurrently handle errors.
func (p *Page) Check(ch chan<- error) {
	ch <- p.check()
}

// check is an non-exported function for better error handling.
func (p *Page) check() (err error) {
	if settings.Verbose {
		fmt.Println("[/] Downloading:", p.ReqUrl.String())
	}

	// Retrieve result from download or return timeout error.
	var r struct {
		*html.Node
		error
	}
	select {
	case r = <-p.errWrapDownload():
		if r.error != nil {
			return errutil.Err(r.error)
		}
	case <-time.After(settings.TimeoutDuration):
		return errutil.NewNoPosf("timeout: %s", p.ReqUrl.String())
	}

	// Extract selection from downloaded source.
	selection, err := p.makeSelection(r.Node)
	if err != nil {
		return errutil.Err(err)
	}

	// File name is a escaped URL in a cache folder.
	linuxPath, err := filename.Encode(p.ReqUrl.String())
	if err != nil {
		return errutil.Err(err)
	}
	cachePathName := settings.CacheRoot + linuxPath + ".htm"

	// Read in comparison.
	buf, err := ioutil.ReadFile(cachePathName)
	if err != nil {
		if !os.IsNotExist(err) {
			return errutil.Err(err)
		}

		// If the page hasn't been checked before, create a new comparison file.
		err = ioutil.WriteFile(
			cachePathName,
			[]byte(selection),
			settings.Global.FilePerms,
		)
		if err != nil {
			return errutil.Err(err)
		}

		prevPathName := settings.PrevRoot + linuxPath + ".htm"
		// If the page hasn't been checked before, create a new comparison file.
		err = ioutil.WriteFile(
			prevPathName,
			[]byte(selection),
			settings.Global.FilePerms,
		)
		if err != nil {
			return errutil.Err(err)
		}

		if settings.Verbose {
			fmt.Println("[+] New site added:", p.ReqUrl.String())
		}

		return nil
	}

	// The distance between to strings in percentage.
	dist := strmetr.Approx(string(buf), selection)

	// If the distance is within the threshold level, i.e if the check was a
	// match.
	if dist > p.Settings.Threshold {
		u := settings.Update{p.ReqUrl.String()}
		settings.Updates[u] = true

		if settings.Verbose {
			fmt.Println("[!] Updated:", p.ReqUrl.String())
		}

		// Save updates to file.
		err = settings.SaveUpdates()
		if err != nil {
			return errutil.Err(err)
		}

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
			mailPage.Settings.Regexp = ""
			sel, err := mailPage.makeSelection(r.Node)
			if err != nil {
				return errutil.Err(err)
			}

			err = mail.Send(*p.ReqUrl, p.Settings.RecvMail, sel)
			if err != nil {
				return errutil.Err(err)
			}
		}

		// Update the comparison file.
		err = ioutil.WriteFile(cachePathName, []byte(selection), settings.Global.FilePerms)
		if err != nil {
			return errutil.Err(err)
		}
	} else {
		if settings.Verbose {
			fmt.Println("[-] No update:", p.ReqUrl.String())
		}
	}
	return nil
}

// An error wrapping convenience function for p.download() used because of
// timeout implementation.
// Credits to: Dave Cheney and ilyia (https://groups.google.com/forum/?fromgroups=#!topic/golang-nuts/cTrBcyjqCxg)
func (p *Page) errWrapDownload() <-chan struct {
	*html.Node
	error
} {
	doc, err := p.download()
	result := make(chan struct {
		*html.Node
		error
	})
	go func() {
		result <- struct {
			*html.Node
			error
		}{doc, err}
	}()
	return result
}

// Download the page with or without user specified headers.
func (p *Page) download() (doc *html.Node, err error) {

	// Construct the request.
	req, err := http.NewRequest("GET", p.ReqUrl.String(), nil)
	if err != nil {
		return nil, errutil.Err(err)
	}

	// If special headers were specified, add them to the request.
	if p.Settings.Header != nil {
		for key, val := range p.Settings.Header {
			req.Header.Add(key, val)
		}
	}

	// Do request and read response.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errutil.Err(err)
	}
	defer resp.Body.Close()

	// If response contained a client or server error, fail with that error.
	if resp.StatusCode >= 400 {
		return nil, errutil.Newf("%s: (%d) - %s", p.ReqUrl.String(), resp.StatusCode, resp.Status)
	}

	// Read the response body to []byte.
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errutil.Err(err)
	}

	// Fix charset problems with servers that doesn't use utf-8
	charset := "utf-8"
	content := string(buf)

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
		content = mahonia.NewDecoder(charset).ConvertString(content)
	}
	// Parse response into html.Node.
	return html.Parse(strings.NewReader(content))
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
			return "", errutil.Err(err)
		}

		// Find all nodes that matches selection s.
		result = s.MatchAll(htmlNode)
	}

	// Loop through all the hits and render them to string.
	for _, hit := range result {
		selection += htmlutil.RenderToString(hit)
	}

	// --- [ /CSS selection ] -------------------------------------------------/

	// --- [ Strip funcs ] ----------------------------------------------------/

	for _, stripFunc := range p.Settings.StripFuncs {
		doc, err := html.Parse(strings.NewReader(selection))
		if err != nil {
			return "", errutil.Err(err)
		}
		stripFunc = strings.ToLower(stripFunc)
		switch stripFunc {
		case "numbers":
			strip.Numbers(doc)
		case "attrs":
			strip.Attrs(doc)
		case "html":
			strip.HTML(doc)
		case "scripts":
			strip.Scripts(doc)
		}

		selection, err = htmlutil.RenderClean(doc)
		if err != nil {
			return "", errutil.Err(err)
		}
	}

	// --- [ /Strip funcs ] ---------------------------------------------------/

	// --- [ Regexp ] ---------------------------------------------------------/

	if p.Settings.Regexp != "" {
		re, err := regexp.Compile(p.Settings.Regexp)
		if err != nil {
			return "", errutil.Err(err)
		}

		// -1 means to find all.
		result := re.FindAllString(selection, -1)

		selection = ""
		for _, res := range result {
			selection += res + settings.Newline
		}
	}

	// --- [ /Regexp ] --------------------------------------------------------/

	// --- [ Negexp ] ---------------------------------------------------------/

	if p.Settings.Negexp != "" {
		ne, err := regexp.Compile(p.Settings.Negexp)
		if err != nil {
			return "", errutil.Err(err)
		}

		// Remove all that matches the regular expression ne
		selection = ne.ReplaceAllString(selection, "")
	}

	// --- [ /Negexp ] --------------------------------------------------------/

	return selection, nil
}
