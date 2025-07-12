// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
)

var _ = Describe("WriteAndGlog", func() {
	Context("WriteAndGlog", func() {
		It("writes formatted string to writer", func() {
			var buffer bytes.Buffer

			n, err := libhttp.WriteAndGlog(&buffer, "hello %s", "world")

			Expect(err).To(BeNil())
			Expect(n).To(Equal(12)) // "hello world\n"
			Expect(buffer.String()).To(Equal("hello world\n"))
		})

		It("handles no arguments", func() {
			var buffer bytes.Buffer

			n, err := libhttp.WriteAndGlog(&buffer, "test")

			Expect(err).To(BeNil())
			Expect(n).To(Equal(5)) // "test\n"
			Expect(buffer.String()).To(Equal("test\n"))
		})
	})
})
