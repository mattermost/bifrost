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
	var configFile string
	flag.StringVar(&configFile, "config", "config/config.json", "Configuration file for the Bifrost service.")
	flag.Parse()

	config, err := server.ParseConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not parse config file: %s\n", err)
		os.Exit(1)
	}

	s := server.New(config)
	go func() {
		fmt.Printf("listening on port %s\n", config.ServiceSettings.Port)
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
