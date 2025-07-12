// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
)

var _ = Describe("ContentTypes", func() {
	Context("ApplicationJsonContentType", func() {
		It("returns correct content type", func() {
			Expect(libhttp.ApplicationJsonContentType).To(Equal("application/json"))
		})
	})
	Context("TextHtml", func() {
		It("returns correct content type", func() {
			Expect(libhttp.TextHtml).To(Equal("text/html"))
		})
	})
})
