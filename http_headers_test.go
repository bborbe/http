// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
)

var _ = Describe("Headers", func() {
	Context("ContentTypeHeaderName", func() {
		It("returns correct header name", func() {
			Expect(libhttp.ContentTypeHeaderName).To(Equal("Content-Type"))
		})
	})
})
