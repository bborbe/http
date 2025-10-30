// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bborbe/errors"
	"github.com/bborbe/run"
	"github.com/golang/glog"
)

// ServerOptions configures HTTP server behavior.
// ReadHeaderTimeout sets the maximum duration for reading request headers (Slowloris protection).
// ReadTimeout sets the maximum duration for reading the entire request (headers + body).
// WriteTimeout sets the maximum duration for writing the response.
// IdleTimeout sets the maximum duration to wait for the next request when keep-alives are enabled.
// ShutdownTimeout sets the maximum duration to wait for graceful shutdown.
// MaxHeaderBytes limits the size of request headers to prevent memory exhaustion attacks.
// TLSConfig provides custom TLS configuration for HTTPS servers.
// CertFile and KeyFile specify paths to TLS certificate and private key files for HTTPS.
type ServerOptions struct {
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	MaxHeaderBytes    int
	TLSConfig         *tls.Config
	CertFile          string
	KeyFile           string
}

// NewServer creates an HTTP server that listens on the specified address.
// It returns a run.Func that handles graceful shutdown when the context is cancelled.
// The addr parameter should be in the format ":port" or "host:port".
func NewServer(
	addr string,
	router http.Handler,
	optionFns ...func(serverOptions *ServerOptions),
) run.Func {
	serverOptions := CreateServerOptions(optionFns...)
	server := CreateHttpServer(addr, router, serverOptions)
	return func(ctx context.Context) error {
		go func() {
			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(
				context.Background(),
				serverOptions.ShutdownTimeout,
			)
			defer cancel()
			if err := server.Shutdown(shutdownCtx); err != nil {
				glog.Warningf("shutdown failed: %v", err)
			}
		}()
		err := runServer(server, serverOptions)
		if errors.Is(err, http.ErrServerClosed) {
			glog.V(0).Info(err)
			return nil
		}
		return errors.Wrap(ctx, err, "httpServer failed")
	}
}

func CreateHttpServer(
	addr string,
	router http.Handler,
	serverOptions ServerOptions,
) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           router,
		TLSConfig:         serverOptions.TLSConfig,
		ReadHeaderTimeout: serverOptions.ReadHeaderTimeout,
		ReadTimeout:       serverOptions.ReadTimeout,
		WriteTimeout:      serverOptions.WriteTimeout,
		IdleTimeout:       serverOptions.IdleTimeout,
		MaxHeaderBytes:    serverOptions.MaxHeaderBytes,
	}
}

func CreateServerOptions(optionFns ...func(serverOptions *ServerOptions)) ServerOptions {
	serverOptions := ServerOptions{
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		ShutdownTimeout:   5 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
		TLSConfig:         nil,
	}
	for _, optionFn := range optionFns {
		optionFn(&serverOptions)
	}
	return serverOptions
}

// NewServerWithPort creates an HTTP server that listens on the specified port.
// It returns a run.Func that can be used with the run package for graceful shutdown.
// The server will bind to all interfaces on the given port (e.g., port 8080 becomes ":8080").
func NewServerWithPort(
	port int,
	router http.Handler,
	optionFns ...func(serverOptions *ServerOptions),
) run.Func {
	return NewServer(
		fmt.Sprintf(":%d", port),
		router,
		optionFns...,
	)
}

// NewServerTLS creates an HTTPS server with TLS support.
// It listens on the specified address using the provided certificate and key files.
// The server includes error log filtering to skip common TLS handshake errors.
// Returns a run.Func for graceful shutdown management.
func NewServerTLS(
	addr string,
	router http.Handler,
	serverCertPath string,
	serverKeyPath string,
	optionFns ...func(serverOptions *ServerOptions),
) run.Func {
	return NewServer(addr, router, append(
		optionFns,
		func(serverOptions *ServerOptions) {
			serverOptions.CertFile = serverCertPath
			serverOptions.KeyFile = serverKeyPath
		},
	)...)
}

func runServer(server *http.Server, serverOption ServerOptions) error {
	if serverOption.CertFile != "" && serverOption.KeyFile != "" {
		return server.ListenAndServeTLS(serverOption.CertFile, serverOption.KeyFile)
	}
	return server.ListenAndServe()
}

// NewSkipErrorWriter creates a writer that filters out TLS handshake error messages.
// It wraps the given writer and skips writing messages containing "http: TLS handshake error from".
// This is useful for reducing noise in server logs when dealing with automated scanners or bots.
func NewSkipErrorWriter(writer io.Writer) io.Writer {
	return &skipErrorWriter{
		writer: writer,
	}
}

// skipErrorWriter is an io.Writer implementation that filters out specific error messages.
type skipErrorWriter struct {
	writer io.Writer
}

// Write implements io.Writer, filtering out TLS handshake error messages.
func (s *skipErrorWriter) Write(p []byte) (n int, err error) {
	if bytes.Contains(p, []byte("http: TLS handshake error from")) {
		// skip
		return len(p), nil
	}
	return s.writer.Write(p)
}
