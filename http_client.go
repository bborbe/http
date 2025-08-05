// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
	"time"
)

// CreateDefaultHttpClient creates an HTTP client with default configuration.
// It uses a 30-second timeout and the default RoundTripper with retry logic and logging.
// The client disables automatic redirects by returning ErrUseLastResponse.
func CreateDefaultHttpClient() *http.Client {
	return CreateHttpClient(30 * time.Second)
}

// CreateHttpClient creates an HTTP client with the specified timeout.
// It uses the default RoundTripper with retry logic and logging, and disables automatic redirects.
// The timeout applies to the entire request including connection, redirects, and reading the response.
func CreateHttpClient(
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
