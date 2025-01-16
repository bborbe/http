// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	libhttp "github.com/bborbe/http"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Background Run Handler", func() {
	var h http.Handler
	var request *http.Request
	var err error
	var counter int
	var ctx context.Context
	BeforeEach(func() {
		ctx = context.Background()
		request, err = http.NewRequest(http.MethodGet, "http://example.com", nil)
		Expect(err).To(BeNil())
		counter = 0
	})
	Context("single call", func() {
		var response *httptest.ResponseRecorder
		BeforeEach(func() {
			response = httptest.NewRecorder()
			var wg sync.WaitGroup
			wg.Add(1)
			h = libhttp.NewBackgroundRunHandler(ctx, func(ctx context.Context) error {
				defer wg.Done()
				counter++
				return nil
			})
			h.ServeHTTP(response, request)
			wg.Wait()
		})
		It("calls func", func() {
			Expect(counter).To(Equal(1))
		})
	})
	Context("concurrent call", func() {
		var response1 *httptest.ResponseRecorder
		var response2 *httptest.ResponseRecorder
		BeforeEach(func() {
			response1 = httptest.NewRecorder()
			response2 = httptest.NewRecorder()
			var wg sync.WaitGroup
			wg.Add(1)
			h = libhttp.NewBackgroundRunHandler(ctx, func(ctx context.Context) error {
				defer wg.Done()
				time.Sleep(100 * time.Millisecond)
				counter++
				return nil
			})
			h.ServeHTTP(response1, request)
			h.ServeHTTP(response2, request)
			wg.Wait()
		})
		It("calls func", func() {
			Expect(counter).To(Equal(1))
		})
	})
})
