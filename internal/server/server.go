// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package server

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Version information, assigned by ldflags
var (
	CommitHash string
	BuildDate  string
)

// Server contains all the necessary information to run Bifrost
type Server struct {
	cfg       Config
	srv       *http.Server
	logger    *log.Logger
	client    *http.Client
	getHostFn func(bucket, endPoint string) string
	creds     *credentials.Credentials
}

// New creates a new Bifrost server
func New(cfg Config) *Server {
	// All settings are same as DefaultTransport,
	// with MaxConnsPerHost and ResponseHeaderTimeout added.
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxConnsPerHost:       cfg.ServiceSettings.MaxConnsPerHost,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	server := &http.Server{
		Addr:         cfg.ServiceSettings.Host,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  30 * time.Second,
		TLSConfig: &tls.Config{
			MinVersion:               tls.VersionTLS12,
			PreferServerCipherSuites: true,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
			},
		},
	}

	s := &Server{
		srv:    server,
		client: client,
		logger: log.New(os.Stderr, "[bifrost] ", log.Lshortfile|log.LstdFlags),
		cfg:    cfg,
		creds:  credentials.NewStatic(cfg.S3Settings.AccessKeyID, cfg.S3Settings.SecretAccessKey, "", credentials.SignatureV4),
	}

	s.getHostFn = s.getHost
	s.srv.Handler = s.handler()

	return s
}

// Start starts the server
func (s *Server) Start() error {
	s.logger.Println("Listening on ", s.cfg.ServiceSettings.Host)
	var err error
	if s.cfg.ServiceSettings.TLSCertFile != "" && s.cfg.ServiceSettings.TLSKeyFile != "" {
		err = s.srv.ListenAndServeTLS(s.cfg.ServiceSettings.TLSCertFile, s.cfg.ServiceSettings.TLSKeyFile)
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
