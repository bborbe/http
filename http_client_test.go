// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
)

var _ = Describe("HTTP Client", func() {
	Context("CreateDefaultHTTPClient", func() {
		It("creates client with 30 second timeout", func() {
			client := libhttp.CreateDefaultHTTPClient()
			Expect(client).NotTo(BeNil())
			Expect(client.Timeout).To(Equal(30 * time.Second))
		})

		It("disables redirects", func() {
			client := libhttp.CreateDefaultHTTPClient()
			Expect(client.CheckRedirect).NotTo(BeNil())

			// CheckRedirect should return ErrUseLastResponse
			err := client.CheckRedirect(nil, nil)
			Expect(err).To(Equal(http.ErrUseLastResponse))
		})

		It("has default round tripper", func() {
			client := libhttp.CreateDefaultHTTPClient()
			Expect(client.Transport).NotTo(BeNil())
		})
	})

	Context("CreateHTTPClient", func() {
		It("creates client with custom timeout", func() {
			timeout := 10 * time.Second
			client := libhttp.CreateHTTPClient(timeout)
			Expect(client).NotTo(BeNil())
			Expect(client.Timeout).To(Equal(timeout))
		})

		It("accepts zero timeout", func() {
			client := libhttp.CreateHTTPClient(0)
			Expect(client).NotTo(BeNil())
			Expect(client.Timeout).To(Equal(time.Duration(0)))
		})

		It("accepts very long timeout", func() {
			timeout := 5 * time.Minute
			client := libhttp.CreateHTTPClient(timeout)
			Expect(client.Timeout).To(Equal(timeout))
		})
	})

	Context("Deprecated functions", func() {
		It("CreateDefaultHttpClient works", func() {
			client := libhttp.CreateDefaultHttpClient()
			Expect(client).NotTo(BeNil())
			Expect(client.Timeout).To(Equal(30 * time.Second))
		})

		It("CreateHttpClient works", func() {
			client := libhttp.CreateHttpClient(15 * time.Second)
			Expect(client).NotTo(BeNil())
			Expect(client.Timeout).To(Equal(15 * time.Second))
		})
	})
})
