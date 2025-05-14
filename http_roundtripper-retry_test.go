// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"bytes"
	"io"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
	"github.com/bborbe/http/mocks"
)

var _ = Describe("RoundTripperRetry", func() {
	var err error
	var req *http.Request
	var roundTripperRetry http.RoundTripper
	var resp *http.Response
	var roundTripper *mocks.HttpRoundTripper
	BeforeEach(func() {
		req, err = http.NewRequest(http.MethodGet, "http://example.com", &bytes.Buffer{})
		Expect(err).To(BeNil())

		roundTripper = &mocks.HttpRoundTripper{}

		roundTripperRetry = libhttp.NewRoundTripperRetry(
			roundTripper,
			3,
			time.Nanosecond,
		)
	})
	JustBeforeEach(func() {
		resp, err = roundTripperRetry.RoundTrip(req)
	})
	Context("success", func() {
		BeforeEach(func() {
			roundTripper.RoundTripReturns(&http.Response{StatusCode: 200}, nil)
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("returns resp", func() {
			Expect(resp).NotTo(BeNil())
			Expect(resp.StatusCode).To(Equal(200))
		})
		It("calls roundTrip once", func() {
			Expect(roundTripper.RoundTripCallCount()).To(Equal(1))
		})
	})
	Context("400", func() {
		BeforeEach(func() {
			roundTripper.RoundTripReturns(&http.Response{StatusCode: 200}, nil)
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("returns resp", func() {
			Expect(resp).NotTo(BeNil())
			Expect(resp.StatusCode).To(Equal(200))
		})
		It("calls roundTrip once", func() {
			Expect(roundTripper.RoundTripCallCount()).To(Equal(1))
		})
	})
	Context("500", func() {
		BeforeEach(func() {
			roundTripper.RoundTripReturns(&http.Response{StatusCode: 500}, nil)
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("returns resp", func() {
			Expect(resp).NotTo(BeNil())
			Expect(resp.StatusCode).To(Equal(500))
		})
		It("calls roundTrip once", func() {
			Expect(roundTripper.RoundTripCallCount()).To(Equal(4))
		})
	})
	Context("eof", func() {
		BeforeEach(func() {
			roundTripper.RoundTripReturns(nil, io.EOF)
		})
		It("returns error", func() {
			Expect(err).NotTo(BeNil())
		})
		It("returns no resp", func() {
			Expect(resp).To(BeNil())
		})
		It("calls roundTrip once", func() {
			Expect(roundTripper.RoundTripCallCount()).To(Equal(1))
		})
	})
	Context("500 with recover", func() {
		BeforeEach(func() {
			roundTripper.RoundTripReturnsOnCall(0, &http.Response{StatusCode: 500}, nil)
			roundTripper.RoundTripReturnsOnCall(1, &http.Response{StatusCode: 200}, nil)
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("returns resp", func() {
			Expect(resp).NotTo(BeNil())
			Expect(resp.StatusCode).To(Equal(200))
		})
		It("calls roundTrip once", func() {
			Expect(roundTripper.RoundTripCallCount()).To(Equal(2))
		})
	})
})
