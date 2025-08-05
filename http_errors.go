// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"errors"
	"io"
	"net/http"
	"syscall"

	liberrors "github.com/bborbe/errors"
)

// HasTimeoutError defines an interface for errors that can indicate timeout conditions.
// Errors implementing this interface can be checked for timeout status.
type HasTimeoutError interface {
	Timeout() bool
}

// IsRetryError determines whether an error should trigger a retry attempt.
// It checks for common transient errors like EOF, connection refused, timeouts, and handler timeouts.
// Returns true if the error is considered retryable, false otherwise.
func IsRetryError(err error) bool {
	if errors.Is(err, io.EOF) {
		return true
	}
	if errors.Is(err, syscall.ECONNREFUSED) {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if errors.Is(err, http.ErrHandlerTimeout) {
		return true
	}
	if timeoutError, ok := liberrors.Unwrap(err).(HasTimeoutError); ok {
		return timeoutError.Timeout()
	}
	return false
}

// HasTemporaryError defines an interface for errors that can indicate temporary conditions.
// Errors implementing this interface can be checked for temporary failure status.
type HasTemporaryError interface {
	Temporary() bool
}
