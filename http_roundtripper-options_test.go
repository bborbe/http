// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
)

var _ = Describe("RoundTripper Options", func() {
	Context("CreateRoundTripper", func() {
		It("creates RoundTripper with defaults", func() {
			rt := libhttp.CreateRoundTripper()
			Expect(rt).NotTo(BeNil())
		})

		It("creates RoundTripper with custom retry", func() {
			rt := libhttp.CreateRoundTripper(
				libhttp.WithRetry(10, 2*time.Second),
			)
			Expect(rt).NotTo(BeNil())
		})

		It("creates RoundTripper without retry", func() {
			rt := libhttp.CreateRoundTripper(
				libhttp.WithoutRetry(),
			)
			Expect(rt).NotTo(BeNil())
		})

		It("creates RoundTripper with logging disabled", func() {
			rt := libhttp.CreateRoundTripper(
				libhttp.WithLogging(false),
			)
			Expect(rt).NotTo(BeNil())
		})

		It("creates RoundTripper with custom timeouts", func() {
			rt := libhttp.CreateRoundTripper(
				libhttp.WithDialTimeout(10*time.Second),
				libhttp.WithTLSHandshakeTimeout(5*time.Second),
				libhttp.WithResponseHeaderTimeout(15*time.Second),
			)
			Expect(rt).NotTo(BeNil())
		})

		It("creates RoundTripper with combined timeout option", func() {
			rt := libhttp.CreateRoundTripper(
				libhttp.WithTimeouts(10*time.Second, 5*time.Second, 15*time.Second),
			)
			Expect(rt).NotTo(BeNil())
		})

		It("creates RoundTripper with HTTP/2 disabled", func() {
			rt := libhttp.CreateRoundTripper(
				libhttp.WithHTTP2(false),
			)
			Expect(rt).NotTo(BeNil())
		})

		It("creates RoundTripper with custom connection pool settings", func() {
			rt := libhttp.CreateRoundTripper(
				libhttp.WithMaxIdleConns(50),
				libhttp.WithIdleConnTimeout(60*time.Second),
			)
			Expect(rt).NotTo(BeNil())
		})

		It("creates RoundTripper with custom keep-alive", func() {
			rt := libhttp.CreateRoundTripper(
				libhttp.WithDialKeepAlive(15 * time.Second),
			)
			Expect(rt).NotTo(BeNil())
		})

		It("creates RoundTripper with custom Expect-Continue timeout", func() {
			rt := libhttp.CreateRoundTripper(
				libhttp.WithExpectContinueTimeout(2 * time.Second),
			)
			Expect(rt).NotTo(BeNil())
		})

		It("creates RoundTripper without proxy", func() {
			rt := libhttp.CreateRoundTripper(
				libhttp.WithoutProxy(),
			)
			Expect(rt).NotTo(BeNil())
		})

		It("creates RoundTripper with custom proxy", func() {
			proxyURL, err := url.Parse("http://proxy.example.com:8080")
			Expect(err).To(BeNil())

			rt := libhttp.CreateRoundTripper(
				libhttp.WithProxy(func(*http.Request) (*url.URL, error) {
					return proxyURL, nil
				}),
			)
			Expect(rt).NotTo(BeNil())
		})

		It("creates RoundTripper with TLS config", func() {
			tlsConfig := &tls.Config{
				MinVersion: tls.VersionTLS13,
			}
			rt := libhttp.CreateRoundTripper(
				libhttp.WithTLSConfig(tlsConfig),
			)
			Expect(rt).NotTo(BeNil())
		})

		It("creates RoundTripper with multiple options", func() {
			rt := libhttp.CreateRoundTripper(
				libhttp.WithRetry(3, time.Second),
				libhttp.WithLogging(true),
				libhttp.WithHTTP2(true),
				libhttp.WithMaxIdleConns(200),
				libhttp.WithDialTimeout(20*time.Second),
			)
			Expect(rt).NotTo(BeNil())
		})

		It("handles negative timeout values gracefully", func() {
			rt := libhttp.CreateRoundTripper(
				libhttp.WithDialTimeout(-1*time.Second),
				libhttp.WithTLSHandshakeTimeout(-1*time.Second),
				libhttp.WithIdleConnTimeout(-1*time.Second),
			)
			Expect(rt).NotTo(BeNil())
		})

		It("handles negative connection pool values gracefully", func() {
			rt := libhttp.CreateRoundTripper(
				libhttp.WithMaxIdleConns(-10),
			)
			Expect(rt).NotTo(BeNil())
		})
	})

	Context("Backward Compatibility", func() {
		It("CreateDefaultRoundTripper works as before", func() {
			rt := libhttp.CreateDefaultRoundTripper()
			Expect(rt).NotTo(BeNil())
		})
	})
})
