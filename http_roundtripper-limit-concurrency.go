// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"

	"github.com/bborbe/errors"
	"github.com/golang/glog"
)

// NewRoundTripperLimitConcurrency wraps a RoundTripper with a bound on how many
// requests may be in flight at once. A request that cannot get a slot waits for
// one, and returns the context error if the context ends first.
//
// The bound is host-agnostic: the caller scopes the limiter by wrapping only the
// client whose requests should be bounded. A maxConcurrent below 1 is raised to
// 1, because an unbuffered slot channel would deadlock. A nil logf falls back to
// glog.Warningf. The effective bound is logged once at construction.
func NewRoundTripperLimitConcurrency(
	roundTripper http.RoundTripper,
	maxConcurrent int,
	logf func(format string, args ...any),
) http.RoundTripper {
	if maxConcurrent < 1 {
		maxConcurrent = 1
	}
	if logf == nil {
		logf = glog.Warningf
	}
	logf("inflight_bound=%d", maxConcurrent)

	slots := make(chan struct{}, maxConcurrent)
	return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		ctx := req.Context()
		select {
		case slots <- struct{}{}:
		case <-ctx.Done():
			return nil, errors.Wrap(ctx, ctx.Err(), "acquire inflight slot")
		}
		defer func() {
			<-slots
		}()

		resp, err := roundTripper.RoundTrip(req)
		if err != nil {
			return nil, errors.Wrap(ctx, err, "roundtrip failed")
		}
		return resp, nil
	})
}
