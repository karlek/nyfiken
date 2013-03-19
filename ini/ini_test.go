package ini

import (
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/karlek/nyfiken/page"
	"github.com/karlek/nyfiken/settings"
)

// Tests ReadSettings
func TestReadSettings(t *testing.T) {
	// Expected output of ReadSettings.
	expected := settings.Prog{
		Interval:  10 * time.Minute,
		RecvMail:  "global@example.com",
		FilePerms: os.FileMode(0777),
		PortNum:   ":4113",
		Browser:   "/usr/bin/browser",

		SenderMail: struct {
			Address    string
			Password   string
			AuthServer string
			OutServer  string
		}{
			Address:    "sender@example.com",
			Password:   "123456",
			AuthServer: "auth.server.com",
			OutServer:  "out.server.com:587",
		},
	}

	err := ReadSettings("ini_test_config.ini")
	if err != nil {
		t.Errorf("ReadSettings: %s", err)
	}

	/// "invalid operation: expected != settings.Global (struct containing []string cannot be compared)"
	/// Need to find a better way to compare slices.
	/// Ugly solution
	if fmt.Sprintf("%v", settings.Global) != fmt.Sprintf("%v", expected) {
		t.Errorf("output %v != %v", settings.Global, expected)
	}
}

// Tests ReadPages
func TestReadPages(t *testing.T) {
	reqUrl, err := url.Parse("http://example.org")
	if err != nil {
		t.Errorf("url.Parse: %s", err)
	}
	anotherReqUrl, err := url.Parse("http://another.example.org")
	if err != nil {
		t.Errorf("url.Parse: %s", err)
	}

	expected := []*page.Page{
		&page.Page{
			ReqUrl: reqUrl,
			Settings: settings.Page{
				Interval:  3 * time.Minute,
				Threshold: 0.05,
				RecvMail:  "mail@example.org",
				Selection: "html body",
				StripFuncs: []string{
					"html",
					"numbers",
				},
				Regexp: "(love)",
				Negexp: "(hate)",
				Header: map[string]string{
					"Cookie":     "IloveCookies=1;",
					"User-Agent": "I come in peace",
				},
			},
		},
		&page.Page{
			ReqUrl: anotherReqUrl,
			Settings: settings.Page{
				Interval:  settings.Global.Interval,
				RecvMail:  settings.Global.RecvMail,
				Selection: "#main-content",
			},
		},
	}

	pages, err := ReadPages("ini_test_pages.ini")
	if err != nil {
		t.Errorf("ReadPages: %s", err)
	}
	/// Need to find a better way to compare slices.
	/// Ugly solution
	if len(expected) != len(pages) {
		t.Errorf("output length (%d) != (%d) expected length", len(pages), len(expected))
	}
	for _, expectedP := range expected {
		pageFound := false
		for _, p := range pages {
			// If error message is only page not found, two or more pages have
			// only different URLs.
			if *p.ReqUrl != *expectedP.ReqUrl {
				continue
			}

			// Compare all fields that have defined equality.
			switch {
			case p.Settings.Interval != expectedP.Settings.Interval:
				t.Errorf("interval output %v != %v", p.Settings.Interval, expectedP.Settings.Interval)
			case p.Settings.Negexp != expectedP.Settings.Negexp:
				t.Errorf("Negexp output %v != %v", p.Settings.Negexp, expectedP.Settings.Negexp)
			case p.Settings.Regexp != expectedP.Settings.Regexp:
				t.Errorf("Regexp output %v != %v", p.Settings.Regexp, expectedP.Settings.Regexp)
			case p.Settings.RecvMail != expectedP.Settings.RecvMail:
				t.Errorf("RecvMail output %v != %v", p.Settings.RecvMail, expectedP.Settings.RecvMail)
			case p.Settings.Selection != expectedP.Settings.Selection:
				t.Errorf("Selection output %v != %v", p.Settings.Selection, expectedP.Settings.Selection)
			case p.Settings.Threshold != expectedP.Settings.Threshold:
				t.Errorf("Threshold output %v != %v", p.Settings.Threshold, expectedP.Settings.Threshold)
			case !isStripFuncsEqual(p.Settings.StripFuncs, expectedP.Settings.StripFuncs):
				t.Errorf("StripFuncs output %v != %v", p.Settings.StripFuncs, expectedP.Settings.StripFuncs)
			case !isHeadersEqual(p.Settings.Header, expectedP.Settings.Header):
				t.Errorf("Header output %v != %v", p.Settings.Header, expectedP.Settings.Header)
			default:
				pageFound = true
				break
			}
		}
		if !pageFound {
			t.Errorf("Page not found")
		}
	}
}

// Temporary equality function for string slices.
func isStripFuncsEqual(strip, expected []string) bool {
	if len(strip) != len(expected) {
		return false
	}
	for _, stripFunc := range strip {
		funcFound := false
		for _, expectedStripFunc := range expected {
			if expectedStripFunc == stripFunc {
				funcFound = true
				break
			}
		}
		if !funcFound {
			return false
		}
	}
	return true
}

// Temporary equality function for map[string]string.
func isHeadersEqual(headers, expected map[string]string) bool {
	if len(expected) != len(headers) {
		return false
	}
	for key, val := range headers {
		if expected[key] != val {
			return false
		}
	}
	return true
}
