// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import "net/http"

// RoundTripperFunc is an adapter to allow the use of ordinary functions as http.RoundTripper.
// If f is a function with the appropriate signature, RoundTripperFunc(f) is a RoundTripper that calls f.
type RoundTripperFunc func(req *http.Request) (*http.Response, error)

// RoundTrip calls f(req).
func (r RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return r(req)
}
