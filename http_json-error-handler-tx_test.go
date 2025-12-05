// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	liberrors "github.com/bborbe/errors"
	libkv "github.com/bborbe/kv"
	kvmocks "github.com/bborbe/kv/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
	httpmocks "github.com/bborbe/http/mocks"
)

var _ = Describe("JSONUpdateErrorHandler", func() {
	var req *http.Request
	var resp *httptest.ResponseRecorder
	var db *kvmocks.DB
	var tx *kvmocks.Tx
	var subhandler *httpmocks.HttpWithErrorTx

	BeforeEach(func() {
		req = httptest.NewRequest(http.MethodPost, "http://example.com", nil)
		resp = httptest.NewRecorder()
		db = &kvmocks.DB{}
		tx = &kvmocks.Tx{}
		subhandler = &httpmocks.HttpWithErrorTx{}
	})

	JustBeforeEach(func() {
		handler := libhttp.NewJSONUpdateErrorHandler(db, subhandler)
		handler.ServeHTTP(resp, req)
	})

	Context("success", func() {
		BeforeEach(func() {
			db.UpdateCalls(
				func(ctx context.Context, fn func(context.Context, libkv.Tx) error) error {
					return fn(ctx, tx)
				},
			)
			subhandler.ServeHTTPReturns(nil)
		})

		It("calls subhandler", func() {
			Expect(subhandler.ServeHTTPCallCount()).To(Equal(1))
		})

		It("returns 200 OK", func() {
			Expect(resp.Result().StatusCode).To(Equal(http.StatusOK))
		})

		It("calls db.Update", func() {
			Expect(db.UpdateCallCount()).To(Equal(1))
		})
	})

	Context("error from handler", func() {
		BeforeEach(func() {
			db.UpdateCalls(
				func(ctx context.Context, fn func(context.Context, libkv.Tx) error) error {
					return fn(ctx, tx)
				},
			)
			ctx := context.Background()
			originalErr := liberrors.New(ctx, "handler failed")
			wrappedErr := libhttp.WrapWithCode(
				originalErr,
				libhttp.ErrorCodeValidation,
				http.StatusBadRequest,
			)
			subhandler.ServeHTTPReturns(wrappedErr)
		})

		It("returns JSON error response", func() {
			Expect(resp.Result().StatusCode).To(Equal(http.StatusBadRequest))
			Expect(resp.Result().Header.Get("Content-Type")).To(Equal("application/json"))

			body, err := io.ReadAll(resp.Result().Body)
			Expect(err).To(BeNil())

			var errorResp libhttp.ErrorResponse
			err = json.Unmarshal(body, &errorResp)
			Expect(err).To(BeNil())

			Expect(errorResp.Error.Code).To(Equal(libhttp.ErrorCodeValidation))
			Expect(errorResp.Error.Message).To(Equal("handler failed"))
		})
	})

	Context("error from db.Update", func() {
		BeforeEach(func() {
			ctx := context.Background()
			dbErr := liberrors.New(ctx, "database error")
			db.UpdateReturns(dbErr)
		})

		It("returns JSON error response", func() {
			Expect(resp.Result().StatusCode).To(Equal(http.StatusInternalServerError))
			Expect(resp.Result().Header.Get("Content-Type")).To(Equal("application/json"))

			body, err := io.ReadAll(resp.Result().Body)
			Expect(err).To(BeNil())

			var errorResp libhttp.ErrorResponse
			err = json.Unmarshal(body, &errorResp)
			Expect(err).To(BeNil())

			Expect(errorResp.Error.Code).To(Equal(libhttp.ErrorCodeInternal))
			Expect(errorResp.Error.Message).To(Equal("database error"))
		})
	})
})

var _ = Describe("JSONViewErrorHandler", func() {
	var req *http.Request
	var resp *httptest.ResponseRecorder
	var db *kvmocks.DB
	var tx *kvmocks.Tx
	var subhandler *httpmocks.HttpWithErrorTx

	BeforeEach(func() {
		req = httptest.NewRequest(http.MethodGet, "http://example.com", nil)
		resp = httptest.NewRecorder()
		db = &kvmocks.DB{}
		tx = &kvmocks.Tx{}
		subhandler = &httpmocks.HttpWithErrorTx{}
	})

	JustBeforeEach(func() {
		handler := libhttp.NewJSONViewErrorHandler(db, subhandler)
		handler.ServeHTTP(resp, req)
	})

	Context("success", func() {
		BeforeEach(func() {
			db.ViewCalls(func(ctx context.Context, fn func(context.Context, libkv.Tx) error) error {
				return fn(ctx, tx)
			})
			subhandler.ServeHTTPReturns(nil)
		})

		It("calls subhandler", func() {
			Expect(subhandler.ServeHTTPCallCount()).To(Equal(1))
		})

		It("returns 200 OK", func() {
			Expect(resp.Result().StatusCode).To(Equal(http.StatusOK))
		})

		It("calls db.View", func() {
			Expect(db.ViewCallCount()).To(Equal(1))
		})
	})

	Context("error from handler", func() {
		BeforeEach(func() {
			db.ViewCalls(func(ctx context.Context, fn func(context.Context, libkv.Tx) error) error {
				return fn(ctx, tx)
			})
			ctx := context.Background()
			originalErr := liberrors.New(ctx, "not found")
			wrappedErr := libhttp.WrapWithCode(
				originalErr,
				libhttp.ErrorCodeNotFound,
				http.StatusNotFound,
			)
			subhandler.ServeHTTPReturns(wrappedErr)
		})

		It("returns JSON error response", func() {
			Expect(resp.Result().StatusCode).To(Equal(http.StatusNotFound))
			Expect(resp.Result().Header.Get("Content-Type")).To(Equal("application/json"))

			body, err := io.ReadAll(resp.Result().Body)
			Expect(err).To(BeNil())

			var errorResp libhttp.ErrorResponse
			err = json.Unmarshal(body, &errorResp)
			Expect(err).To(BeNil())

			Expect(errorResp.Error.Code).To(Equal(libhttp.ErrorCodeNotFound))
			Expect(errorResp.Error.Message).To(Equal("not found"))
		})
	})

	Context("error from db.View", func() {
		BeforeEach(func() {
			ctx := context.Background()
			dbErr := liberrors.New(ctx, "database error")
			db.ViewReturns(dbErr)
		})

		It("returns JSON error response", func() {
			Expect(resp.Result().StatusCode).To(Equal(http.StatusInternalServerError))
			Expect(resp.Result().Header.Get("Content-Type")).To(Equal("application/json"))

			body, err := io.ReadAll(resp.Result().Body)
			Expect(err).To(BeNil())

			var errorResp libhttp.ErrorResponse
			err = json.Unmarshal(body, &errorResp)
			Expect(err).To(BeNil())

			Expect(errorResp.Error.Code).To(Equal(libhttp.ErrorCodeInternal))
			Expect(errorResp.Error.Message).To(Equal("database error"))
		})
	})
})
