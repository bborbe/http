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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
	"github.com/bborbe/http/mocks"
)

var _ = Describe("JSONErrorHandler", func() {
	var req *http.Request
	var subhandler *mocks.HttpWithError
	var resp *httptest.ResponseRecorder

	BeforeEach(func() {
		req = httptest.NewRequest(http.MethodGet, "http://example.com", nil)
		resp = httptest.NewRecorder()
		subhandler = &mocks.HttpWithError{}
	})

	JustBeforeEach(func() {
		handler := libhttp.NewJSONErrorHandler(subhandler)
		handler.ServeHTTP(resp, req)
	})

	Context("success", func() {
		BeforeEach(func() {
			subhandler.ServeHTTPReturns(nil)
		})

		It("calls subhandler", func() {
			Expect(subhandler.ServeHTTPCallCount()).To(Equal(1))
		})

		It("returns 200 OK", func() {
			Expect(resp.Result().StatusCode).To(Equal(http.StatusOK))
		})

		It("does not write error response", func() {
			body, err := io.ReadAll(resp.Result().Body)
			Expect(err).To(BeNil())
			Expect(body).To(BeEmpty())
		})
	})

	Context("error with status code", func() {
		var originalErr error

		BeforeEach(func() {
			ctx := context.Background()
			originalErr = liberrors.New(ctx, "test error")
			wrappedErr := libhttp.WrapWithStatusCode(originalErr, http.StatusBadRequest)
			subhandler.ServeHTTPReturns(wrappedErr)
		})

		It("calls subhandler", func() {
			Expect(subhandler.ServeHTTPCallCount()).To(Equal(1))
		})

		It("returns correct status code", func() {
			Expect(resp.Result().StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("returns JSON content type", func() {
			Expect(resp.Result().Header.Get("Content-Type")).To(Equal("application/json"))
		})

		It("returns error response with default INTERNAL_ERROR code", func() {
			body, err := io.ReadAll(resp.Result().Body)
			Expect(err).To(BeNil())

			var errorResp libhttp.ErrorResponse
			err = json.Unmarshal(body, &errorResp)
			Expect(err).To(BeNil())

			Expect(errorResp.Error.Code).To(Equal(libhttp.ErrorCodeInternal))
			Expect(errorResp.Error.Message).To(Equal("test error"))
		})
	})

	Context("error with code", func() {
		BeforeEach(func() {
			ctx := context.Background()
			originalErr := liberrors.New(ctx, "validation failed")
			wrappedErr := libhttp.WrapWithCode(
				originalErr,
				libhttp.ErrorCodeValidation,
				http.StatusBadRequest,
			)
			subhandler.ServeHTTPReturns(wrappedErr)
		})

		It("returns correct status code", func() {
			Expect(resp.Result().StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("returns JSON content type", func() {
			Expect(resp.Result().Header.Get("Content-Type")).To(Equal("application/json"))
		})

		It("returns error response with correct code", func() {
			body, err := io.ReadAll(resp.Result().Body)
			Expect(err).To(BeNil())

			var errorResp libhttp.ErrorResponse
			err = json.Unmarshal(body, &errorResp)
			Expect(err).To(BeNil())

			Expect(errorResp.Error.Code).To(Equal(libhttp.ErrorCodeValidation))
			Expect(errorResp.Error.Message).To(Equal("validation failed"))
		})
	})

	Context("error with details", func() {
		BeforeEach(func() {
			ctx := context.Background()
			originalErr := liberrors.New(ctx, "columnGroup '' is unknown")
			details := map[string]string{
				"field":    "columnGroup",
				"expected": "day|week|month|year",
			}
			wrappedErr := libhttp.WrapWithDetails(
				originalErr,
				libhttp.ErrorCodeValidation,
				http.StatusBadRequest,
				details,
			)
			subhandler.ServeHTTPReturns(wrappedErr)
		})

		It("returns correct status code", func() {
			Expect(resp.Result().StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("returns error response with details", func() {
			body, err := io.ReadAll(resp.Result().Body)
			Expect(err).To(BeNil())

			var errorResp libhttp.ErrorResponse
			err = json.Unmarshal(body, &errorResp)
			Expect(err).To(BeNil())

			Expect(errorResp.Error.Code).To(Equal(libhttp.ErrorCodeValidation))
			Expect(errorResp.Error.Message).To(Equal("columnGroup '' is unknown"))
			Expect(errorResp.Error.Details).To(HaveKeyWithValue("field", "columnGroup"))
			Expect(errorResp.Error.Details).To(HaveKeyWithValue("expected", "day|week|month|year"))
		})
	})

	Context("plain error without code or status", func() {
		BeforeEach(func() {
			ctx := context.Background()
			originalErr := liberrors.New(ctx, "something went wrong")
			subhandler.ServeHTTPReturns(originalErr)
		})

		It("returns 500 status code", func() {
			Expect(resp.Result().StatusCode).To(Equal(http.StatusInternalServerError))
		})

		It("returns JSON content type", func() {
			Expect(resp.Result().Header.Get("Content-Type")).To(Equal("application/json"))
		})

		It("returns error response with default INTERNAL_ERROR code", func() {
			body, err := io.ReadAll(resp.Result().Body)
			Expect(err).To(BeNil())

			var errorResp libhttp.ErrorResponse
			err = json.Unmarshal(body, &errorResp)
			Expect(err).To(BeNil())

			Expect(errorResp.Error.Code).To(Equal(libhttp.ErrorCodeInternal))
			Expect(errorResp.Error.Message).To(Equal("something went wrong"))
		})
	})

	Context("all error codes", func() {
		DescribeTable(
			"maps error codes correctly",
			func(code string, expectedCode string) {
				ctx := context.Background()
				originalErr := liberrors.New(ctx, "test error")
				wrappedErr := libhttp.WrapWithCode(originalErr, code, http.StatusBadRequest)
				subhandler.ServeHTTPReturns(wrappedErr)

				handler := libhttp.NewJSONErrorHandler(subhandler)
				handler.ServeHTTP(resp, req)

				body, err := io.ReadAll(resp.Result().Body)
				Expect(err).To(BeNil())

				var errorResp libhttp.ErrorResponse
				err = json.Unmarshal(body, &errorResp)
				Expect(err).To(BeNil())

				Expect(errorResp.Error.Code).To(Equal(expectedCode))
			},
			Entry("validation error", libhttp.ErrorCodeValidation, libhttp.ErrorCodeValidation),
			Entry("not found error", libhttp.ErrorCodeNotFound, libhttp.ErrorCodeNotFound),
			Entry(
				"unauthorized error",
				libhttp.ErrorCodeUnauthorized,
				libhttp.ErrorCodeUnauthorized,
			),
			Entry("forbidden error", libhttp.ErrorCodeForbidden, libhttp.ErrorCodeForbidden),
			Entry("internal error", libhttp.ErrorCodeInternal, libhttp.ErrorCodeInternal),
		)
	})
})
