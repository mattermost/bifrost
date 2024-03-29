// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/stretchr/testify/require"
)

type panicHandler struct {
}

func (ph panicHandler) ServeHTTP(_ http.ResponseWriter, _ *http.Request) {
	panic("")
}

func TestWithRecovery(t *testing.T) {
	require.NotPanics(t, func() {
		s := Server{
			logger: mlog.NewTestingLogger(t, os.Stderr),
		}
		ph := panicHandler{}
		handler := s.withRecovery(ph)

		req := httptest.NewRequest("GET", "http://random", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.Body != nil {
			defer resp.Body.Close()
			_, err := io.Copy(io.Discard, resp.Body)
			require.NoError(t, err)
		}
	})
}
