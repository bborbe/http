// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
)

type classifyRateLimitCase struct {
	status  int
	headers map[string]string
	body    []byte
}

var _ = Describe("ClassifyRateLimitResponse", func() {
	DescribeTable("classifies",
		func(c classifyRateLimitCase, expected libhttp.RateLimitClass) {
			resp := &http.Response{
				StatusCode: c.status,
				Header:     http.Header{},
			}
			for key, value := range c.headers {
				resp.Header.Set(key, value)
			}
			Expect(libhttp.ClassifyRateLimitResponse(resp, c.body)).To(Equal(expected))
		},
		Entry("429 with positive remaining is secondary",
			classifyRateLimitCase{
				status:  http.StatusTooManyRequests,
				headers: map[string]string{"X-RateLimit-Remaining": "12"},
			},
			libhttp.RateLimitClassSecondary,
		),
		Entry("429 with exhausted remaining is primary",
			classifyRateLimitCase{
				status:  http.StatusTooManyRequests,
				headers: map[string]string{"X-RateLimit-Remaining": "0"},
			},
			libhttp.RateLimitClassPrimary,
		),
		Entry("429 without remaining or retry-after is unclassified",
			classifyRateLimitCase{status: http.StatusTooManyRequests},
			libhttp.RateLimitClassUnclassified,
		),
		Entry("403 with positive remaining is none",
			classifyRateLimitCase{
				status:  http.StatusForbidden,
				headers: map[string]string{"X-RateLimit-Remaining": "12"},
			},
			libhttp.RateLimitClassNone,
		),
		Entry("429 with malformed remaining is unclassified",
			classifyRateLimitCase{
				status:  http.StatusTooManyRequests,
				headers: map[string]string{"X-RateLimit-Remaining": "abc"},
			},
			libhttp.RateLimitClassUnclassified,
		),
		Entry("429 with retry-after is secondary",
			classifyRateLimitCase{
				status:  http.StatusTooManyRequests,
				headers: map[string]string{"Retry-After": "30"},
			},
			libhttp.RateLimitClassSecondary,
		),
	)
})
