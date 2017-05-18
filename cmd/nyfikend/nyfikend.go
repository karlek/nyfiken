// Nyfikend is a daemon which checks if pages have been updated and notifies the
// user.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/howeyc/fsnotify"
	"github.com/karlek/nyfiken/cli"
	"github.com/karlek/nyfiken/filename"
	"github.com/karlek/nyfiken/ini"
	"github.com/karlek/nyfiken/page"
	"github.com/karlek/nyfiken/settings"
	"github.com/mewkiz/pkg/errutil"
)

var flagClean bool

func init() {
	flag.BoolVar(&settings.Verbose, "v", false, "Verbose.")
	flag.BoolVar(&flagClean, "c", false, "Remove old cache files.")
	flag.Usage = usage
}

func usage() {
	fmt.Fprintln(os.Stderr, "nyfikend [OPTION]")
	fmt.Fprintln(os.Stderr)
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr)
}

// Error wrapper.
func main() {
	flag.Parse()
	err := nyfikend()
	if err != nil {
		log.Fatalln(errutil.Err(err))
	}
}

var pages []*page.Page

func nyfikend() (err error) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if flagClean {
		return clean()
	}

	pages, err = ini.ReadIni(settings.ConfigPath, settings.PagesPath)
	if err != nil {
		return errutil.Err(err)
	}

	// Change settings files only when config files are modified.
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		return errutil.Err(err)
	}
	go errWrapWatchConfig(watcher)
	err = watcher.Watch(settings.ConfigPath)
	if err != nil {
		return errutil.Err(err)
	}
	err = watcher.Watch(settings.PagesPath)
	if err != nil {
		return errutil.Err(err)
	}

	// Listen for nyfikenc queries.
	go cli.Listen()

	var secondsElapsed float64
	for ; ; secondsElapsed++ {
		// A channel in which errors are sent from p.Check().
		errChan := make(chan error)

		// The number of checks currently taking place.
		var numChecks int
		for _, p := range pages {
			// If the seconds elapsed modulo the duration of the interval in
			// seconds is equal to zero, the page should be checked.
			if math.Mod(float64(secondsElapsed), p.Settings.Interval.Seconds()) != 0 {
				continue
			}
			// Start a go-routine to check if the page has been updated.
			go p.Check(errChan)
			numChecks++
		}

		// For each check that took place, listen if any check returned an
		// error.
		go func(ch chan error, nChecks int) {
			for i := 0; i < nChecks; i++ {
				if err := <-ch; err != nil {
					log.Println(errutil.Err(err))
				}
			}
		}(errChan, numChecks)

		time.Sleep(1 * time.Second)
	}

	return nil
}

// Reads config files only when they are modified.
func errWrapWatchConfig(watcher *fsnotify.Watcher) {
	err := watchConfig(watcher)
	if err != nil {
		log.Fatalln(errutil.Err(err))
	}
}

func watchConfig(watcher *fsnotify.Watcher) (err error) {
	for {
		select {
		case ev := <-watcher.Event:
			if ev != nil {
				if !ev.IsModify() {
					logrus.Println(ev)
					continue
				}
				logrus.Println(ev)
				if ev.Name == settings.ConfigPath {
					// Read settings from config file.
					err = ini.ReadSettings(settings.ConfigPath)
					if err != nil {
						return errutil.Err(err)
					}
				}
				// Retrieve an array of pages from INI file.
				pages, err := ini.ReadPages(settings.PagesPath)
				if err != nil {
					return errutil.Err(err)
				}
				err = page.ForceUpdate(pages)
				if err != nil {
					return errutil.Err(err)
				}
			}
		case err = <-watcher.Error:
			if err != nil {
				return errutil.Err(err)
			}
		}
	}
}

// clean removes old cache files from cache root.
func clean() (err error) {
	// Get a list of all pages.
	pages, err := ini.ReadPages(settings.PagesPath)
	if err != nil {
		return errutil.Err(err)
	}

	// Get a list of all cached pages.
	caches, err := ioutil.ReadDir(settings.CacheRoot)
	if err != nil {
		return errutil.Err(err)
	}

	for _, cache := range caches {
		remove := true
		for _, p := range pages {
			pageName, err := filename.Encode(p.UrlAsFilename())
			if err != nil {
				return errutil.Err(err)
			}
			if cache.Name() == pageName+".htm" {
				remove = false
				break
			}
		}
		if remove {
			err = os.Remove(settings.CacheRoot + cache.Name())
			if err != nil {
				return err
			}
		}
	}

	return nil
}
