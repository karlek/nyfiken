// Nyfikend is a daemon which checks if pages have been updated and notifies the
// user.
package main

import "log"
import "math"
import "runtime"
import "time"

import "github.com/karlek/nyfiken/cli"
import "github.com/karlek/nyfiken/ini"
import "github.com/karlek/nyfiken/settings"

// Error wrapper.
func main() {
	err := nyfikend()
	if err != nil {
		log.Fatalln("nyfikend:", err)
	}
}

func nyfikend() (err error) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var secondsElapsed float64
	go cli.Listen()
	for ; ; secondsElapsed++ {
		// Retrieve an array of pages from INI file and read settings.
		pages, err := ini.ReadIni(settings.ConfigPath, settings.PagesPath)
		if err != nil {
			return err
		}

		// A channel in which errors are sent from p.Check()
		errChan := make(chan error)

		// The number of checks currently taking place
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

		// For each check that took place, listen if any check returned an error
		go func(ch chan error, nChecks int) {
			for i := 0; i < nChecks; i++ {
				if err := <-ch; err != nil {
					log.Println(err)
				}
			}
		}(errChan, numChecks)

		time.Sleep(1 * time.Second)
	}

	return nil
}
