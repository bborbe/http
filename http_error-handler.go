// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"errors"
	"fmt"
	"net/http"

	liberrors "github.com/bborbe/errors"
	"github.com/golang/glog"
)

// ErrorWithStatusCode defines an error that can provide an HTTP status code.
// This interface allows errors to specify which HTTP status code should be returned to clients.
type ErrorWithStatusCode interface {
	error
	StatusCode() int
}

// ErrorWithCode defines an error that can provide a typed error code.
// This interface allows errors to specify an error code (e.g., VALIDATION_ERROR, NOT_FOUND)
// for structured JSON error responses.
type ErrorWithCode interface {
	error
	Code() string
}

// WrapWithStatusCode wraps a existing error with statusCode used by ErrorHandler
func WrapWithStatusCode(err error, code int) ErrorWithStatusCode {
	return &statusCodeError{
		err:  err,
		code: code,
	}
}

//nolint:errname // private implementation type
type statusCodeError struct {
	err  error
	code int
}

// Error returns the error message.
func (e statusCodeError) Error() string {
	return e.err.Error()
}

// StatusCode returns the HTTP status code associated with this error.
func (e statusCodeError) StatusCode() int {
	return e.code
}

// WrapWithCode wraps an error with both an error code and HTTP status code.
// This allows the error to be used with JSON error handlers that return structured error responses.
//
// Example:
//
//	err := WrapWithCode(
//	    errors.New(ctx, "columnGroup '' is unknown"),
//	    ErrorCodeValidation,
//	    http.StatusBadRequest,
//	)
func WrapWithCode(err error, code string, statusCode int) error {
	return &errorWithCodeAndStatus{
		err:        err,
		code:       code,
		statusCode: statusCode,
	}
}

// WrapWithDetails wraps an error with code, status, and structured details.
// This is a convenience helper that combines WrapWithCode with adding data to the error.
// Details can include any JSON-serializable values including arrays and nested objects.
//
// Example:
//
//	err := WrapWithDetails(
//	    errors.New(ctx, "columnGroup '' is unknown"),
//	    ErrorCodeValidation,
//	    http.StatusBadRequest,
//	    map[string]any{
//	        "field":    "columnGroup",
//	        "expected": []string{"day", "week", "month", "year"},
//	    },
//	)
func WrapWithDetails(err error, code string, statusCode int, details map[string]any) error {
	wrappedErr := WrapWithCode(err, code, statusCode)
	return liberrors.AddDataToError(wrappedErr, details)
}

//nolint:errname // private implementation type
type errorWithCodeAndStatus struct {
	err        error
	code       string
	statusCode int
}

// Error returns the error message.
func (e *errorWithCodeAndStatus) Error() string {
	return e.err.Error()
}

// Code returns the error code.
func (e *errorWithCodeAndStatus) Code() string {
	return e.code
}

// StatusCode returns the HTTP status code.
func (e *errorWithCodeAndStatus) StatusCode() int {
	return e.statusCode
}

// Unwrap returns the wrapped error.
func (e *errorWithCodeAndStatus) Unwrap() error {
	return e.err
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
