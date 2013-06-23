// Nyfikenc is a client program to check and handle updates from nyfiken daemon.
package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"

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

func init() {
	flag.BoolVar(&flagRecheck, "f", false, "forces a recheck.")
	flag.BoolVar(&flagReadAll, "r", false, "read all updated pages in your browser.")
	flag.BoolVar(&flagClearAll, "c", false, "will clear list of updated sites.")
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
		flagReadAll {
		if flagRecheck {
			return force(&bw)
		}
		if flagClearAll {
			return clearAll(&bw)
		}
		if flagReadAll {
			return readAll(&bw, conn)
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

	for up, _ := range ups {
		fmt.Printf("%s\n", up.ReqUrl)
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

	argumentStr := ""
	// Loop through all updates and open them with the browser
	for up, _ := range ups {
		argumentStr += up.ReqUrl + " "
	}
	cmd := exec.Command(settings.Global.Browser, argumentStr)
	err = cmd.Start()
	if err != nil {
		return errutil.Err(err)
	}
	err = cmd.Wait()
	if err != nil {
		return errutil.Err(err)
	}
	fmt.Println("arg string")
	// fmt.Println("Opening all updates with:", settings.Global.Browser)
	return nil
}

// Removes all updates.
func clearAll(bw *bufioutil.Writer) (err error) {
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
func getUpdates(bw *bufioutil.Writer, conn net.Conn) (ups map[settings.Update]bool, err error) {
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
