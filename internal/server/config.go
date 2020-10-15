// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package server

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
)

// Config is the configration for a bifrost server.
type Config struct {
	ServiceSettings ServiceSettings  `split_words:"true"`
	S3Settings      AmazonS3Settings `split_words:"true"`
}

// ServiceSettings is the configuration related to the web server.
type ServiceSettings struct {
	Port                  string
	TLSCertFile           string `split_words:"true"`
	TLSKeyFile            string `split_words:"true"`
	MaxConnsPerHost       int    `split_words:"true"`
	ResponseHeaderTimeout int    `split_words:"true"`
}

// AmazonS3Settings is the configuration related to the Amazon S3.
type AmazonS3Settings struct {
	AccessKeyID     string `split_words:"true"`
	SecretAccessKey string `split_words:"true"`
	Bucket          string
	Region          string
	Endpoint        string
	Scheme          string
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
