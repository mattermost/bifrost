// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package server

import (
	"context"
	"crypto/tls"
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
	certFile string
	keyFile  string
	srv      *http.Server
}

// New creates a new Bifrost server
func New(cfg *Config) *Server {
	server := &http.Server{
		Addr:         net.JoinHostPort("0.0.0.0", cfg.ServiceSettings.Port),
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	server.TLSConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		},
	}

	s := &Server{
		certFile: cfg.ServiceSettings.TLSCertFile,
		keyFile:  cfg.ServiceSettings.TLSKeyFile,
		srv:      server,
	}
	return s
}

// Start starts the server
func (s *Server) Start() error {
	var err error
	if s.certFile != "" && s.keyFile != "" {
		err = s.srv.ListenAndServeTLS(s.certFile, s.keyFile)
	} else {
		err = s.srv.ListenAndServe()
	}

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
