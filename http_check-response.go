// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bytes"
	stderrors "errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/bborbe/errors"
)

// RequestFailedError represents an HTTP request that failed with a non-success status code.
// It contains information about the failed request including method, URL, and status code.
type RequestFailedError struct {
	Method     string
	URL        string
	StatusCode int
}

func (r RequestFailedError) Error() string {
	return fmt.Sprintf("%s request to %s failed with statusCode %d", r.Method, r.URL, r.StatusCode)
}

// ErrNotFound is a sentinel error used to indicate that a requested resource was not found.
var ErrNotFound = stderrors.New("not found")

// NotFound is deprecated. Use ErrNotFound instead.
//
// Deprecated: Use ErrNotFound for correct Go error naming conventions (ST1012).
//
//nolint:revive
var NotFound = ErrNotFound

// CheckResponseIsSuccessful validates that an HTTP response indicates success.
// It returns ErrNotFound error for 404 responses, and RequestFailedError for other non-success status codes.
// Success is defined as 2xx or 3xx status codes. The response body is preserved for further reading.
func CheckResponseIsSuccessful(req *http.Request, resp *http.Response) error {
	if resp.StatusCode == 404 {
		return errors.Wrapf(
			req.Context(),
			ErrNotFound,
			"%s to %s failed with status %d",
			req.Method,
			req.URL.String(),
			resp.StatusCode,
		)
	}
	if resp.StatusCode/100 != 2 && resp.StatusCode/100 != 3 {
		content, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewBuffer(content))
		return errors.AddDataToError(
			errors.Wrapf(
				req.Context(),
				RequestFailedError{
					Method:     req.Method,
					URL:        req.URL.String(),
					StatusCode: resp.StatusCode,
				},
				"request failed content: %s",
				string(content),
			),
			map[string]string{
				"status_code": strconv.Itoa(resp.StatusCode),
				"status":      resp.Status,
				"method":      req.Method,
				"url":         req.URL.String(),
				"body":        string(content),
			},
		)
	}
	return nil
}
