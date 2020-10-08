// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mattermost/bifrost/internal/server"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8087, "Port number for the http server.")
	flag.Parse()

	s := server.New(port)
	go func() {
		fmt.Printf("listening on port %d\n", port)
		if err := s.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "could not start the server: %s\n", err)
			os.Exit(1)
		}
	}()
	defer s.Stop()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-sig
}
