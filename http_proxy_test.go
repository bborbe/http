// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"

	libhttp "github.com/bborbe/http"
	"github.com/bborbe/http/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Proxy", func() {
	var err error
	var url *url.URL
	var proxy http.Handler
	var roundTripper *mocks.HttpRoundTripper
	var errorHandler *mocks.HttpProxyErrorHandler
	BeforeEach(func() {
		roundTripper = &mocks.HttpRoundTripper{}
		roundTripper.RoundTripReturns(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(&bytes.Buffer{}),
		}, nil)

		url, err = url.Parse("http://proxy.example.com")
		Expect(err).To(BeNil())

		errorHandler = &mocks.HttpProxyErrorHandler{}
		errorHandler.HandleErrorStub = func(resp http.ResponseWriter, req *http.Request, err error) {
			resp.WriteHeader(http.StatusBadGateway)
		}
	})
	Context("ServeHTTP", func() {
		var resp *httptest.ResponseRecorder
		var req *http.Request
		BeforeEach(func() {
			resp = &httptest.ResponseRecorder{}
			req, err = http.NewRequest(http.MethodGet, "http://target.example.com", nil)
			Expect(err).To(BeNil())
		})
		JustBeforeEach(func() {
			proxy = libhttp.NewProxy(
				roundTripper,
				url,
				errorHandler,
			)
			proxy.ServeHTTP(resp, req)
		})
		Context("success", func() {
			It("returns no error", func() {
				Expect(err).To(BeNil())
			})
			It("calls not proxy error", func() {
				Expect(errorHandler.HandleErrorCallCount()).To(Equal(0))
			})
			It("returns statusCode StatusOK", func() {
				Expect(resp.Result().StatusCode).To(Equal(http.StatusOK))
			})
		})
		Context("error", func() {
			BeforeEach(func() {
				roundTripper.RoundTripReturns(nil, errors.New("banana"))
			})
			It("returns no error", func() {
				Expect(err).To(BeNil())
			})
			It("calls proxy error", func() {
				Expect(errorHandler.HandleErrorCallCount()).To(Equal(1))
			})
			It("returns statusCode StatusBadGateway", func() {
				Expect(resp.Result().StatusCode).To(Equal(http.StatusBadGateway))
			})
		})
	})
})
