// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"net/http"
	"time"

	libtime "github.com/bborbe/time"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
)

var _ = Describe("RoundTripperLog", func() {
	var originalNow func() time.Time
	var nowCalls int
	var roundTripper http.RoundTripper
	var resp *http.Response
	var err error

	// libtime.Now is a package-level var (time_now.go: `var Now = time.Now`),
	// so it can be swapped for a fake clock. Counting the calls is what pins the
	// fix: deriving the duration from libtime calls it twice per request (once
	// for the start, once inside .Sub), whereas the previous
	// time.Since(now) form called it only once and took the delta from the
	// stdlib wall clock. A real-time test cannot tell those apart, which is why
	// the mixed-clock bug survived.
	BeforeEach(func() {
		originalNow = libtime.Now
		nowCalls = 0
		libtime.Now = func() time.Time {
			nowCalls++
			return time.Date(2026, 8, 12, 10, 0, 0, 0, time.UTC)
		}
	})

	AfterEach(func() {
		libtime.Now = originalNow
	})

	Context("successful request", func() {
		BeforeEach(func() {
			roundTripper = libhttp.NewRoundTripperLog(
				libhttp.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusOK}, nil
				}),
			)
			req, _ := http.NewRequest(http.MethodGet, "http://example.com/path", nil)
			resp, err = roundTripper.RoundTrip(req)
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("returns the response", func() {
			Expect(resp).NotTo(BeNil())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
		It("derives the duration from libtime, not the wall clock", func() {
			Expect(nowCalls).To(Equal(2))
		})
	})

	Context("failing request", func() {
		BeforeEach(func() {
			roundTripper = libhttp.NewRoundTripperLog(
				libhttp.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
					return nil, http.ErrHandlerTimeout
				}),
			)
			req, _ := http.NewRequest(http.MethodGet, "http://example.com/path", nil)
			resp, err = roundTripper.RoundTrip(req)
		})
		It("returns the error", func() {
			Expect(err).NotTo(BeNil())
		})
		It("derives the duration from libtime on the error path too", func() {
			Expect(nowCalls).To(Equal(2))
		})
	})
})
