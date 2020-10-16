// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package server

import (
	"bytes"
	"encoding/xml"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/signer"
)

func (s *Server) handler() http.HandlerFunc {
	// We need a separate function to compute the host so that we can override
	// it during testing. And we don't want to have a variable for this because
	// later we can have a dynamic mapping between requests and buckets, in which
	// case we need to compute the host for every request.
	host := s.getHostFn(s.cfg.S3Settings.Bucket, s.cfg.S3Settings.Endpoint)

	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Scheme = s.cfg.S3Settings.Scheme
		r.URL.Host = host
		r.Host = host
		// Wiping out RequestURI
		r.RequestURI = ""
		// Stripping the bucket name from the path which gets added by Minio
		// if the S3 hostname does not match a URL pattern.
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/"+s.cfg.S3Settings.Bucket)
		s.logger.Println(r.Method, r.URL.String())

		// Get credentials.
		val, err := s.creds.Get()
		if err != nil {
			s.writeError(w, err)
			return
		}
		// TODO: do streaming sign for PUT requests
		// Need to sign the header, just before sending it
		r = signer.SignV4(*r, s.cfg.S3Settings.AccessKeyID,
			s.cfg.S3Settings.SecretAccessKey,
			val.SessionToken,
			s.cfg.S3Settings.Region)

		resp, err := s.client.Do(r)
		if err != nil {
			s.writeError(w, err)
			return
		}
		defer resp.Body.Close()

		// We copy over the response headers
		for _, h := range []string{"Content-Type", "Date", "Etag", "Last-Modified", "Server",
			"X-Amz-Bucket-Region", "X-Amz-Id-2", "X-Amz-Request-Id"} {
			w.Header().Set(h, resp.Header.Get(h))
		}

		_, err = io.Copy(w, resp.Body)
		if err != nil {
			s.logger.Println(err)
		}
	}
}

func (s *Server) getHost(bucket, endPoint string) string {
	return bucket + "." + endPoint
}

func (s *Server) writeError(w http.ResponseWriter, sourceErr error) {
	s.logger.Println(sourceErr)

	resp := minio.ErrorResponse{
		Code:       strconv.Itoa(http.StatusInternalServerError),
		Message:    sourceErr.Error(),
		BucketName: s.cfg.S3Settings.Bucket,
	}
	// We write an XML response back to the client to match what AWS would return.
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	err := xml.NewEncoder(&buf).Encode(resp)
	if err != nil {
		s.logger.Println(err)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	_, err = w.Write(buf.Bytes())
	if err != nil {
		s.logger.Println(err)
	}
}