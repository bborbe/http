// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/bborbe/errors"
)

// RoundTripperOptions holds all optional configuration for RoundTripper.
type RoundTripperOptions struct {
	// TLS configuration for client certificates and server verification
	TLSConfig *tls.Config

	// Retry configuration
	RetryCount int
	RetryDelay time.Duration

	// Transport settings
	DialTimeout           time.Duration
	DialKeepAlive         time.Duration
	ForceAttemptHTTP2     bool
	MaxIdleConns          int
	IdleConnTimeout       time.Duration
	TLSHandshakeTimeout   time.Duration
	ExpectContinueTimeout time.Duration
	ResponseHeaderTimeout time.Duration

	// Proxy configuration
	Proxy func(*http.Request) (*url.URL, error)

	// Enable request/response logging
	EnableLogging bool
}

// RoundTripperOption defines a function type for modifying RoundTripper configuration.
type RoundTripperOption func(*RoundTripperOptions)

// CreateRoundTripper creates a RoundTripper with the specified options.
// If no options are provided, sensible defaults are used.
//
// Default configuration:
//   - Retry: 5 attempts with 1 second delay
//   - Logging: enabled
//   - HTTP/2: enabled
//   - Timeouts: 30s dial, 10s TLS handshake, 30s response headers
//   - Max idle connections: 100
//   - Idle timeout: 90s
//   - Proxy: from environment variables
func CreateRoundTripper(options ...RoundTripperOption) RoundTripper {
	// Initialize with default values
	opts := RoundTripperOptions{
		TLSConfig:             nil,
		RetryCount:            5,
		RetryDelay:            time.Second,
		DialTimeout:           30 * time.Second,
		DialKeepAlive:         30 * time.Second,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		Proxy:                 http.ProxyFromEnvironment,
		EnableLogging:         true,
	}

	// Apply all provided options
	for _, option := range options {
		option(&opts)
	}

	// Create the base transport
	transport := &http.Transport{
		Proxy: opts.Proxy,
		DialContext: (&net.Dialer{
			Timeout:   opts.DialTimeout,
			KeepAlive: opts.DialKeepAlive,
		}).DialContext,
		ForceAttemptHTTP2:     opts.ForceAttemptHTTP2,
		MaxIdleConns:          opts.MaxIdleConns,
		IdleConnTimeout:       opts.IdleConnTimeout,
		TLSHandshakeTimeout:   opts.TLSHandshakeTimeout,
		ExpectContinueTimeout: opts.ExpectContinueTimeout,
		ResponseHeaderTimeout: opts.ResponseHeaderTimeout,
		TLSClientConfig:       opts.TLSConfig,
	}

	// Wrap with logging if enabled
	var roundTripper RoundTripper = transport
	if opts.EnableLogging {
		roundTripper = NewRoundTripperLog(roundTripper)
	}

	// Wrap with retry if configured
	if opts.RetryCount > 0 {
		roundTripper = NewRoundTripperRetry(
			roundTripper,
			opts.RetryCount,
			opts.RetryDelay,
		)
	}

	return roundTripper
}

// WithTLSConfig sets the TLS configuration for client certificates and server verification.
func WithTLSConfig(config *tls.Config) RoundTripperOption {
	return func(opts *RoundTripperOptions) {
		opts.TLSConfig = config
	}
}

// WithTLSFiles loads TLS configuration from certificate files.
// It loads the specified CA certificate, client certificate, and private key for mutual TLS authentication.
func WithTLSFiles(
	ctx context.Context,
	caCertPath, clientCertPath, clientKeyPath string,
) (RoundTripperOption, error) {
	tlsConfig, err := CreateTLSClientConfig(ctx, caCertPath, clientCertPath, clientKeyPath)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "create TLS client config failed")
	}
	return WithTLSConfig(tlsConfig), nil
}

// WithRetry configures retry behavior.
// Set count to 0 to disable retries.
func WithRetry(count int, delay time.Duration) RoundTripperOption {
	return func(opts *RoundTripperOptions) {
		opts.RetryCount = count
		opts.RetryDelay = delay
	}
}

