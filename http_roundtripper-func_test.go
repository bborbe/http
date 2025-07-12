// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
)

var _ = Describe("RoundTripperFunc", func() {
	Context("RoundTrip", func() {
		It("calls the function", func() {
			var called bool
			var capturedReq *http.Request

			roundTripper := libhttp.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
				called = true
				capturedReq = req
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       http.NoBody,
				}, nil
			})

			req := httptest.NewRequest("GET", "http://example.com", nil)

			resp, err := roundTripper.RoundTrip(req)

			Expect(err).To(BeNil())
			Expect(called).To(BeTrue())
			Expect(capturedReq).To(Equal(req))
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})
})
