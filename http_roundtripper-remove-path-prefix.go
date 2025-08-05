// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
	"strings"
)

// NewRoundTripperRemovePathPrefix wraps a RoundTripper to remove a path prefix from request URLs.
// If the request URL path starts with the specified prefix, it removes that prefix before forwarding the request.
// This is useful for proxying requests where the upstream service expects different path structures.
func NewRoundTripperRemovePathPrefix(
	roundTripper http.RoundTripper,
	prefix string,
) http.RoundTripper {
	return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		path := req.URL.Path
		if strings.HasPrefix(path, prefix) {
			req.URL.Path = path[len(prefix):]
		}
		return roundTripper.RoundTrip(req)
	})
}
