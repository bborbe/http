// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"net/http"
	"os"
	"runtime/pprof"

	"github.com/bborbe/errors"
)

// NewMemoryProfileHandler creates a handler that writes a memory profile to a local file.
// The profile is saved to "memprofile.pprof" in the current working directory.
// On success, it writes a confirmation message to the HTTP response.
func NewMemoryProfileHandler() WithError {
	return WithErrorFunc(
		func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
			memoryFile, err := os.Create("memprofile.pprof")
			if err != nil {
				return errors.Wrap(ctx, err, "create memory profile file failed")
			}
			defer memoryFile.Close()

			if err := pprof.WriteHeapProfile(memoryFile); err != nil {
				return errors.Wrap(ctx, err, "write heap profile failed")
			}

			_, _ = WriteAndGlog(resp, "Memory profile written to memprofile.pprof")
			return nil
		},
	)
}

// NewMemoryProfileDownloadHandler creates a handler that generates a memory profile
// and sends it as a downloadable file to the client.
// The profile is streamed directly to the HTTP response without buffering in memory
// to avoid additional memory pressure on the service.
func NewMemoryProfileDownloadHandler() WithError {
	return WithErrorFunc(
		func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
			resp.Header().Set("Content-Type", "application/octet-stream")
			resp.Header().Set("Content-Disposition", "attachment; filename=memprofile.pprof")

			if err := pprof.WriteHeapProfile(resp); err != nil {
				return errors.Wrap(ctx, err, "write heap profile failed")
			}

			return nil
		},
	)
}
