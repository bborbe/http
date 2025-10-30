// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"crypto/tls"
	stderrors "errors"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/bborbe/errors"
	"github.com/golang/glog"
)

// ErrTooManyRedirects is a sentinel error indicating that the maximum number of redirects has been exceeded.
var ErrTooManyRedirects = stderrors.New("too many redirects")

// Proxy defines a function that determines which proxy to use for a given request.
// It returns the proxy URL to use, or nil if no proxy should be used.
type Proxy func(req *http.Request) (*url.URL, error)

// CheckRedirect defines a function that controls the behavior of redirects.
// It receives the upcoming request and the requests made already in oldest-to-newest order.
type CheckRedirect func(req *http.Request, via []*http.Request) error

// DialFunc defines a function that establishes network connections.
// It should return a connection to the given network address.
type DialFunc func(ctx context.Context, network, address string) (net.Conn, error)

// ClientBuilder defines the interface for building configured HTTP clients.
// It provides a fluent API for configuring various aspects of HTTP client behavior.
type ClientBuilder interface {
	WithRetry(retryLimit int, retryDelay time.Duration) ClientBuilder
	WithoutRetry() ClientBuilder
	WithProxy() ClientBuilder
	WithoutProxy() ClientBuilder
	// WithRedirects controls how many redirects are allowed
	// 0 = no redirects, -1 = infinit redirects, 10 = 10 max redirects
	WithRedirects(maxRedirect int) ClientBuilder
	// WithoutRedirects is equal to WithRedirects(0)
	WithoutRedirects() ClientBuilder
	WithTimeout(timeout time.Duration) ClientBuilder
	WithDialFunc(dialFunc DialFunc) ClientBuilder
	WithInsecureSkipVerify(insecureSkipVerify bool) ClientBuilder
	WithClientCert(caCertPath string, clientCertPath string, clientKeyPath string) ClientBuilder
	Build(ctx context.Context) (*http.Client, error)
	BuildRoundTripper(ctx context.Context) (http.RoundTripper, error)
}

// HTTPClientBuilder is deprecated. Use ClientBuilder instead.
//
// Deprecated: Use ClientBuilder to avoid package name stuttering.
//
//nolint:revive
type HTTPClientBuilder = ClientBuilder

// HttpClientBuilder is deprecated. Use ClientBuilder instead.
//
// Deprecated: Use ClientBuilder for correct Go naming conventions and to avoid package name stuttering.
//
//nolint:revive
type HttpClientBuilder = ClientBuilder

// NewClientBuilder creates a new HTTP client builder with sensible defaults.
// Default configuration includes: no proxy, max 10 redirects, 30 second timeout, no retry.
func NewClientBuilder() ClientBuilder {
	b := new(httpClientBuilder)
	b.WithoutProxy()
	b.WithRedirects(10)
	b.WithTimeout(30 * time.Second)
	b.WithoutRetry()
	return b
}

type httpClientBuilder struct {
	proxy Proxy
	// maxRedirect -1 = infinit, 0 = none, and other number limits the redirects
	maxRedirect        int
	timeout            time.Duration
	dialFunc           DialFunc
	insecureSkipVerify bool
	retryLimit         int
	retryDelay         time.Duration
	caCertPath         string
	clientCertPath     string
	clientKeyPath      string
}

func (h *httpClientBuilder) WithRetry(retryLimit int, retryDelay time.Duration) ClientBuilder {
	h.retryLimit = retryLimit
	h.retryDelay = retryDelay
	return h
}

func (h *httpClientBuilder) WithoutRetry() ClientBuilder {
	h.retryLimit = 0
	h.retryDelay = 0
	return h
}

func (h *httpClientBuilder) WithClientCert(
	caCertPath string,
	clientCertPath string,
	clientKeyPath string,
) ClientBuilder {
	h.caCertPath = caCertPath
	h.clientCertPath = clientCertPath
	h.clientKeyPath = clientKeyPath
	return h
}

func (h *httpClientBuilder) WithTimeout(timeout time.Duration) ClientBuilder {
	h.timeout = timeout
	return h
}

func (h *httpClientBuilder) WithDialFunc(dialFunc DialFunc) ClientBuilder {
	h.dialFunc = dialFunc
	return h
}

func (h *httpClientBuilder) BuildDialFunc() DialFunc {
	if h.dialFunc != nil {
		return h.dialFunc
	}
	return (&net.Dialer{
		Timeout: h.timeout,
	}).DialContext
}

func (h *httpClientBuilder) BuildRoundTripper(ctx context.Context) (http.RoundTripper, error) {
	if glog.V(5) {
		glog.Infof("build http roundTripper")
	}
	tlsClientConfig, err := h.createTLSConfig(ctx)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "create tlsConfig failed")
	}
	var roundTripper http.RoundTripper = &http.Transport{
		Proxy:           h.proxy,
		DialContext:     h.BuildDialFunc(),
		TLSClientConfig: tlsClientConfig,
	}
	if h.retryDelay > 0 && h.retryLimit > 0 {
		roundTripper = NewRoundTripperRetry(roundTripper, h.retryLimit, h.retryDelay)
	}
	return roundTripper, nil
}

func (h *httpClientBuilder) createTLSConfig(ctx context.Context) (*tls.Config, error) {
	tlsClientConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if h.caCertPath != "" && h.clientCertPath != "" && h.clientKeyPath != "" {
		var err error
		tlsClientConfig, err = CreateTLSClientConfig(
			ctx,
			h.caCertPath,
			h.clientCertPath,
			h.clientKeyPath,
		)
		if err != nil {
			return nil, errors.Wrap(ctx, err, "create tls config failed")
		}
	}
	tlsClientConfig.InsecureSkipVerify = h.insecureSkipVerify

	return tlsClientConfig, nil
}

func (h *httpClientBuilder) Build(ctx context.Context) (*http.Client, error) {
	if glog.V(5) {
		glog.Infof("build http client")
	}
	roundTripper, err := h.BuildRoundTripper(ctx)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "build roundTripper failed")
	}

	return &http.Client{
		Transport:     roundTripper,
		CheckRedirect: h.createCheckRedirect(),
	}, nil
}

func (h *httpClientBuilder) WithProxy() ClientBuilder {
	h.proxy = http.ProxyFromEnvironment
	return h
}

func (h *httpClientBuilder) WithoutProxy() ClientBuilder {
	h.proxy = nil
	return h
}

func (h *httpClientBuilder) WithRedirects(maxRedirect int) ClientBuilder {
	h.maxRedirect = maxRedirect
	return h
}

func (h *httpClientBuilder) WithoutRedirects() ClientBuilder {
	h.maxRedirect = 0
	return h
}

func (h *httpClientBuilder) WithInsecureSkipVerify(insecureSkipVerify bool) ClientBuilder {
	h.insecureSkipVerify = insecureSkipVerify
	return h
}

func (h *httpClientBuilder) createCheckRedirect() func(req *http.Request, via []*http.Request) error {
	switch h.maxRedirect {
	case -1:
		return func(req *http.Request, via []*http.Request) error {
			return nil
		}
	case 0:
		return func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	default:
		return func(req *http.Request, via []*http.Request) error {
			if len(via) >= h.maxRedirect {
				return errors.Wrapf(
					req.Context(),
					ErrTooManyRedirects,
					"stopped after %d redirects",
					h.maxRedirect,
				)
			}
			return nil
		}
	}
}
