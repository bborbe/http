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

//counterfeiter:generate -o mocks/http-json-handler.go --fake-name HttpJsonHandler . JsonHandler

// JsonHandler defines the interface for handlers that return JSON responses.
// Implementations should return the data to be JSON-encoded and any error that occurred.
type JsonHandler interface {
	ServeHTTP(ctx context.Context, req *http.Request) (interface{}, error)
}

// JsonHandlerFunc is an adapter to allow the use of ordinary functions as JsonHandlers.
// If f is a function with the appropriate signature, JsonHandlerFunc(f) is a JsonHandler that calls f.
type JsonHandlerFunc func(ctx context.Context, req *http.Request) (interface{}, error)

// ServeHTTP calls f(ctx, req).
func (j JsonHandlerFunc) ServeHTTP(ctx context.Context, req *http.Request) (interface{}, error) {
	return j(ctx, req)
}

// NewJsonHandler wraps a JsonHandler to automatically encode responses as JSON.
// It sets the appropriate Content-Type header and handles JSON marshaling.
// Returns a WithError handler that can be used with error handling middleware.
func NewJsonHandler(jsonHandler JsonHandler) WithError {
	return WithErrorFunc(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
		result, err := jsonHandler.ServeHTTP(ctx, req)
		if err != nil {
			return errors.Wrapf(ctx, err, "json handler failed")
		}
		resp.Header().Add(ContentTypeHeaderName, ApplicationJsonContentType)
		if err := json.NewEncoder(resp).Encode(result); err != nil {
			return errors.Wrapf(ctx, err, "encode json failed")
		}
		return nil
	})
}
