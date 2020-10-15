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
		_, err := ParseConfig(f)
		require.NoError(t, err)
	})
}
