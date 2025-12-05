// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

// Standard error codes for JSON error responses
const (
	ErrorCodeValidation   = "VALIDATION_ERROR" // 400 Bad Request
	ErrorCodeNotFound     = "NOT_FOUND"        // 404 Not Found
	ErrorCodeUnauthorized = "UNAUTHORIZED"     // 401 Unauthorized
	ErrorCodeForbidden    = "FORBIDDEN"        // 403 Forbidden
	ErrorCodeInternal     = "INTERNAL_ERROR"   // 500 Internal Server Error
)

// ErrorResponse wraps error details in standard JSON format.
// It provides a consistent structure for error responses across HTTP services.
//
// Example JSON output:
//
//	{
//	  "error": {
//	    "code": "VALIDATION_ERROR",
//	    "message": "columnGroup '' is unknown",
//	    "details": {
//	      "field": "columnGroup",
//	      "expected": "day|week|month|year"
//	    }
//	  }
//	}
type ErrorResponse struct {
	Error ErrorDetails `json:"error"`
}

// ErrorDetails contains structured error information for client responses.
type ErrorDetails struct {
	// Code is the error type identifier (e.g., VALIDATION_ERROR, NOT_FOUND)
	Code string `json:"code"`

	// Message is the human-readable error message
	Message string `json:"message"`

	// Details contains optional structured data extracted from errors.HasData interface.
	// Omitted from JSON if nil or empty.
	Details map[string]string `json:"details,omitempty"`
}
