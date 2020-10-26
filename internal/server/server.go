// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package server

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/v5/mlog"
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
	healthSrv *http.Server
	logger    *mlog.Logger
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

	healthMux := mux.NewRouter()
	healthServer := &http.Server{
		Addr:         cfg.ServiceSettings.HealthHost,
		Handler:      healthMux,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	s := &Server{
		srv:       server,
		healthSrv: healthServer,
		client:    client,
		logger: mlog.NewLogger(&mlog.LoggerConfiguration{
			ConsoleJson:   cfg.LogSettings.ConsoleJSON,
			ConsoleLevel:  strings.ToLower(cfg.LogSettings.ConsoleLevel),
			EnableConsole: cfg.LogSettings.EnableConsole,
			EnableFile:    cfg.LogSettings.EnableFile,
			FileJson:      cfg.LogSettings.FileJSON,
			FileLevel:     strings.ToLower(cfg.LogSettings.FileLevel),
			FileLocation:  cfg.LogSettings.FileLocation,
		}),
		cfg:   cfg,
		creds: credentials.NewStatic(cfg.S3Settings.AccessKeyID, cfg.S3Settings.SecretAccessKey, "", credentials.SignatureV4),
	}

	s.getHostFn = s.getHost
	s.srv.Handler = s.withRecovery(s.handler())
	healthMux.HandleFunc("/health", s.healthHandler).Methods("GET")

	return s
}

// Start starts the server
func (s *Server) Start() error {
	var wg sync.WaitGroup

	errChan := make(chan error, 2)
	wg.Add(1)
	go func() {
		s.logger.Info("server started", mlog.String("host", s.cfg.ServiceSettings.Host))
		var err error
		if s.cfg.ServiceSettings.TLSCertFile != "" && s.cfg.ServiceSettings.TLSKeyFile != "" {
			err = s.srv.ListenAndServeTLS(s.cfg.ServiceSettings.TLSCertFile, s.cfg.ServiceSettings.TLSKeyFile)
		} else {
			err = s.srv.ListenAndServe()
		}
		errChan <- err
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		errChan <- s.healthSrv.ListenAndServe()
		wg.Done()
	}()

	wg.Wait()
	close(errChan)
	for err := range errChan {
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}
	return nil
}

// Stop stops the server
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		return err
	}
	return s.healthSrv.Shutdown(ctx)
}

func (s *Server) withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if x := recover(); x != nil {
				s.logger.Error("recovered from a panic",
					mlog.String("url", r.URL.String()),
					mlog.Any("error", x),
					mlog.String("stack", string(debug.Stack())))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
