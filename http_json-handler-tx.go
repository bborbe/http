// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/bborbe/errors"
	libkv "github.com/bborbe/kv"
)

//counterfeiter:generate -o mocks/http-json-handler-tx.go --fake-name HttpJsonHandlerTx . JsonHandlerTx

// JsonHandlerTx defines the interface for handlers that return JSON responses within database transactions.
// Implementations should return the data to be JSON-encoded and any error that occurred.
type JsonHandlerTx interface {
	ServeHTTP(ctx context.Context, tx libkv.Tx, req *http.Request) (interface{}, error)
}

// JsonHandlerTxFunc is an adapter to allow the use of ordinary functions as JsonHandlerTx handlers.
// If f is a function with the appropriate signature, JsonHandlerTxFunc(f) is a JsonHandlerTx that calls f.
type JsonHandlerTxFunc func(ctx context.Context, tx libkv.Tx, req *http.Request) (interface{}, error)

// ServeHTTP calls f(ctx, tx, req).
func (j JsonHandlerTxFunc) ServeHTTP(
	ctx context.Context,
	tx libkv.Tx,
	req *http.Request,
) (interface{}, error) {
	return j(ctx, tx, req)
}

// NewJsonHandlerViewTx wraps a JsonHandlerTx to automatically encode responses as JSON within a read-only database transaction.
// It executes the handler within a database view transaction and handles JSON marshaling.
// Returns a WithError handler that can be used with error handling middleware.
func NewJsonHandlerViewTx(db libkv.DB, jsonHandler JsonHandlerTx) WithError {
	return WithErrorFunc(
		func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
			return db.View(ctx, func(ctx context.Context, tx libkv.Tx) error {
				result, err := jsonHandler.ServeHTTP(ctx, tx, req)
				if err != nil {
					return errors.Wrapf(ctx, err, "json handler failed")
				}
				resp.Header().Add(ContentTypeHeaderName, ApplicationJsonContentType)
				if err := json.NewEncoder(resp).Encode(result); err != nil {
					return errors.Wrapf(ctx, err, "encode json failed")
				}
				return nil
			})
		},
	)
}

// NewJsonHandlerUpdateTx wraps a JsonHandlerTx to automatically encode responses as JSON within a read-write database transaction.
// It executes the handler within a database update transaction and handles JSON marshaling.
// Returns a WithError handler that can be used with error handling middleware.
func NewJsonHandlerUpdateTx(db libkv.DB, jsonHandler JsonHandlerTx) WithError {
	return WithErrorFunc(
		func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
			return db.Update(ctx, func(ctx context.Context, tx libkv.Tx) error {
				result, err := jsonHandler.ServeHTTP(ctx, tx, req)
				if err != nil {
					return errors.Wrapf(ctx, err, "json handler failed")
				}
				resp.Header().Add(ContentTypeHeaderName, ApplicationJsonContentType)
				if err := json.NewEncoder(resp).Encode(result); err != nil {
					return errors.Wrapf(ctx, err, "encode json failed")
				}
				return nil
			})
		},
	)
}
