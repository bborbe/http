// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"net/http"

	"github.com/bborbe/run"
	"github.com/golang/glog"
)

// NewBackgroundRunHandler creates an HTTP handler that executes a run.Func in the background.
// When the endpoint is called, it triggers the runFunc asynchronously and immediately returns a response.
// The handler uses a ParallelSkipper to prevent multiple concurrent executions of the same function.
func NewBackgroundRunHandler(ctx context.Context, runFunc run.Func) http.Handler {
	parallelSkipper := run.NewParallelSkipper()
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		go func() {
			action := parallelSkipper.SkipParallel(func(ctx context.Context) error {
				if err := runFunc(ctx); err != nil {
					return err
				}
				return nil
			})
			glog.V(2).Infof("run started")
			if err := action(ctx); err != nil {
				glog.V(1).Infof("run failed: %v", err)
			}
			glog.V(2).Infof("run completed")
		}()
		_, _ = WriteAndGlog(resp, "run triggered. Check logs for progress.")
	})
}
