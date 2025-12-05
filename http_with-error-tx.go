// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"net/http"

	libkv "github.com/bborbe/kv"
)

//counterfeiter:generate -o mocks/http-with-error-tx.go --fake-name HttpWithErrorTx . WithErrorTx

// WithErrorTx defines the interface for HTTP handlers that can return errors and work with database transactions.
// This extends the WithError interface by providing access to a key-value transaction for database operations.
type WithErrorTx interface {
	ServeHTTP(ctx context.Context, tx libkv.Tx, resp http.ResponseWriter, req *http.Request) error
}

// WithErrorTxFunc is an adapter to allow the use of ordinary functions as WithErrorTx handlers.
// If f is a function with the appropriate signature, WithErrorTxFunc(f) is a WithErrorTx that calls f.
type WithErrorTxFunc func(ctx context.Context, tx libkv.Tx, resp http.ResponseWriter, req *http.Request) error

// ServeHTTP calls f(ctx, tx, resp, req).
func (w WithErrorTxFunc) ServeHTTP(
	ctx context.Context,
	tx libkv.Tx,
	resp http.ResponseWriter,
	req *http.Request,
) error {
	return w(ctx, tx, resp, req)
}
