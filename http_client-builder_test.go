// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
)

var _ = Describe("ClientBuilder", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	Context("NewClientBuilder", func() {
		It("creates builder with defaults", func() {
			builder := libhttp.NewClientBuilder()
			Expect(builder).NotTo(BeNil())
		})

		It("builds client with default settings", func() {
			client, err := libhttp.NewClientBuilder().Build(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(client).NotTo(BeNil())
			Expect(client.Transport).NotTo(BeNil())
		})
	})

	Context("WithTimeout", func() {
		It("sets custom timeout in dialer", func() {
			timeout := 10 * time.Second
			client, err := libhttp.NewClientBuilder().
				WithTimeout(timeout).
				Build(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(client).NotTo(BeNil())
			// Timeout affects the DialContext function, not client.Timeout
			Expect(client.Transport).NotTo(BeNil())
		})
	})

	Context("WithRetry", func() {
		It("configures retry", func() {
			roundTripper, err := libhttp.NewClientBuilder().
				WithRetry(3, time.Second).
				BuildRoundTripper(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(roundTripper).NotTo(BeNil())
		})
	})

	Context("WithoutRetry", func() {
		It("disables retry", func() {
			roundTripper, err := libhttp.NewClientBuilder().
				WithoutRetry().
				BuildRoundTripper(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(roundTripper).NotTo(BeNil())
		})
	})

	Context("WithProxy", func() {
		It("enables proxy", func() {
			client, err := libhttp.NewClientBuilder().
				WithProxy().
				Build(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(client).NotTo(BeNil())
		})
	})

	Context("WithoutProxy", func() {
		It("disables proxy", func() {
			client, err := libhttp.NewClientBuilder().
				WithoutProxy().
				Build(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(client).NotTo(BeNil())
		})
	})

	Context("WithRedirects", func() {
		It("allows 5 redirects", func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Redirect(w, r, "/final", http.StatusFound)
				}),
			)
			defer server.Close()

			client, err := libhttp.NewClientBuilder().
				WithRedirects(5).
				Build(ctx)
			Expect(err).NotTo(HaveOccurred())

			resp, err := client.Get(server.URL)
			if err == nil {
				resp.Body.Close()
			}
		})

		It("stops after max redirects", func() {
			redirectCount := 0
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					redirectCount++
					http.Redirect(w, r, "/loop", http.StatusFound)
				}),
			)
			defer server.Close()

			client, err := libhttp.NewClientBuilder().
				WithRedirects(3).
				Build(ctx)
			Expect(err).NotTo(HaveOccurred())

			_, err = client.Get(server.URL)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stopped after 3 redirects"))
		})

		It("allows infinite redirects with -1", func() {
			client, err := libhttp.NewClientBuilder().
				WithRedirects(-1).
				Build(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(client.CheckRedirect).NotTo(BeNil())
		})
	})

	Context("WithoutRedirects", func() {
		It("disables redirects", func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Redirect(w, r, "/target", http.StatusFound)
				}),
			)
			defer server.Close()

			client, err := libhttp.NewClientBuilder().
				WithoutRedirects().
				Build(ctx)
			Expect(err).NotTo(HaveOccurred())

			resp, err := client.Get(server.URL)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusFound))
		})
	})

	Context("WithInsecureSkipVerify", func() {
		It("enables insecure skip verify", func() {
			client, err := libhttp.NewClientBuilder().
				WithInsecureSkipVerify(true).
				Build(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(client).NotTo(BeNil())

			// Verify TLS config
			transport := client.Transport.(*http.Transport)
			Expect(transport.TLSClientConfig).NotTo(BeNil())
			Expect(transport.TLSClientConfig.InsecureSkipVerify).To(BeTrue())
			Expect(transport.TLSClientConfig.MinVersion).To(Equal(uint16(tls.VersionTLS12)))
		})

		It("disables insecure skip verify by default", func() {
			client, err := libhttp.NewClientBuilder().
				WithInsecureSkipVerify(false).
				Build(ctx)
			Expect(err).NotTo(HaveOccurred())

			transport := client.Transport.(*http.Transport)
			Expect(transport.TLSClientConfig.InsecureSkipVerify).To(BeFalse())
		})
	})

	Context("WithDialFunc", func() {
		It("sets custom dial function", func() {
			customDialer := func(ctx context.Context, network, address string) (net.Conn, error) {
				return (&net.Dialer{
					Timeout: 5 * time.Second,
				}).DialContext(ctx, network, address)
			}

			client, err := libhttp.NewClientBuilder().
				WithDialFunc(customDialer).
				Build(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(client).NotTo(BeNil())
		})
	})

	Context("BuildRoundTripper", func() {
		It("builds round tripper", func() {
			rt, err := libhttp.NewClientBuilder().BuildRoundTripper(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(rt).NotTo(BeNil())
		})

		It("builds round tripper with retry", func() {
			rt, err := libhttp.NewClientBuilder().
				WithRetry(3, 100*time.Millisecond).
				BuildRoundTripper(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(rt).NotTo(BeNil())
		})
	})

	Context("Fluent API", func() {
		It("chains multiple configurations", func() {
			client, err := libhttp.NewClientBuilder().
				WithTimeout(5*time.Second).
				WithoutProxy().
				WithRedirects(5).
				WithRetry(3, 500*time.Millisecond).
				Build(ctx)

			Expect(err).NotTo(HaveOccurred())
			Expect(client).NotTo(BeNil())
			Expect(client.Transport).NotTo(BeNil())
			Expect(client.CheckRedirect).NotTo(BeNil())
		})
	})

	Context("Error scenarios", func() {
		It("handles ErrTooManyRedirects", func() {
			redirectCount := 0
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					redirectCount++
					http.Redirect(w, r, "/loop", http.StatusMovedPermanently)
				}),
			)
			defer server.Close()

			client, err := libhttp.NewClientBuilder().
				WithRedirects(2).
				Build(ctx)
			Expect(err).NotTo(HaveOccurred())

			_, err = client.Get(server.URL)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stopped after 2 redirects"))
		})
	})
})