// WithoutRetry disables retry logic.
func WithoutRetry() RoundTripperOption {
	return func(opts *RoundTripperOptions) {
		opts.RetryCount = 0
	}
}

// WithLogging enables or disables request/response logging.
func WithLogging(enabled bool) RoundTripperOption {
	return func(opts *RoundTripperOptions) {
		opts.EnableLogging = enabled
	}
}

// WithDialTimeout sets the timeout for establishing connections.
func WithDialTimeout(timeout time.Duration) RoundTripperOption {
	return func(opts *RoundTripperOptions) {
		if timeout < 0 {
			timeout = 30 * time.Second // Use sensible default
		}
		opts.DialTimeout = timeout
	}
}

// WithDialKeepAlive sets the keep-alive period for active connections.
func WithDialKeepAlive(keepAlive time.Duration) RoundTripperOption {
	return func(opts *RoundTripperOptions) {
		if keepAlive < 0 {
			keepAlive = 30 * time.Second // Use sensible default
		}
		opts.DialKeepAlive = keepAlive
	}
}

// WithHTTP2 enables or disables HTTP/2.
func WithHTTP2(enabled bool) RoundTripperOption {
	return func(opts *RoundTripperOptions) {
		opts.ForceAttemptHTTP2 = enabled
	}
}

// WithMaxIdleConns sets the maximum number of idle connections.
func WithMaxIdleConns(n int) RoundTripperOption {
	return func(opts *RoundTripperOptions) {
		if n < 0 {
			n = 100 // Use sensible default
		}
		opts.MaxIdleConns = n
	}
}

// WithIdleConnTimeout sets the timeout for idle connections.
func WithIdleConnTimeout(timeout time.Duration) RoundTripperOption {
	return func(opts *RoundTripperOptions) {
		if timeout < 0 {
			timeout = 90 * time.Second // Use sensible default
		}
		opts.IdleConnTimeout = timeout
	}
}

// WithTLSHandshakeTimeout sets the timeout for TLS handshakes.
func WithTLSHandshakeTimeout(timeout time.Duration) RoundTripperOption {
	return func(opts *RoundTripperOptions) {
		if timeout < 0 {
			timeout = 10 * time.Second // Use sensible default
		}
		opts.TLSHandshakeTimeout = timeout
	}
}

// WithExpectContinueTimeout sets the timeout for Expect: 100-continue responses.
func WithExpectContinueTimeout(timeout time.Duration) RoundTripperOption {
	return func(opts *RoundTripperOptions) {
		if timeout < 0 {
			timeout = 1 * time.Second // Use sensible default
		}
		opts.ExpectContinueTimeout = timeout
	}
}

// WithResponseHeaderTimeout sets the timeout for reading response headers.
func WithResponseHeaderTimeout(timeout time.Duration) RoundTripperOption {
	return func(opts *RoundTripperOptions) {
		if timeout < 0 {
			timeout = 30 * time.Second // Use sensible default
		}
		opts.ResponseHeaderTimeout = timeout
	}
}

// WithProxy sets a custom proxy function.
func WithProxy(proxy func(*http.Request) (*url.URL, error)) RoundTripperOption {
	return func(opts *RoundTripperOptions) {
		opts.Proxy = proxy
	}
}

// WithoutProxy disables proxy usage.
func WithoutProxy() RoundTripperOption {
	return func(opts *RoundTripperOptions) {
		opts.Proxy = nil
	}
}

// WithTimeouts sets all timeout values at once for convenience.
func WithTimeouts(dial, tlsHandshake, responseHeader time.Duration) RoundTripperOption {
	return func(opts *RoundTripperOptions) {
		WithDialTimeout(dial)(opts)
		WithTLSHandshakeTimeout(tlsHandshake)(opts)
		WithResponseHeaderTimeout(responseHeader)(opts)
	}
}
