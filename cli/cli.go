// Package cli handles command-line input for nyfikenc/d communication.
package cli

import (
	"encoding/gob"
	"log"
	"net"

	"github.com/karlek/nyfiken/ini"
	"github.com/karlek/nyfiken/settings"
	"github.com/mewkiz/pkg/bufioutil"
	"github.com/mewkiz/pkg/errorsutil"
)

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
		}

		// Listen for errors from connection.
		errChan := make(chan error, 1)
		go errWrapTakeInput(conn, errChan)
		if err = <-errChan; err != nil {
			log.Fatalln(errorsutil.ErrorfColor("%s", err))
		}
		conn.Close()
	}
}

func errWrapTakeInput(conn net.Conn, outerErrChan chan error) {
	outerErrChan <- takeInput(conn)
}

// Wait for input and send output to client.
func takeInput(conn net.Conn) (err error) {
	query, err := bufioutil.NewReader(conn).ReadLine()
	if err != nil {
		if err.Error() != "EOF" {
			return err
		}
	}

	// Do something with the query
	switch query {
	case settings.QueryUpdates:
		// Will write to network.
		enc := gob.NewEncoder(conn)

		// Encode (send) the value.
		err = enc.Encode(settings.Updates)
	case settings.QueryClearAll:
		settings.Updates = make(map[settings.Update]bool)
	case settings.QueryForceRecheck:
		err = forceUpdate()
	}
	if err != nil {
		return err
	}

	return nil
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
