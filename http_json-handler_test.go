// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"context"
	stderrors "errors"
	"net/http"
	"net/http/httptest"

	libhttp "github.com/bborbe/http"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JsonHandler", func() {
	var ctx context.Context
	var err error
	var jsonHandler libhttp.JsonHandler
	var req *http.Request
	var resp *httptest.ResponseRecorder
	BeforeEach(func() {
		ctx = context.Background()

		req = &http.Request{}
	})
	Context("ServeHTTP", func() {
		JustBeforeEach(func() {
			resp = httptest.NewRecorder()
			err = libhttp.NewJsonHandler(jsonHandler).ServeHTTP(ctx, resp, req)
		})
		Context("success", func() {
			BeforeEach(func() {
				jsonHandler = libhttp.JsonHandlerFunc(func(ctx context.Context, req *http.Request) (interface{}, error) {
					return map[string]interface{}{
						"hello": "world",
					}, nil
				})
			})
			It("returns no error", func() {
				Expect(err).To(BeNil())
			})
			It("returns bodyr", func() {
				Expect(resp).NotTo(BeNil())
				Expect(resp.Body).NotTo(BeNil())
				Expect(resp.Body.String()).To(Equal("{\"hello\":\"world\"}\n"))
			})
		})
		Context("failure", func() {
			BeforeEach(func() {
				jsonHandler = libhttp.JsonHandlerFunc(func(ctx context.Context, req *http.Request) (interface{}, error) {
					return nil, stderrors.New("banana")
				})
			})
			It("returns error", func() {
				Expect(err).NotTo(BeNil())
			})
		})
	})
})
