// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package server

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
)

// Config is the configuration for a bifrost server.
type Config struct {
	ServiceSettings ServiceSettings
	S3Settings      AmazonS3Settings
	LogSettings     LogSettings
}

// ServiceSettings is the configuration related to the web server.
type ServiceSettings struct {
	Host                      string
	ServiceHost               string
	TLSCertFile               string
	TLSKeyFile                string
	MaxConnsPerHost           int
	ResponseHeaderTimeoutSecs int
	ReadTimeoutSecs           int
	WriteTimeoutSecs          int
	IdleTimeoutSecs           int
	RequestValidation         bool
}

// AmazonS3Settings is the configuration related to the Amazon S3.
type AmazonS3Settings struct {
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	Region          string
	Endpoint        string
	Scheme          string
}

// LogSettings is the configuration for the logger.
type LogSettings struct {
	EnableConsole bool
	ConsoleLevel  string
	ConsoleJSON   bool `json:"ConsoleJson"`
	EnableFile    bool
	FileLevel     string
	FileJSON      bool `json:"FileJson"`
	FileLocation  string
}

// ParseConfig reads the config file and returns a new *Config,
// This method overrides values in the file if there is any environment
// variables corresponding to a specific setting.
func ParseConfig(path string) (Config, error) {
	var cfg Config
	file, err := os.Open(path)
	if err != nil {
		return cfg, fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&cfg)
	if err != nil {
		return cfg, fmt.Errorf("could not decode file: %w", err)
	}

	if err = envconfig.Process("bifrost", &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
