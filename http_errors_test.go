// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"context"
	"net/http"
	"syscall"

	"github.com/bborbe/errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
)

var _ = Describe("IsRetryError", func() {
	var ctx context.Context
	var err error
	var isRetryError bool
	BeforeEach(func() {
		ctx = context.Background()
	})
	Context("IsRetryError", func() {
		JustBeforeEach(func() {
			isRetryError = libhttp.IsRetryError(err)
		})
		Context("any error", func() {
			BeforeEach(func() {
				err = errors.New(ctx, "banana")
			})
			It("returns false", func() {
				Expect(isRetryError).To(BeFalse())
			})
		})
		Context("context context.DeadlineExceeded error", func() {
			BeforeEach(func() {
				err = context.DeadlineExceeded
			})
			It("returns true", func() {
				Expect(isRetryError).To(BeTrue())
			})
		})
		Context("wrapped context.DeadlineExceeded error", func() {
			BeforeEach(func() {
				err = errors.Wrapf(ctx, context.DeadlineExceeded, "banana")
			})
			It("returns true", func() {
				Expect(isRetryError).To(BeTrue())
			})
		})
		Context("context http.ErrHandlerTimeout error", func() {
			BeforeEach(func() {
				err = context.DeadlineExceeded
			})
			It("returns true", func() {
				Expect(isRetryError).To(BeTrue())
			})
		})
		Context("wrapped http.ErrHandlerTimeout error", func() {
			BeforeEach(func() {
				err = errors.Wrapf(ctx, http.ErrHandlerTimeout, "banana")
			})
			It("returns true", func() {
				Expect(isRetryError).To(BeTrue())
			})
		})
		Context("context syscall.ECONNREFUSED error", func() {
			BeforeEach(func() {
				err = syscall.ECONNREFUSED
			})
			It("returns true", func() {
				Expect(isRetryError).To(BeTrue())
			})
		})
		Context("wrapped syscall.ECONNREFUSED error", func() {
			BeforeEach(func() {
				err = errors.Wrapf(ctx, syscall.ECONNREFUSED, "banana")
			})
			It("returns true", func() {
				Expect(isRetryError).To(BeTrue())
			})
		})
	})
})
