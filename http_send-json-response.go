// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/bborbe/errors"
)

// SendJSONResponse writes a JSON-encoded response with the specified status code.
// It sets the Content-Type header to application/json and returns any encoding errors.
func SendJSONResponse(
	ctx context.Context,
	resp http.ResponseWriter,
	data interface{},
	statusCode int,
) error {
	resp.Header().Set(ContentTypeHeaderName, ApplicationJsonContentType)
	resp.WriteHeader(statusCode)
	if err := json.NewEncoder(resp).Encode(data); err != nil {
		return errors.Wrap(ctx, err, "encode json response failed")
	}
	return nil
}

// ValidateFilename checks if a filename is safe for use in Content-Disposition header.
// It returns an error if the filename:
//   - Is empty
//   - Contains path separators (/, \) or path traversal (..)
//   - Starts with a slash
//   - Contains control characters (ASCII 0-31, 127) that could enable header injection
//   - Contains quotes that could break the header format
//   - Is not clean according to filepath.Clean (catches sneaky path tricks)
//
// This function expects a bare filename, not a path. If you have a path and want
// to extract the filename, use filepath.Base() first.
func ValidateFilename(ctx context.Context, name string) error {
	// 1. Reject empty names
	if name == "" {
		return errors.Errorf(ctx, "filename cannot be empty")
	}

	// 2. Reject absolute or relative paths
	if strings.Contains(name, "..") {
		return errors.Errorf(ctx, "filename contains '..'")
	}
	if strings.HasPrefix(name, "/") || strings.HasPrefix(name, "\\") {
		return errors.Errorf(ctx, "filename starts with a slash")
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return errors.Errorf(ctx, "filename contains a path separator")
	}

	// 3. Normalize and compare to ensure no sneaky path tricks
	clean := filepath.Clean(name)
	if clean != name {
		return errors.Errorf(ctx, "filename is not clean: %q vs %q", name, clean)
	}

	// 4. Check for control characters (prevents HTTP header injection)
	for i, r := range name {
		if r < 32 || r == 127 {
			return errors.Errorf(ctx, "filename contains control character at position %d", i)
		}
	}

	// 5. Check for quotes (prevents breaking Content-Disposition header format)
	if strings.ContainsAny(name, `"`) {
		return errors.Errorf(ctx, "filename contains quotes")
	}

	return nil
}

// SendJSONFileResponse writes a JSON-encoded response configured for file download.
// It sets Content-Type to application/json, adds a Content-Disposition header with
// the specified filename, and includes Content-Length for the response.
//
// The filename must be a valid, safe filename (not a path). It will be validated
// and an error returned if it contains:
//   - Path separators or traversal sequences (/, \, ..)
//   - Control characters that could enable HTTP header injection
//   - Quotes that could break the header format
//
// If you have a path and want to use just the filename, call filepath.Base() first.
//
// Note: This function uses json.Marshal to buffer the entire response in memory,
// which is necessary to calculate and set the Content-Length header. For very large
// responses (>100MB), consider using streaming alternatives without Content-Length.
//
// Example:
//
//	data := map[string]interface{}{"users": users}
//	err := http.SendJSONFileResponse(ctx, w, data, "users.json", http.StatusOK)
func SendJSONFileResponse(
	ctx context.Context,
	resp http.ResponseWriter,
	data interface{},
	fileName string,
	statusCode int,
) error {
	// Validate filename to prevent header injection and path traversal
	if err := ValidateFilename(ctx, fileName); err != nil {
		return errors.Wrap(ctx, err, "invalid filename")
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(ctx, err, "marshal data to bytes failed")
	}
	resp.Header().Set(ContentTypeHeaderName, ApplicationJsonContentType)
	resp.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	resp.Header().Set("Content-Length", fmt.Sprintf("%d", len(dataBytes)))
	resp.WriteHeader(statusCode)
	if _, err := resp.Write(dataBytes); err != nil {
		return errors.Wrap(ctx, err, "write bytes failed")
	}
	return nil
}
