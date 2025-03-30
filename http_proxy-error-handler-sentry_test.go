// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
)

var _ = Describe("SentryProxyErrorHandler", func() {
	var handler libhttp.ProxyErrorHandler
	BeforeEach(func() {
		handler = libhttp.NewSentryProxyErrorHandler(nil)
	})
	It("returns handler", func() {
		Expect(handler).NotTo(BeNil())
	})
})
