// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"errors"
	"io"
	"net/http"

	libsentry "github.com/bborbe/sentry"
	"github.com/getsentry/sentry-go"
	"github.com/golang/glog"
)

// NewSentryProxyErrorHandler creates a ProxyErrorHandler that reports errors to Sentry.
// It logs errors and sends them to Sentry unless they are ignored errors (like timeouts or retryable errors).
// The handler returns a 502 Bad Gateway status for all proxy errors.
func NewSentryProxyErrorHandler(sentryClient libsentry.Client) ProxyErrorHandler {
	return ProxyErrorHandlerFunc(func(resp http.ResponseWriter, req *http.Request, err error) {
		glog.V(1).
			Infof("handle request to %s for %s failed: %v", req.URL.String(), req.Header.Get("user-agent"), err)
		if IsIgnoredSentryError(err) == false {
			sentryClient.CaptureException(
				err,
				&sentry.EventHint{
					Context: req.Context(),
					Request: req,
					Data: map[string]interface{}{
						"req":  req,
						"resp": resp,
					},
				},
				sentry.NewScope(),
			)
		}
		resp.WriteHeader(http.StatusBadGateway)
	})
}

var sentryIgnoreErrors = []error{
	context.Canceled,
	context.DeadlineExceeded,
	io.EOF,
}

// IsIgnoredSentryError determines whether an error should be ignored when reporting to Sentry.
// It returns true for common transient errors like context cancellation, timeouts, and other retryable errors.
// This helps reduce noise in Sentry by filtering out expected operational errors.
func IsIgnoredSentryError(err error) bool {
	if IsRetryError(err) {
		return true
	}
	for _, ignoredError := range sentryIgnoreErrors {
		if errors.Is(ignoredError, context.Canceled) {
			return true
		}
	}
	return false
}
