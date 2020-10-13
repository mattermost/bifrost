// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

// Version information, assigned by ldflags
var (
	CommitHash string
	BuildDate  string
)

// Server contains all the necessary information to run Bifrost
type Server struct {
	srv *http.Server
}

// New creates a new Bifrost server
func New(port string) *Server {
	s := &Server{
		srv: &http.Server{
			Addr:         net.JoinHostPort("0.0.0.0", port),
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 60 * time.Second,
			IdleTimeout:  30 * time.Second,
		},
	}

	return s
}

// Start starts the server
func (s *Server) Start() error {
	err := s.srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// Stop stops the server
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.srv.Shutdown(ctx)
}
