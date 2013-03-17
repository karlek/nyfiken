// Package cli handles command-line input for nyfikenc/d communication.
package cli

import "encoding/gob"
import "net"
import "log"

import "github.com/karlek/nyfiken/settings"
import "github.com/karlek/nyfiken/ini"
import "github.com/mewkiz/pkg/errorsutil"
import "github.com/mewkiz/pkg/bufioutil"

// Listen makes the server wait for a connection from a client.
func Listen() {
	ln, err := net.Listen("tcp", settings.Global.PortNum)
	if err != nil {
		log.Fatalln(errorsutil.ErrorfColor("%s", err))
	}

	// Wait for request.
	for {
		conn, err := ln.Accept()
		if err != nil {
			if err != nil {
				log.Fatalln(errorsutil.ErrorfColor("%s", err))
			}
			continue
		}

		// Listen for errors from connection.
		errChan := make(chan error)
		go takeInput(conn, errChan)
		if err := <-errChan; err != nil {
			log.Fatalln(errorsutil.ErrorfColor("%s", err))
		}
	}
}

// Wait for input and send output to client.
func takeInput(conn net.Conn, outerErrChan chan error) {

	ch := make(chan string)
	innerErrChan := make(chan error)

	// Start a goroutine to read from our net connection
	go func(ch chan string, innerErrChan chan error) {
		r := bufioutil.NewReader(conn)
		for {
			data, err := r.ReadLine()
			if err != nil {
				// send an error if it's encountered
				innerErrChan <- err
				return
			}

			// send data if we read some.
			ch <- data
		}
	}(ch, innerErrChan)

	// continuously read from the connection
	for {
		select {
		// This case means we recieved data on the connection
		case query := <-ch:
			// Do something with the data
			switch query {
			case settings.QueryUpdates:
				// Will write to network.
				enc := gob.NewEncoder(conn)

				// Encode (send) the value.
				err := enc.Encode(settings.Updates)
				if err != nil {
					outerErrChan <- err
				}
			case settings.QueryClearAll:
				settings.Updates = make(map[settings.Update]bool)
			case settings.QueryForceRecheck:
				err := forceUpdate()
				if err != nil {
					outerErrChan <- err
				}
			}

		// This case means we got an error and the goroutine has finished
		case err := <-innerErrChan:
			if err.Error() != "EOF" {
				outerErrChan <- err
			}
		}
		outerErrChan <- nil
	}
	outerErrChan <- nil
}

// Check all pages immediately
func forceUpdate() (err error) {
	pages, err := ini.ReadPages(settings.PagesPath)
	if err != nil {
		return err
	}

	// A channel in which errors are sent from p.Check()
	errChan := make(chan error)

	// The number of checks currently taking place
	var numChecks int
	for _, p := range pages {
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

	return nil
}
