// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/bborbe/errors"
	"github.com/bborbe/run"
	"github.com/golang/glog"
)

func NewServerWithPort(port int, router http.Handler) run.Func {
	return NewServer(
		fmt.Sprintf(":%d", port),
		router,
	)
}

func NewServer(addr string, router http.Handler) run.Func {
	return newServerTLS(addr, router, nil)
}

func NewServerTLS(addr string, router http.Handler, serverCertPath string, serverKeyPath string) run.Func {
	return func(ctx context.Context) error {
		cert, err := tls.LoadX509KeyPair(serverCertPath, serverKeyPath)
		if err != nil {
			return errors.Wrapf(ctx, err, "load server certificate and key failed")
		}
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		return newServerTLS(addr, router, tlsConfig).Run(ctx)
	}
}

func newServerTLS(addr string, router http.Handler, tlsConfig *tls.Config) run.Func {
	return func(ctx context.Context) error {

		server := &http.Server{
			Addr:      addr,
			Handler:   router,
			TLSConfig: tlsConfig,
		}
		go func() {
			select {
			case <-ctx.Done():
				if err := server.Shutdown(ctx); err != nil {
					glog.Warningf("shutdown failed: %v", err)
				}
			}
		}()
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			glog.V(0).Info(err)
			return nil
		}
		return errors.Wrapf(ctx, err, "httpServer failed")
	}
}
