// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"net/http"
	"os"
	"runtime/pprof"
)

// NewCPUProfileStartHandler creates a handler that starts CPU profiling.
// The profile is written to cpu.pprof in the current directory.
func NewCPUProfileStartHandler() WithError {
	return WithErrorFunc(
		func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
			f, err := os.Create("cpu.pprof")
			if err != nil {
				return err
			}
			return pprof.StartCPUProfile(f)
		},
	)
}

// NewCpuProfileStartHandler is deprecated. Use NewCPUProfileStartHandler instead.
//
// Deprecated: Use NewCPUProfileStartHandler for correct Go naming conventions.
//
//nolint:revive
func NewCpuProfileStartHandler() WithError {
	return NewCPUProfileStartHandler()
}

// NewCPUProfileStopHandler creates a handler that stops CPU profiling.
func NewCPUProfileStopHandler() WithError {
	return WithErrorFunc(
		func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
			pprof.StopCPUProfile()
			return nil
		},
	)
}

// NewCpuProfileStopHandler is deprecated. Use NewCPUProfileStopHandler instead.
//
// Deprecated: Use NewCPUProfileStopHandler for correct Go naming conventions.
//
//nolint:revive
func NewCpuProfileStopHandler() WithError {
	return NewCPUProfileStopHandler()
}
