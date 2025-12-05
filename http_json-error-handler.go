// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"errors"
	"net/http"

	liberrors "github.com/bborbe/errors"
	"github.com/golang/glog"
)

// NewJSONErrorHandler wraps a WithError handler to provide centralized JSON error handling.
// It converts errors to JSON responses with appropriate status codes and logs the results.
// If the error implements ErrorWithStatusCode, it uses that status code; otherwise defaults to 500.
// If the error implements ErrorWithCode, it uses that error code; otherwise defaults to INTERNAL_ERROR.
//
// Example usage:
//
//	handler := libhttp.NewJSONErrorHandler(
//	    libhttp.WithErrorFunc(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
//	        return libhttp.WrapWithCode(
//	            errors.New(ctx, "validation failed"),
//	            libhttp.ErrorCodeValidation,
//	            http.StatusBadRequest,
//	        )
//	    }),
//	)
func NewJSONErrorHandler(withError WithError) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		glog.V(3).Infof("handle %s request to %s started", req.Method, req.URL.Path)

		if err := withError.ServeHTTP(ctx, resp, req); err != nil {
			// Extract status code (existing pattern)
			statusCode := http.StatusInternalServerError
			var errorWithStatusCode ErrorWithStatusCode
			if errors.As(err, &errorWithStatusCode) {
				statusCode = errorWithStatusCode.StatusCode()
			}

			// Extract error code (new pattern)
			errorCode := ErrorCodeInternal
			var errorWithCode ErrorWithCode
			if errors.As(err, &errorWithCode) {
				errorCode = errorWithCode.Code()
			}

			// Extract structured details from error chain (optional)
			// Uses existing github.com/bborbe/errors.HasData interface
			details := liberrors.DataFromError(err)

			// Build error response
			errorResponse := ErrorResponse{
				Error: ErrorDetails{
					Code:    errorCode,
					Message: err.Error(),
					Details: details,
				},
			}

			// Send JSON response
			if err := SendJSONResponse(ctx, resp, errorResponse, statusCode); err != nil {
				glog.Warningf("failed to send JSON error response: %v", err)
				http.Error(resp, "internal server error", http.StatusInternalServerError)
			}

			glog.V(1).
				Infof("handle %s request to %s failed with status %d and code %s: %v",
					req.Method, req.URL.Path, statusCode, errorCode, err)
			return
		}

		glog.V(3).Infof("handle %s request to %s completed", req.Method, req.URL.Path)
	})
}
