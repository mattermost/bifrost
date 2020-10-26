// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package server

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-server/v5/mlog"
)

type healthResponse struct {
	CommitHash string `json:"commit_hash"`
	BuildDate  string `json:"build_date"`
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(healthResponse{CommitHash: CommitHash, BuildDate: BuildDate})
	if err != nil {
		s.logger.Error("failed to write health response", mlog.Err(err))
		http.Error(w, "failed to write health response", http.StatusInternalServerError)
		return
	}
}
