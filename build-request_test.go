// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"context"
	"io"
	"net/http"
	"net/url"

	libhttp "github.com/bborbe/http"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BuildRequest", func() {
	var ctx context.Context
	var err error
	var req *http.Request
	var parameter url.Values
	var urlString string
	var method string
	var body io.Reader
	BeforeEach(func() {
		ctx = context.Background()
		body = nil
		parameter = url.Values{}
		parameter.Set("a", "b")
		urlString = "https://www.example.com/test"
		method = http.MethodGet
	})
	JustBeforeEach(func() {
		req, err = libhttp.BuildRequest(ctx, method, urlString, parameter, body, http.Header{})
	})
	It("returns req", func() {
		Expect(req).NotTo(BeNil())
	})
	It("returns no error", func() {
		Expect(err).To(BeNil())
	})
})
