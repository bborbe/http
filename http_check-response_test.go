// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
)

var _ = Describe("CheckResponse", func() {
	var ctx context.Context
	var req *http.Request
	var resp *http.Response
	var err error

	BeforeEach(func() {
		ctx = context.Background()
		req = httptest.NewRequest("GET", "http://example.com/test", nil)
		req = req.WithContext(ctx)
	})

	Context("RequestFailedError", func() {
		It("returns correct error message", func() {
			err := libhttp.RequestFailedError{
				Method:     "GET",
				URL:        "http://example.com",
				StatusCode: 500,
			}
			Expect(err.Error()).To(Equal("GET request to http://example.com failed with statusCode 500"))
		})
	})

	Context("CheckResponseIsSuccessful", func() {
		JustBeforeEach(func() {
			err = libhttp.CheckResponseIsSuccessful(req, resp)
		})

		Context("200 status", func() {
			BeforeEach(func() {
				resp = &http.Response{
					StatusCode: 200,
					Status:     "200 OK",
					Body:       io.NopCloser(bytes.NewReader([]byte("success"))),
				}
			})
			It("returns no error", func() {
				Expect(err).To(BeNil())
			})
		})

		Context("301 status", func() {
			BeforeEach(func() {
				resp = &http.Response{
					StatusCode: 301,
					Status:     "301 Moved Permanently",
					Body:       io.NopCloser(bytes.NewReader([]byte("moved"))),
				}
			})
			It("returns no error", func() {
				Expect(err).To(BeNil())
			})
		})

		Context("404 status", func() {
			BeforeEach(func() {
				resp = &http.Response{
					StatusCode: 404,
					Status:     "404 Not Found",
					Body:       io.NopCloser(bytes.NewReader([]byte("not found"))),
				}
			})
			It("returns NotFound error", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("not found"))
			})
		})

		Context("500 status", func() {
			BeforeEach(func() {
				resp = &http.Response{
					StatusCode: 500,
					Status:     "500 Internal Server Error",
					Body:       io.NopCloser(bytes.NewReader([]byte("server error"))),
				}
			})
			It("returns RequestFailedError", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("request failed"))
				Expect(err.Error()).To(ContainSubstring("server error"))
			})
		})
	})
})
