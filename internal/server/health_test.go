// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler(t *testing.T) {
	s := &Server{
		logger: mlog.NewTestingLogger(t, os.Stderr),
	}

	ts := httptest.NewServer(http.HandlerFunc(s.healthHandler))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&healthResponse{})
	require.NoError(t, err)
}
