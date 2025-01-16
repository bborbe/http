// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"context"
	"net/http"

	libhttp "github.com/bborbe/http"
	"github.com/bborbe/http/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RoundTripperRemovePathPrefix", func() {
	var ctx context.Context
	var err error
	var roundTripper *mocks.HttpRoundTripper
	var req *http.Request
	var roundTripperRemovePathPrefix http.RoundTripper
	var resp *http.Response
	var target string
	BeforeEach(func() {
		ctx = context.Background()
		roundTripper = &mocks.HttpRoundTripper{}
		roundTripper.RoundTripReturns(&http.Response{}, nil)
		roundTripperRemovePathPrefix = libhttp.NewRoundTripperRemovePathPrefix(
			roundTripper,
			"/my-prefix",
		)
	})
	JustBeforeEach(func() {
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
		Expect(err).To(BeNil())
		resp, err = roundTripperRemovePathPrefix.RoundTrip(req)
	})
	Context("with prefix", func() {
		BeforeEach(func() {
			target = "http://example.com/my-prefix/index.html"
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("returns resp", func() {
			Expect(resp).NotTo(BeNil())
		})
		It("calls roundtripper", func() {
			Expect(roundTripper.RoundTripCallCount()).To(Equal(1))
			argRequest := roundTripper.RoundTripArgsForCall(0)
			Expect(argRequest).NotTo(BeNil())
			Expect(argRequest.URL.String()).To(Equal("http://example.com/index.html"))
		})
	})
	Context("without prefix", func() {
		BeforeEach(func() {
			target = "http://example.com/index.html"
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("returns resp", func() {
			Expect(resp).NotTo(BeNil())
		})
		It("calls roundtripper", func() {
			Expect(roundTripper.RoundTripCallCount()).To(Equal(1))
			argRequest := roundTripper.RoundTripArgsForCall(0)
			Expect(argRequest).NotTo(BeNil())
			Expect(argRequest.URL.String()).To(Equal("http://example.com/index.html"))
		})
	})
})
