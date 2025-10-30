// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/bborbe/errors"
)

// CreateDefaultRoundTripper creates a RoundTripper with default configuration.
// It includes retry logic (5 retries with 1 second delay) and request/response logging.
// The transport uses standard HTTP/2 settings with reasonable timeouts.
func CreateDefaultRoundTripper() RoundTripper {
	return createDefaultRoundTripper(nil)
}

// CreateDefaultRoundTripperTLS creates a RoundTripper with TLS client certificate authentication.
// It loads the specified CA certificate, client certificate, and private key for mutual TLS authentication.
// The returned RoundTripper includes the same retry logic and logging as CreateDefaultRoundTripper.
func CreateDefaultRoundTripperTLS(
	ctx context.Context,
	caCertPath string,
	clientCertPath string,
	clientKeyPath string,
) (RoundTripper, error) {
	tlsClientConfig, err := CreateTLSClientConfig(ctx, caCertPath, clientCertPath, clientKeyPath)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "create tls client config failed")
	}
	return createDefaultRoundTripper(tlsClientConfig), nil
}

// CreateDefaultRoundTripperTls is deprecated. Use CreateDefaultRoundTripperTLS instead.
//
// Deprecated: Use CreateDefaultRoundTripperTLS for correct Go naming conventions.
//
//nolint:revive
func CreateDefaultRoundTripperTls(
	ctx context.Context,
	caCertPath string,
	clientCertPath string,
	clientKeyPath string,
) (RoundTripper, error) {
	return CreateDefaultRoundTripperTLS(ctx, caCertPath, clientCertPath, clientKeyPath)
}

// CreateTLSClientConfig creates a TLS configuration for mutual TLS authentication.
// It loads the CA certificate for server verification and the client certificate for client authentication.
// The configuration enforces server certificate verification (InsecureSkipVerify is false).
func CreateTLSClientConfig(
	ctx context.Context,
	caCertPath string,
	clientCertPath string,
	clientKeyPath string,
) (*tls.Config, error) {
	// Load the client certificate and private key
	clientCert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "load client certificate and key failed")
	}

	// Load the CA certificate to verify the server
	// #nosec G304 -- caCertPath comes from application configuration, not user input
	caCertPEM, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "read CA certificate failed")
	}
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCertPEM); !ok {
		return nil, errors.Wrap(ctx, err, "append CA certificate to pool failed")
	}

	// Set up TLS configuration with the client certificate and CA
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{clientCert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: false, // Ensures server certificate is verified
		MinVersion:         tls.VersionTLS12,
	}
	return tlsConfig, nil
}

// CreateTlsClientConfig is deprecated. Use CreateTLSClientConfig instead.
//
// Deprecated: Use CreateTLSClientConfig for correct Go naming conventions.
//
//nolint:revive
func CreateTlsClientConfig(
	ctx context.Context,
	caCertPath string,
	clientCertPath string,
	clientKeyPath string,
) (*tls.Config, error) {
	return CreateTLSClientConfig(ctx, caCertPath, clientCertPath, clientKeyPath)
}

func createDefaultRoundTripper(tlsClientConfig *tls.Config) RoundTripper {
	return NewRoundTripperRetry(
		NewRoundTripperLog(
			&http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: defaultTransportDialContext(&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}),
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				ResponseHeaderTimeout: 30 * time.Second,
				TLSClientConfig:       tlsClientConfig,
			},
		),
		5,
		time.Second,
	)
}

func defaultTransportDialContext(
	dialer *net.Dialer,
) func(context.Context, string, string) (net.Conn, error) {
	return dialer.DialContext
}
