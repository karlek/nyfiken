// Package cli handles command-line input for nyfikenc/d communication.
package cli

import (
	"encoding/gob"
	"log"
	"net"

	"github.com/karlek/nyfiken/ini"
	"github.com/karlek/nyfiken/page"
	"github.com/karlek/nyfiken/settings"
	"github.com/mewkiz/pkg/bufioutil"
	"github.com/mewkiz/pkg/errutil"
)

// Listen makes nyfikend wait for a connection from nyfikenc.
func Listen() {
	err := errWrapListen()
	if err != nil {
		log.Fatalln(errutil.Err(err))
	}
}

func errWrapListen() (err error) {
	ln, err := net.Listen("tcp", settings.Global.PortNum)
	if err != nil {
		return errutil.Err(err)
	}

	// Wait for request.
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(errutil.Err(err))
			continue
		}

		// Listen for errors from connection.
		errChan := make(chan error, 1)
		go errWrapTakeInput(conn, errChan)
		if err = <-errChan; err != nil {
			log.Println(errutil.Err(err))
		}
		conn.Close()
	}
}

func errWrapTakeInput(conn net.Conn, outerErrChan chan error) {
	outerErrChan <- takeInput(conn)
}

// Wait for input and send output to client.
func takeInput(conn net.Conn) (err error) {
	for {
		query, err := bufioutil.NewReader(conn).ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return errutil.Err(err)
		}

		// Do something with the query
		switch query {
		case settings.QueryUpdates:
			// Encode (send) the value.
			err = gob.NewEncoder(conn).Encode(settings.Updates)
		case settings.QueryClearAll:
			settings.Updates = make(map[string]bool)
			err = settings.SaveUpdates()
		case settings.QueryForceRecheck:
			pages, err := ini.ReadPages(settings.PagesPath)
			if err != nil {
				return errutil.Err(err)
			}
			err = page.ForceUpdate(pages)
		}
		if err != nil {
			return errutil.Err(err)
		}
	}
	return nil
}
