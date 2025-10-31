// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bborbe/errors"
	libtime "github.com/bborbe/time"
	"github.com/golang/glog"
)

// NewDangerousHandlerWrapper wraps dangerous HTTP handlers with passphrase protection.
// Each instance generates a unique passphrase that expires after 5 minutes.
// The passphrase is logged to stdout/stderr, requiring operators to have log access
// in addition to HTTP access to execute dangerous operations.
func NewDangerousHandlerWrapper(handler http.Handler) http.Handler {
	return NewDangerousHandlerWrapperWithCurrentDateTime(handler, libtime.NewCurrentDateTime())
}

// NewDangerousHandlerWrapperWithCurrentDateTime wraps dangerous HTTP handlers with passphrase protection
// using the provided CurrentDateTime interface for testability.
func NewDangerousHandlerWrapperWithCurrentDateTime(
	handler http.Handler,
	currentDateTime libtime.CurrentDateTime,
) http.Handler {
	return &dangerousHandlerWrapper{
		handler:         handler,
		currentDateTime: currentDateTime,
	}
}

type dangerousHandlerWrapper struct {
	handler         http.Handler
	currentDateTime libtime.CurrentDateTime
	mu              sync.Mutex
	passphrase      string
	expiry          libtime.DateTime
}

// ServeHTTP implements http.Handler.
// It checks for a valid passphrase in the query parameter before executing the wrapped handler.
func (w *dangerousHandlerWrapper) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	// Get current passphrase (generate new if expired)
	_ = w.getCurrentPassphrase(path)

	// Check provided passphrase
	providedPassphrase := req.URL.Query().Get("passphrase")

	if providedPassphrase == "" {
		http.Error(resp,
			fmt.Sprintf(
				"⚠️  DANGEROUS OPERATION REQUIRES PASSPHRASE\n\n"+
					"This endpoint (%s) requires a passphrase for security.\n"+
					"Check service logs for the current passphrase, then retry:\n\n"+
					"  %s?passphrase=YOUR_PASSPHRASE\n\n"+
					"The passphrase expires after 5 minutes and changes on each expiry.",
				path, path,
			),
			http.StatusForbidden)
		return
	}

	// Verify passphrase
	w.mu.Lock()
	valid := providedPassphrase == w.passphrase && w.currentDateTime.Now().Before(w.expiry)
	w.mu.Unlock()

	if !valid {
		glog.Warningf("Invalid or expired passphrase attempt for %s", path)
		http.Error(resp,
			"⚠️  INVALID OR EXPIRED PASSPHRASE\n\n"+
				"The passphrase is incorrect or has expired.\n"+
				"Check service logs for the current passphrase.",
			http.StatusForbidden)
		return
	}

	// Passphrase valid - log and execute
	glog.Warningf("⚠️  Executing dangerous operation: %s (passphrase verified)", path)

	// Execute wrapped handler
	w.handler.ServeHTTP(resp, req)
}

// getCurrentPassphrase returns the current passphrase, generating a new one if expired.
func (w *dangerousHandlerWrapper) getCurrentPassphrase(path string) string {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := w.currentDateTime.Now()

	// Check if current passphrase is still valid
	if now.Before(w.expiry) {
		return w.passphrase
	}

	// Generate new passphrase
	passphrase, err := generatePassphrase(12) // 12 bytes = 16 characters base64url
	if err != nil {
		glog.Errorf("Failed to generate passphrase: %v", err)
		// Fallback to timestamp-based passphrase
		passphrase = fmt.Sprintf("fallback-%d", now.Unix())
	}

	w.passphrase = passphrase
	w.expiry = now.Add(libtime.Duration(5 * time.Minute))

	// Log the new passphrase
	glog.Warningf("⚠️  DANGER PASSPHRASE for %s: %s (expires: %s)",
		path, w.passphrase, w.expiry.Time().Format(time.RFC3339))

	return w.passphrase
}

// generatePassphrase generates a cryptographically secure random passphrase.
// numBytes specifies how many random bytes to generate (will be base64url encoded).
func generatePassphrase(numBytes int) (string, error) {
	b := make([]byte, numBytes)
	_, err := rand.Read(b)
	if err != nil {
		return "", errors.Wrap(context.TODO(), err, "read random bytes failed")
	}
	// Use URL encoding to avoid "/" and "+" characters which are problematic in URLs
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b), nil
}
