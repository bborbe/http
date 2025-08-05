// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import "net/http"

//counterfeiter:generate -o mocks/http-proxy-error-handler.go --fake-name HttpProxyErrorHandler . ProxyErrorHandler

// ProxyErrorHandler defines the interface for handling errors that occur in reverse proxy operations.
// Implementations should handle the error appropriately, such as returning error responses to clients.
type ProxyErrorHandler interface {
	HandleError(resp http.ResponseWriter, req *http.Request, err error)
}

// ProxyErrorHandlerFunc is an adapter to allow the use of ordinary functions as ProxyErrorHandler.
// If f is a function with the appropriate signature, ProxyErrorHandlerFunc(f) is a ProxyErrorHandler that calls f.
type ProxyErrorHandlerFunc func(resp http.ResponseWriter, req *http.Request, err error)

// HandleError calls f(resp, req, err).
func (p ProxyErrorHandlerFunc) HandleError(resp http.ResponseWriter, req *http.Request, err error) {
	p(resp, req, err)
}
