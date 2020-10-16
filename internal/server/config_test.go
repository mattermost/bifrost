// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package server

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	dir, err := ioutil.TempDir("", "bifrost-test")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	t.Run("empty input should fail", func(t *testing.T) {
		_, err = ParseConfig("")
		require.Error(t, err)
	})

	t.Run("non existing file should fail", func(t *testing.T) {
		_, err = ParseConfig(filepath.Join(dir, "nonexistent.json"))
		require.Error(t, err)
	})

	t.Run("invalid json format should fail", func(t *testing.T) {
		f := filepath.Join(dir, "empty.json")
		err = ioutil.WriteFile(f, []byte("\n"), 0644)
		_, err = ParseConfig(f)
		require.Error(t, err)
	})

	t.Run("should be able to read a valid json file", func(t *testing.T) {
		f := filepath.Join(dir, "valid.json")
		err = ioutil.WriteFile(f, []byte("{}\n"), 0644)
		_, err = ParseConfig(f)
		require.NoError(t, err)
	})

	t.Run("should be able to read a valid json file and override with env. var.", func(t *testing.T) {
		f := filepath.Join(dir, "valid.json")
		err = ioutil.WriteFile(f, []byte(`{
			"ServiceSettings": {
				"Host": "localhost:8087",
				"TLSCertFile": ""
			}
		}
		`), 0644)
		os.Setenv("BIFROST_SERVICE_SETTINGS_HOST", "localhost:8099")
		os.Setenv("BIFROST_SERVICE_SETTINGS_TLS_CERT_FILE", "/home/test/file.cert")
		cfg, err := ParseConfig(f)
		require.NoError(t, err)
		require.Equal(t, cfg.ServiceSettings.Host, "localhost:8099")
		require.Equal(t, cfg.ServiceSettings.TLSCertFile, "/home/test/file.cert")
	})
}
