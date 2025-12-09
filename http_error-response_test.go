// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"context"
	"encoding/json"
	"net/http"

	liberrors "github.com/bborbe/errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
)

var _ = Describe("ErrorResponse", func() {
	Context("JSON marshaling", func() {
		It("marshals complete error response", func() {
			errorResp := libhttp.ErrorResponse{
				Error: libhttp.ErrorDetails{
					Code:    libhttp.ErrorCodeValidation,
					Message: "test error",
					Details: map[string]any{"field": "columnGroup"},
				},
			}

			data, err := json.Marshal(errorResp)
			Expect(err).To(BeNil())

			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			Expect(err).To(BeNil())

			errorObj, ok := result["error"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(errorObj["code"]).To(Equal(libhttp.ErrorCodeValidation))
			Expect(errorObj["message"]).To(Equal("test error"))
			Expect(errorObj["details"]).NotTo(BeNil())
		})

		It("omits empty details field", func() {
			errorResp := libhttp.ErrorResponse{
				Error: libhttp.ErrorDetails{
					Code:    libhttp.ErrorCodeInternal,
					Message: "test error",
				},
			}

			data, err := json.Marshal(errorResp)
			Expect(err).To(BeNil())

			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			Expect(err).To(BeNil())

			errorObj, ok := result["error"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			_, hasDetails := errorObj["details"]
			Expect(hasDetails).To(BeFalse())
		})
	})
})

var _ = Describe("WrapWithCode", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	It("wraps error with code and status", func() {
		originalErr := liberrors.New(ctx, "test error")
		wrappedErr := libhttp.WrapWithCode(
			originalErr,
			libhttp.ErrorCodeValidation,
			http.StatusBadRequest,
		)

		Expect(wrappedErr.Error()).To(Equal("test error"))

		var errorWithCode libhttp.ErrorWithCode
		Expect(liberrors.As(wrappedErr, &errorWithCode)).To(BeTrue())
		Expect(errorWithCode.Code()).To(Equal(libhttp.ErrorCodeValidation))

		var errorWithStatusCode libhttp.ErrorWithStatusCode
		Expect(liberrors.As(wrappedErr, &errorWithStatusCode)).To(BeTrue())
		Expect(errorWithStatusCode.StatusCode()).To(Equal(http.StatusBadRequest))
	})

	It("supports error unwrapping", func() {
		originalErr := liberrors.New(ctx, "test error")
		wrappedErr := libhttp.WrapWithCode(
			originalErr,
			libhttp.ErrorCodeValidation,
			http.StatusBadRequest,
		)

		Expect(liberrors.Is(wrappedErr, originalErr)).To(BeTrue())
	})
})

var _ = Describe("WrapWithDetails", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	It("wraps error with code, status, and details", func() {
		originalErr := liberrors.New(ctx, "test error")
		details := map[string]any{
			"field":    "columnGroup",
			"expected": "day|week|month|year",
		}
		wrappedErr := libhttp.WrapWithDetails(
			originalErr,
			libhttp.ErrorCodeValidation,
			http.StatusBadRequest,
			details,
		)

		Expect(wrappedErr.Error()).To(Equal("test error"))

		var errorWithCode libhttp.ErrorWithCode
		Expect(liberrors.As(wrappedErr, &errorWithCode)).To(BeTrue())
		Expect(errorWithCode.Code()).To(Equal(libhttp.ErrorCodeValidation))

		var errorWithStatusCode libhttp.ErrorWithStatusCode
		Expect(liberrors.As(wrappedErr, &errorWithStatusCode)).To(BeTrue())
		Expect(errorWithStatusCode.StatusCode()).To(Equal(http.StatusBadRequest))

		extractedDetails := liberrors.DataFromError(wrappedErr)
		Expect(extractedDetails).To(HaveKeyWithValue("field", "columnGroup"))
		Expect(extractedDetails).To(HaveKeyWithValue("expected", "day|week|month|year"))
	})

	It("works with empty details map", func() {
		originalErr := liberrors.New(ctx, "test error")
		details := map[string]any{}
		wrappedErr := libhttp.WrapWithDetails(
			originalErr,
			libhttp.ErrorCodeValidation,
			http.StatusBadRequest,
			details,
		)

		Expect(wrappedErr.Error()).To(Equal("test error"))

		extractedDetails := liberrors.DataFromError(wrappedErr)
		Expect(extractedDetails).To(BeEmpty())
	})
})
