// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//nolint:dupl // JSON variant mirrors plain text variant intentionally
package http

import (
	"context"
	"net/http"

	libkv "github.com/bborbe/kv"
)

// NewJSONUpdateErrorHandler wraps a WithErrorTx handler for update transactions,
// returning JSON error responses instead of plain text.
//
// Example usage:
//
//	handler := libhttp.NewJSONUpdateErrorHandler(
//	    db,
//	    libhttp.WithErrorTxFunc(func(ctx context.Context, tx libkv.Tx, resp http.ResponseWriter, req *http.Request) error {
//	        // Handle update logic with transaction
//	        return nil
//	    }),
//	)
func NewJSONUpdateErrorHandler(db libkv.DB, withErrorTx WithErrorTx) http.Handler {
	return NewJSONErrorHandler(
		WithErrorFunc(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
			return db.Update(ctx, func(ctx context.Context, tx libkv.Tx) error {
				return withErrorTx.ServeHTTP(ctx, tx, resp, req)
			})
		}),
	)
}

// NewJSONViewErrorHandler wraps a WithErrorTx handler for view (read-only) transactions,
// returning JSON error responses instead of plain text.
//
// Example usage:
//
//	handler := libhttp.NewJSONViewErrorHandler(
//	    db,
//	    libhttp.WithErrorTxFunc(func(ctx context.Context, tx libkv.Tx, resp http.ResponseWriter, req *http.Request) error {
//	        // Handle read-only logic with transaction
//	        return nil
//	    }),
//	)
func NewJSONViewErrorHandler(db libkv.DB, withErrorTx WithErrorTx) http.Handler {
	return NewJSONErrorHandler(
		WithErrorFunc(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
			return db.View(ctx, func(ctx context.Context, tx libkv.Tx) error {
				return withErrorTx.ServeHTTP(ctx, tx, resp, req)
			})
		}),
	)
}
