// Nyfikenc is a client program to check and handle updates from nyfiken daemon.
package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"

	"github.com/karlek/nyfiken/filename"
	"github.com/karlek/nyfiken/ini"
	"github.com/karlek/nyfiken/settings"
	"github.com/mewkiz/pkg/bufioutil"
	"github.com/mewkiz/pkg/errutil"
)

const (
	ErrNodata = 61 // No data available
)

// command-line flags
var flagRecheck bool
var flagClearAll bool
var flagReadAll bool
var flagReadAndClearAll bool

func init() {
	flag.BoolVar(&flagRecheck, "f", false, "forces a recheck.")
	flag.BoolVar(&flagReadAll, "r", false, "read all updated pages in your browser.")
	flag.BoolVar(&flagClearAll, "c", false, "will clear list of updated sites.")
	flag.BoolVar(&flagReadAndClearAll, "rc", false, "read all updated pages in your browser and clear the list of updated sites.")
	flag.Usage = usage
}

func usage() {
	fmt.Fprintln(os.Stderr, "nyfikenc [OPTION]")
	fmt.Fprintln(os.Stderr)
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr)
}

// Error wrapper.
func main() {
	flag.Parse()
	err := nyfikenc()
	if err != nil {
		log.Fatalln(err)
	}
}

func nyfikenc() (err error) {
	// Connect to nyfikend.
	conn, err := net.Dial("tcp", "localhost"+settings.Global.PortNum)
	if err != nil {
		if e, ok := err.(*net.OpError); ok {
			if e.Err.Error() == "connection refused" {
				return errutil.NewNoPos("nyfikenc: unable to connect to nyfikend. Please make sure that the daemon is running.")
			}
		} else {
			return err
		}
	}
	bw := bufioutil.NewWriter(conn)

	// Command-line flag check
	if flagRecheck ||
		flagClearAll ||
		flagReadAndClearAll ||
		flagReadAll {
		if flagRecheck {
			return force(&bw)
		}
		if flagClearAll {
			return clearAll(&bw, conn)
		}
		if flagReadAll {
			return readAll(&bw, conn)
		}
		if flagReadAndClearAll {
			err = readAll(&bw, conn)
			if err != nil {
				return err
			}
			return clearAll(&bw, conn)
		}
	}

	// If no updates where found -> apologize.
	ups, err := getUpdates(&bw, conn)
	if err != nil {
		return errutil.Err(err)
	}

	lenUps := len(ups)
	if lenUps == 0 {
		fmt.Println("Sorry, no updates :(")
		os.Exit(ErrNodata)
		return nil
	}

	for up := range ups {
		fmt.Printf("%s\n", up)
	}

	return nil
}

// Opens all links with browser.
func readAll(bw *bufioutil.Writer, conn net.Conn) (err error) {
	// Read in config file to settings.Global
	err = ini.ReadSettings(settings.ConfigPath)
	if err != nil {
		return errutil.Err(err)
	}

	ups, err := getUpdates(bw, conn)
	if err != nil {
		return errutil.Err(err)
	}

	if settings.Global.Browser == "" {
		fmt.Println("No browser path set in:", settings.ConfigPath)
		return nil
	}

	// If no updates was found, ask for forgiveness.
	if len(ups) == 0 {
		fmt.Println("Sorry, no updates :(")
		return nil
	}

	var arguments []string
	// Loop through all updates and open them with the browser
	for up := range ups {
		arguments = append(arguments, up)
	}
	cmd := exec.Command(settings.Global.Browser, arguments...)
	err = cmd.Start()
	if err != nil {
		return errutil.Err(err)
	}
	return nil
}

// Removes all updates.
func clearAll(bw *bufioutil.Writer, conn net.Conn) (err error) {
	ups, err := getUpdates(bw, conn)
	if err != nil {
		return errutil.Err(err)
	}

	for up := range ups {
		u, err := url.Parse(up)
		if err != nil {
			return errutil.Err(err)
		}

		urlAsFilename := u.Host + u.Path + u.RawQuery
		fname, err := filename.Encode(urlAsFilename)
		if err != nil {
			return errutil.Err(err)
		}

		cacheFile, err := os.Open(settings.CacheRoot + fname + ".htm")
		if err != nil {
			return errutil.Err(err)
		}
		defer cacheFile.Close()
		readFile, err := os.Create(settings.ReadRoot + fname + ".htm")
		if err != nil {
			return errutil.Err(err)
		}
		defer readFile.Close()

		_, err = io.Copy(readFile, cacheFile)
		if err != nil {
			return errutil.Err(err)
		}

		// Debug
		debugCacheFile, err := os.Open(settings.DebugCacheRoot + fname + ".htm")
		if err != nil {
			return errutil.Err(err)
		}
		defer cacheFile.Close()
		debugReadFile, err := os.Create(settings.DebugReadRoot + fname + ".htm")
		if err != nil {
			return errutil.Err(err)
		}
		defer readFile.Close()

		_, err = io.Copy(debugReadFile, debugCacheFile)
		if err != nil {
			return errutil.Err(err)
		}
	}

	// Send nyfikend a query to clear updates.
	_, err = bw.WriteLine(settings.QueryClearAll)
	if err != nil {
		return errutil.Err(err)
	}

	fmt.Println("Updates list has been cleared!")
	return nil
}

// Forces nyfikend to check all pages immediately.
func force(bw *bufioutil.Writer) (err error) {
	// Send nyfikend a query to force a recheck.
	_, err = bw.WriteLine(settings.QueryForceRecheck)
	if err != nil {
		return errutil.Err(err)
	}

	fmt.Println("Pages will be checked immediately by your demand.")
	return nil
}

// Receive updates from nyfikend.
func getUpdates(bw *bufioutil.Writer, conn net.Conn) (ups map[string]bool, err error) {
	// Ask for updates.
	_, err = bw.WriteLine(settings.QueryUpdates)
	if err != nil {
		return nil, errutil.Err(err)
	}

	// Will read from network.
	dec := gob.NewDecoder(conn)

	// Decode (receive) the value.
	err = dec.Decode(&ups)
	if err != nil {
		return nil, errutil.Err(err)
	}
	return ups, nil
}
