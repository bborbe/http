// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/golang/glog"
)

// ErrorWithStatusCode defines an error that can provide an HTTP status code.
// This interface allows errors to specify which HTTP status code should be returned to clients.
type ErrorWithStatusCode interface {
	error
	StatusCode() int
}

// WrapWithStatusCode wraps a existing error with statusCode used by ErrorHandler
func WrapWithStatusCode(err error, code int) ErrorWithStatusCode {
	return &errorWithStatusCode{
		err:  err,
		code: code,
	}
}

type errorWithStatusCode struct {
	err  error
	code int
}

// Error returns the error message.
func (e errorWithStatusCode) Error() string {
	return e.err.Error()
}

// StatusCode returns the HTTP status code associated with this error.
func (e errorWithStatusCode) StatusCode() int {
	return e.code
}

// NewErrorHandler wraps a WithError handler to provide centralized error handling.
// It converts errors to HTTP responses with appropriate status codes and logs the results.
// If the error implements ErrorWithStatusCode, it uses that status code; otherwise defaults to 500.
func NewErrorHandler(withError WithError) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		glog.V(3).Infof("handle %s request to %s started", req.Method, req.URL.Path)
		if err := withError.ServeHTTP(ctx, resp, req); err != nil {
			var errorWithStatusCode ErrorWithStatusCode
			var statusCode = http.StatusInternalServerError
			if errors.As(err, &errorWithStatusCode) {
				statusCode = errorWithStatusCode.StatusCode()
			}
			http.Error(resp, fmt.Sprintf("request failed: %v", err), statusCode)
			glog.V(1).
				Infof("handle %s request to %s failed with status %d: %v", req.Method, req.URL.Path, statusCode, err)
			return
		}
		glog.V(3).Infof("handle %s request to %s completed", req.Method, req.URL.Path)
	})
}
