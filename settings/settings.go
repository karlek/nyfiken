// Package settings contains default- and user-settings for nyfikenc/d.
package settings

import (
	"log"
	"os"
	"time"

	"github.com/mewkiz/pkg/osutil"
)

// When an update is found, log it incase we get asked by nyfikenc for updates.
type Update struct {
	ReqUrl string
}

// Page settings on how the program shall check the page.
type Page struct {
	Interval   time.Duration     // Duration of time to wait between scrapes.
	Threshold  float64           // Percentage of accepted deviation from last scrape.
	RecvMail   string            // Mail address to send a notification when a page has been updated.
	Regexp     string            // Regular expression to further specify what to select.
	Negexp     string            // Removes with regular expression everything that matches.
	StripFuncs []string          // Strip functions to further specify what to select.
	Header     map[string]string // HTTP headers to request targeted site with.
	Selection  string            // CSS selector string to specify what to select.
}

// Program global settings which regards all pages unless overwritten with page
// specific.
type Prog struct {
	Interval   time.Duration // Duration of time to wait between scrapes.
	RecvMail   string        // Mail address to send a notification when a page has been updated.
	StripFuncs []string      // Strip functions to further specify what to select.
	FilePerms  os.FileMode   // Permissions to create files with.
	PortNum    string        // On which port should the nyfikenc/d communication take place.
	Browser    string        // The path to the browser to open updates in.

	// Information about the mail address to send updates.
	SenderMail struct {
		Address    string // Mail address of the sending mail.
		Password   string // Password to that mail address.
		AuthServer string // Authorization server to the mail address.
		OutServer  string // Outgoing server to the mail address.
	}
}

const (
	// Queries sent from the client to the daemon.
	QueryClearAll     = "clear all!"
	QueryForceRecheck = "recheck!"
	QueryUpdates      = "updates?"

	// Default interval between updates unless overwritten in config file.
	DefaultInterval = 1 * time.Minute

	// Default permissions to create files: user read and write permissions.
	DefaultFilePerms = os.FileMode(0600)

	// Default newline character
	Newline = "\n"

	// Default port number for nyfikenc/d connection.
	DefaultPortNum = ":5239"
)

var (
	// Paths to nyfiken files.
	NyfikenRoot string
	ConfigPath  string
	PagesPath   string
	CacheRoot   string

	// A map of updates that have been logged.
	Updates map[Update]bool

	// Duration until a timeout is issued
	TimeoutDuration = 10 * time.Second

	// Settings which will be used unless overwritten by site-specific settings.
	Global = Prog{
		Interval:  DefaultInterval,
		FilePerms: DefaultFilePerms,
		PortNum:   DefaultPortNum,
	}
)

// Error wrapper.
func init() {
	err := initialize()
	if err != nil {
		log.Fatalln(err)
	}
}

func initialize() (err error) {
	Updates = make(map[Update]bool)

	NyfikenRoot = os.Getenv("HOME") + "/.config/nyfiken/"
	ConfigPath = NyfikenRoot + "config.ini"
	PagesPath = NyfikenRoot + "pages.ini"
	CacheRoot = NyfikenRoot + "cache/"

	// Create a nyifken config folder if it doesn't exist.
	found, err := osutil.Exists(NyfikenRoot)
	if err != nil {
		return err
	}
	if !found {
		err := os.Mkdir(NyfikenRoot, DefaultFilePerms)
		if err != nil {
			return err
		}
	}

	found, err = osutil.Exists(PagesPath)
	if err != nil {
		return err
	}
	if !found {
		err := os.Mkdir(PagesPath, DefaultFilePerms)
		if err != nil {
			return err
		}
	}

	return nil
}
