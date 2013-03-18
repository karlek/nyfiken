// Test cases for ini package.
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
		t.Errorf("%s", err)
	}

	/// "invalid operation: expected != settings.Global (struct containing []string cannot be compared)"
	/// Need to find a better way to compare slices.
	/// Ugly solution
	if fmt.Sprintf("%v", settings.Global) != fmt.Sprintf("%v", expected) {
		t.Errorf("Unexpected output")
	}
}

// Tests ReadPages
func TestReadPages(t *testing.T) {

	reqUrl, err := url.Parse("http://example.org")
	if err != nil {
		t.Errorf("%s", err)
	}

	anotherReqUrl, err := url.Parse("http://another.example.org")
	if err != nil {
		t.Errorf("%s", err)
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
		t.Errorf("%s", err)
	}
	/// Need to find a better way to compare slices.
	/// Ugly solution
	if len(expected) != len(pages) {
		t.Errorf("Unexpected output")
	}
	for _, expectedP := range expected {
		pageFound := false
		for _, p := range pages {
			// Compare all fields that have defined equality.
			if *p.ReqUrl == *expectedP.ReqUrl &&
				p.Settings.Interval == expectedP.Settings.Interval &&
				p.Settings.Negexp == expectedP.Settings.Negexp &&
				p.Settings.Regexp == expectedP.Settings.Regexp &&
				p.Settings.RecvMail == expectedP.Settings.RecvMail &&
				p.Settings.Selection == expectedP.Settings.Selection &&
				p.Settings.Threshold == expectedP.Settings.Threshold &&
				isStripFuncsEqual(p.Settings.StripFuncs, expectedP.Settings.StripFuncs) &&
				isHeadersEqual(p.Settings.Header, expectedP.Settings.Header) {
				pageFound = true
			}
		}
		if !pageFound {
			t.Errorf("Unexpected output")
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
