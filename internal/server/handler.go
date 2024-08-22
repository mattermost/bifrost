// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package server

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/s3utils"
	"github.com/minio/minio-go/v7/pkg/signer"
	"github.com/pkg/errors"
)

func (s *Server) handler() http.HandlerFunc {
	// We need a separate function to compute the host so that we can override
	// it during testing. And we don't want to have a variable for this because
	// later we can have a dynamic mapping between requests and buckets, in which
	// case we need to compute the host for every request.
	host := s.getHostFn(s.cfg.S3Settings.Bucket, s.cfg.S3Settings.Endpoint)

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		var installationID string
		var elapsed float64
		statusCode := -1
		defer func() {
			s.metrics.observeRequest(r.Method, installationID, statusCode, elapsed)
		}()

		if s := strings.Split(r.URL.Path, "/"); len(s) > 2 {
			installationID = s[2]
		}

		if s.cfg.ServiceSettings.RequestValidation {
			if err := s.validateRequestMatchesInstallationID(r, installationID); err != nil {
				s.writeError(w, errors.Wrap(err, "installation ID request validation failed"))
				return
			}
		}

		// Strip the bucket name from the path which gets added by Minio
		// if the S3 hostname does not match a URL pattern.
		objectName := strings.TrimPrefix(r.URL.Path, "/"+s.cfg.S3Settings.Bucket)

		// Rebuild the URL from scratch, using s3utils.EncodePath on the unescaped
		// objectName from the path.
		//
		// When Mattermost makes an S3 request, the minio library already calls
		// s3utils.EncodePath, translating (among other characters) both ' ' to %20 and
		// '+' to %2B. This is actually more strict than RFC 3986 requires, since a
		// '+' in the path doesn't actually need to be escaped.
		//
		// When Bifrost receives the request, net.URL happily unescapes both characters.
		// The signer package generates a canonical URL before signing, also using
		// s3utils.EncodePath. But if we don't re-encode with s3utils.EncodePath ourselves,
		// then our replayed request upstream will only encode the ' ' and not the '+',
		// resulting in a signature mismatch.
		//
		// We have to do it here, within Bifrost, and not from Mattermost, otherwise we're
		// effectively just double escaping. While that works to avoid the signature
		// mismatch, it changes the lookup paths for previously created files.
		urlStr := s.cfg.S3Settings.Scheme + "://" + host + s3utils.EncodePath(objectName)
		if len(r.URL.RawQuery) > 0 {
			urlStr += "?" + r.URL.RawQuery
		}

		targetURL, err := url.Parse(urlStr)
		if err != nil {
			s.writeError(w, err)
			return
		}

		originalURL := r.URL
		r.URL = targetURL
		r.Host = host
		// Wiping out RequestURI
		r.RequestURI = ""

		s.logger.Debug("received request", mlog.String("method", r.Method), mlog.String("url", originalURL.String()), mlog.String("target_url", targetURL.String()))

		// Get credentials.
		val, err := s.creds.Get()
		if err != nil {
			s.writeError(w, err)
			return
		}

		if s.isUsingIAMRoleCredentials() {
			s.cfg.S3Settings.AccessKeyID = val.AccessKeyID
			s.cfg.S3Settings.SecretAccessKey = val.SecretAccessKey
		}

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
		statusCode = resp.StatusCode

		// We copy over the response headers
		for key, value := range resp.Header {
			w.Header().Set(key, strings.Join(value, ", "))
		}

		w.WriteHeader(resp.StatusCode)

		_, err = io.Copy(w, resp.Body)
		if err != nil {
			s.logger.Warn("failed to copy response body", mlog.Err(err))
		}
		elapsed = float64(time.Since(start)) / float64(time.Second)
	}
}

func (s *Server) getHost(bucket, endPoint string) string {
	return bucket + "." + endPoint
}

func (s *Server) isUsingIAMRoleCredentials() bool {
	return s.cfg.S3Settings.AccessKeyID == "" && s.cfg.S3Settings.SecretAccessKey == ""
}

func (s *Server) writeError(w http.ResponseWriter, sourceErr error) {
	s.logger.Error("error", mlog.Err(sourceErr))

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
		s.logger.Error("failed to encode error body", mlog.Err(err))
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	_, err = w.Write(buf.Bytes())
	if err != nil {
		s.logger.Warn("failed to write error response", mlog.Err(err))
	}
}

func (s *Server) validateRequestMatchesInstallationID(r *http.Request, installationID string) error {
	addr := strings.Split(r.RemoteAddr, ":")[0]
	names, err := s.lookupAddrFn(addr)
	if err != nil {
		return errors.Wrap(err, "failed to perform reverse domain name lookup")
	}
	if len(names) == 0 {
		return errors.New("no names returned in reverse lookup")
	}
	name := names[0]

	// Perform validation by comparing reverse lookup to expected namespace.
	// Example Reverse Lookup:
	//   IP_ADDR.SERVICE_NAME.NAMESPACE/INSTALLATION_ID.svc.cluster.local.
	if !s.requestIsValid(name, installationID) {
		return errors.Errorf("reverse name lookup validation failed; name=%s, installationID=%s", name, installationID)
	}

	s.logger.Debug("reverse name lookup validation passed", mlog.String("name", name), mlog.String("installationID", installationID))

	return nil
}

func (s *Server) requestIsValid(name, installationID string) bool {
	return strings.HasSuffix(name, fmt.Sprintf(".%s.%s", installationID, s.cfg.ServiceSettings.RequestValidationExpectedNameSuffix))
}
