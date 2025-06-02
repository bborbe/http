// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"context"
	"net/http"

	"github.com/bborbe/errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
	"github.com/bborbe/http/mocks"
)

var _ = Describe("MetricsRoundTripper", func() {
	var transport *mocks.HttpRoundTripper
	var roundtripper http.RoundTripper
	BeforeEach(func() {
		transport = &mocks.HttpRoundTripper{}
		roundtripper = libhttp.NewMetricsRoundTripper(
			transport,
			libhttp.NewRoundTripperMetrics(),
		)
	})
	It("handles errors", func() {
		transport.RoundTripReturns(nil, errors.New(context.Background(), "banana"))
		req, err := http.NewRequest(http.MethodGet, "http://www.example.com", nil)
		Expect(err).To(BeNil())
		response, err := roundtripper.RoundTrip(req)
		Expect(err).NotTo(BeNil())
		Expect(response).To(BeNil())
	})
	It("handles errors", func() {
		resp := &http.Response{}
		transport.RoundTripReturns(resp, nil)
		req, err := http.NewRequest(http.MethodGet, "http://www.example.com", nil)
		Expect(err).To(BeNil())
		response, err := roundtripper.RoundTrip(req)
		Expect(err).To(BeNil())
		Expect(response).To(Equal(resp))
	})
})
