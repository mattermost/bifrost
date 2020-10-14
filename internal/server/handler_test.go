// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// regCred matches credential string in HTTP header
var regCred = regexp.MustCompile("Credential=([A-Za-z0-9]+)/([0-9]+)/")

// regCred matches signature string in HTTP header
var regSign = regexp.MustCompile("Signature=([[0-9a-f]+)")

func TestHandler(t *testing.T) {
	cfg := Config{
		S3Settings: AmazonS3Settings{
			AccessKeyID:     "AKIA2AccessKey",
			SecretAccessKey: "start/secretkey/end",
			Region:          "us-east-1",
			Endpoint:        "s3.dualstack.us-east-1.amazonaws.com",
			Scheme:          "http",
			Bucket:          "agnivatest",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We test that the bucket name is stripped.
		assert.Equal(t, "/foo", r.URL.Path)

		now := time.Now()

		// Validate request headers
		authHeader := r.Header.Get("Authorization")
		date := r.Header.Get("X-Amz-Date")

		assert.True(t, strings.HasPrefix(authHeader, "AWS4-HMAC-SHA256"), "unexpected prefix for Authorization header: %s", authHeader)
		assert.True(t, strings.HasPrefix(date, now.Format("20060102")), "unexpected prefix for X-Amz-Date header: %s", date)

		matches := regCred.FindStringSubmatch(authHeader)
		require.Len(t, matches, 3, "unexpected number of matches")
		assert.Equal(t, cfg.S3Settings.AccessKeyID, matches[1], "unexpected access key")
		assert.Equal(t, now.Format("20060102"), matches[2], "unexpected date value")

		matches = regSign.FindStringSubmatch(authHeader)
		require.Len(t, matches, 2, "unexpected number of matches")

		w.Header().Set("Content-Type", "application/xml")
		w.Header().Set("Date", now.Format(time.RFC1123))
		w.Header().Set("Last-Modified", now.Format(time.RFC1123))
		w.Header().Set("Server", "Asgard")
		w.Header().Set("X-Amz-Bucket-Region", cfg.S3Settings.Region)
		w.Header().Set("X-Amz-Id-2", "id")
		w.Header().Set("X-Amz-Request-Id", "reqId")

		fmt.Fprintln(w, "Welcome to the realm eternal")
	}))
	defer ts.Close()

	dummyGetHost := func(bucket, endPoint string) string {
		return strings.TrimPrefix(ts.URL, "http://")
	}

	s := &Server{
		logger:    log.New(os.Stderr, "", log.Lshortfile|log.LstdFlags),
		cfg:       cfg,
		getHostFn: dummyGetHost,
		client:    http.DefaultClient,
		creds: credentials.NewStatic(cfg.S3Settings.AccessKeyID,
			cfg.S3Settings.SecretAccessKey, "", credentials.SignatureV4),
	}

	req := httptest.NewRequest("GET", "http://example.com/"+cfg.S3Settings.Bucket+"/foo", nil)
	w := httptest.NewRecorder()

	s.handler()(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	io.Copy(ioutil.Discard, resp.Body)

	// Verify response headers
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status code")
	assert.Equal(t, "application/xml", resp.Header.Get("Content-Type"), "unexpected content type")
	assert.Equal(t, "Asgard", resp.Header.Get("Server"), "unexpected server")
	assert.Equal(t, cfg.S3Settings.Region, resp.Header.Get("X-Amz-Bucket-Region"), "unexpected region")
	assert.Equal(t, "id", resp.Header.Get("X-Amz-Id-2"), "unexpected id")
	assert.Equal(t, "reqId", resp.Header.Get("X-Amz-Request-Id"), "unexpected request id")
	assert.NotEmpty(t, resp.Header.Get("Date"), "empty date")
	assert.NotEmpty(t, resp.Header.Get("Last-Modified"), "empty last-modified")
}
