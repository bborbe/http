// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/bborbe/errors"
	"github.com/bborbe/run"
	"github.com/golang/glog"
)

// NewServerWithPort creates an HTTP server that listens on the specified port.
// It returns a run.Func that can be used with the run package for graceful shutdown.
// The server will bind to all interfaces on the given port (e.g., port 8080 becomes ":8080").
func NewServerWithPort(port int, router http.Handler) run.Func {
	return NewServer(
		fmt.Sprintf(":%d", port),
		router,
	)
}

// NewServer creates an HTTP server that listens on the specified address.
// It returns a run.Func that handles graceful shutdown when the context is cancelled.
// The addr parameter should be in the format ":port" or "host:port".
func NewServer(addr string, router http.Handler) run.Func {
	return func(ctx context.Context) error {

		server := &http.Server{
			Addr:      addr,
			Handler:   router,
			TLSConfig: nil,
		}
		go func() {
			select {
			case <-ctx.Done():
				if err := server.Shutdown(ctx); err != nil {
					glog.Warningf("shutdown failed: %v", err)
				}
			}
		}()
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			glog.V(0).Info(err)
			return nil
		}
		return errors.Wrapf(ctx, err, "httpServer failed")
	}
}

// NewServerTLS creates an HTTPS server with TLS support.
// It listens on the specified address using the provided certificate and key files.
// The server includes error log filtering to skip common TLS handshake errors.
// Returns a run.Func for graceful shutdown management.
func NewServerTLS(addr string, router http.Handler, serverCertPath string, serverKeyPath string) run.Func {
	return func(ctx context.Context) error {
		server := &http.Server{
			Addr:     addr,
			Handler:  router,
			ErrorLog: log.New(NewSkipErrorWriter(log.Writer()), "", log.LstdFlags),
		}
		go func() {
			select {
			case <-ctx.Done():
				if err := server.Shutdown(ctx); err != nil {
					glog.Warningf("shutdown failed: %v", err)
				}
			}
		}()
		err := server.ListenAndServeTLS(serverCertPath, serverKeyPath)
		if errors.Is(err, http.ErrServerClosed) {
			glog.V(0).Info(err)
			return nil
		}
		return errors.Wrapf(ctx, err, "httpServer failed")
	}
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
