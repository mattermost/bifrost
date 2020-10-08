// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Version information, assigned by ldflags
var (
	CommitHash string
	BuildDate  string
)

// Server is the wrapper of http.Server
type Server struct {
	server *http.Server
}

// New creates a new Bifrost server
func New(port int) *Server {
	r := mux.NewRouter()

	s := &Server{
		server: &http.Server{
			Addr:         fmt.Sprintf("0.0.0.0:%d", port),
			Handler:      r,
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 60 * time.Second,
			IdleTimeout:  30 * time.Second,
		},
	}

	return s
}

// Start starts the server
func (s *Server) Start() error {
	err := s.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// Stop stops the server
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}
