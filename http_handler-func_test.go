// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
)

var _ = Describe("HandlerFunc", func() {
	Context("type alias", func() {
		It("can be used as http.HandlerFunc", func() {
			var handlerFunc libhttp.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("test"))
			}

			req := httptest.NewRequest("GET", "/", nil)
			resp := httptest.NewRecorder()

			http.HandlerFunc(handlerFunc).ServeHTTP(resp, req)

			Expect(resp.Code).To(Equal(http.StatusOK))
			Expect(resp.Body.String()).To(Equal("test"))
		})
	})
})
