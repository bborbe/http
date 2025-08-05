// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
)

// NewRoundTripperBasicAuth wraps a RoundTripper with HTTP Basic Authentication.
// It automatically adds Basic Auth headers to all requests using the provided username and password.
// If either username or password is empty, no authentication header is added.
func NewRoundTripperBasicAuth(
	roundTripper RoundTripper,
	username string,
	password string,
) http.RoundTripper {
	return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if username != "" && password != "" {
			req.SetBasicAuth(username, password)
		}
		return roundTripper.RoundTrip(req)
	})
}
