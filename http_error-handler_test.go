// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
	"github.com/bborbe/http/mocks"
)

var _ = Describe("ErrorHandler", func() {
	var req *http.Request
	var subhandler *mocks.HttpWithError
	var resp *httptest.ResponseRecorder
	BeforeEach(func() {
		req = httptest.NewRequest("GET", "http://example.com", nil)
		resp = httptest.NewRecorder()
		subhandler = &mocks.HttpWithError{}
	})
	JustBeforeEach(func() {
		handler := libhttp.NewErrorHandler(subhandler)
		handler.ServeHTTP(resp, req)
	})
	Context("success", func() {
		BeforeEach(func() {
			subhandler.ServeHTTPReturns(nil)
		})
		It("call subhandler", func() {
			Expect(subhandler.ServeHTTPCallCount()).To(Equal(1))
		})
		It("call subhandler", func() {
			Expect(resp.Result().StatusCode).To(Equal(http.StatusOK))
		})
	})
	Context("failed", func() {
		BeforeEach(func() {
			subhandler.ServeHTTPReturns(
				errors.New("banana"),
			)
		})
		It("call subhandler", func() {
			Expect(subhandler.ServeHTTPCallCount()).To(Equal(1))
		})
		It("call subhandler", func() {
			Expect(resp.Result().StatusCode).To(Equal(http.StatusInternalServerError))
		})
		It("call subhandler", func() {
			body, _ := io.ReadAll(resp.Result().Body)
			Expect(string(body)).To(Equal("request failed: banana\n"))
		})
	})
	Context("failed with ErrorWithStatusCode", func() {
		BeforeEach(func() {
			subhandler.ServeHTTPReturns(
				libhttp.WrapWithStatusCode(
					errors.New("banana"),
					http.StatusNotFound,
				),
			)
		})
		It("call subhandler", func() {
			Expect(subhandler.ServeHTTPCallCount()).To(Equal(1))
		})
		It("call subhandler", func() {
			Expect(resp.Result().StatusCode).To(Equal(http.StatusNotFound))
		})
		It("call subhandler", func() {
			body, _ := io.ReadAll(resp.Result().Body)
			Expect(string(body)).To(Equal("request failed: banana\n"))
		})
	})
})
