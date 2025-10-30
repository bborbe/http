// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
	"time"
)

// CreateDefaultHTTPClient creates an HTTP client with default configuration.
// It uses a 30-second timeout and the default RoundTripper with retry logic and logging.
// The client disables automatic redirects by returning ErrUseLastResponse.
func CreateDefaultHTTPClient() *http.Client {
	return CreateHTTPClient(30 * time.Second)
}

// CreateDefaultHttpClient is deprecated. Use CreateDefaultHTTPClient instead.
//
// Deprecated: Use CreateDefaultHTTPClient for correct Go naming conventions.
//
//nolint:revive
func CreateDefaultHttpClient() *http.Client {
	return CreateDefaultHTTPClient()
}

// CreateHTTPClient creates an HTTP client with the specified timeout.
// It uses the default RoundTripper with retry logic and logging, and disables automatic redirects.
// The timeout applies to the entire request including connection, redirects, and reading the response.
func CreateHTTPClient(
	timeout time.Duration,
) *http.Client {
	return &http.Client{
		Transport: CreateDefaultRoundTripper(),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: timeout,
	}
}

// CreateHttpClient is deprecated. Use CreateHTTPClient instead.
//
// Deprecated: Use CreateHTTPClient for correct Go naming conventions.
//
//nolint:revive
func CreateHttpClient(timeout time.Duration) *http.Client {
	return CreateHTTPClient(timeout)
}
