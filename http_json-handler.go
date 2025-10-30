// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/bborbe/errors"
)

//counterfeiter:generate -o mocks/http-json-handler.go --fake-name HttpJsonHandler . JSONHandler

// JSONHandler defines the interface for handlers that return JSON responses.
// Implementations should return the data to be JSON-encoded and any error that occurred.
type JSONHandler interface {
	ServeHTTP(ctx context.Context, req *http.Request) (interface{}, error)
}

// JSONHandlerFunc is an adapter to allow the use of ordinary functions as JSONHandlers.
// If f is a function with the appropriate signature, JSONHandlerFunc(f) is a JSONHandler that calls f.
type JSONHandlerFunc func(ctx context.Context, req *http.Request) (interface{}, error)

// ServeHTTP calls f(ctx, req).
func (j JSONHandlerFunc) ServeHTTP(ctx context.Context, req *http.Request) (interface{}, error) {
	return j(ctx, req)
}

// JsonHandler is deprecated. Use JSONHandler instead.
//
// Deprecated: Use JSONHandler for correct Go naming conventions.
//
//nolint:revive
type JsonHandler = JSONHandler

// JsonHandlerFunc is deprecated. Use JSONHandlerFunc instead.
//
// Deprecated: Use JSONHandlerFunc for correct Go naming conventions.
//
//nolint:revive
type JsonHandlerFunc = JSONHandlerFunc

// NewJSONHandler wraps a JSONHandler to automatically encode responses as JSON.
// It sets the appropriate Content-Type header and handles JSON marshaling.
// Returns a WithError handler that can be used with error handling middleware.
func NewJSONHandler(jsonHandler JSONHandler) WithError {
	return WithErrorFunc(
		func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
			result, err := jsonHandler.ServeHTTP(ctx, req)
			if err != nil {
				return errors.Wrap(ctx, err, "json handler failed")
			}
			resp.Header().Add(ContentTypeHeaderName, ApplicationJsonContentType)
			if err := json.NewEncoder(resp).Encode(result); err != nil {
				return errors.Wrap(ctx, err, "encode json failed")
			}
			return nil
		},
	)
}

// NewJsonHandler is deprecated. Use NewJSONHandler instead.
//
// Deprecated: Use NewJSONHandler for correct Go naming conventions.
//
//nolint:revive
func NewJsonHandler(jsonHandler JsonHandler) WithError {
	return NewJSONHandler(jsonHandler)
}
