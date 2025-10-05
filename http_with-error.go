// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"net/http"
)

//counterfeiter:generate -o mocks/http-with-error.go --fake-name HttpWithError . WithError

// WithError defines the interface for HTTP handlers that can return errors.
// Unlike standard http.Handler, this interface allows returning errors for centralized error handling.
type WithError interface {
	ServeHTTP(ctx context.Context, resp http.ResponseWriter, req *http.Request) error
}

// WithErrorFunc is an adapter to allow the use of ordinary functions as WithError handlers.
// If f is a function with the appropriate signature, WithErrorFunc(f) is a WithError that calls f.
type WithErrorFunc func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error

// ServeHTTP calls f(ctx, resp, req).
func (w WithErrorFunc) ServeHTTP(
	ctx context.Context,
	resp http.ResponseWriter,
	req *http.Request,
) error {
	return w(ctx, resp, req)
}
