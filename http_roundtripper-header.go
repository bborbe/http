// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import "net/http"

// NewRoundTripperHeader wraps a RoundTripper to add custom headers to all requests.
// The provided headers are added to every request, replacing any existing headers with the same keys.
// This is useful for adding authentication headers, API keys, or other standard headers.
func NewRoundTripperHeader(
	roundTripper http.RoundTripper,
	header http.Header,
) http.RoundTripper {
	return &roundTripperHeader{
		roundTripper: roundTripper,
		header:       header,
	}
}

type roundTripperHeader struct {
	roundTripper http.RoundTripper
	header       http.Header
}

// RoundTrip implements http.RoundTripper by adding the configured headers to the request.
func (a *roundTripperHeader) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, values := range a.header {
		req.Header.Del(key)
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	return a.roundTripper.RoundTrip(req)
}
