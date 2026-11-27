// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"context"
	stderrors "errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
	"github.com/bborbe/http/mocks"
)

type retryBackoffResponse struct {
	status  int
	headers map[string]string
	body    string
}

type retryBackoffCase struct {
	url           string
	requestHeader map[string]string
	responses     []retryBackoffResponse
	transportErr  error
	maxAttempts   int
	maxWait       time.Duration
	sleepErr      error
	wantCalls     int
	wantLogs      []string
	wantWait      []time.Duration
	wantErr       bool
	wantErrIs     error
	wantStatus    int
	wantBody      string
}

var retryBackoffNow = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func buildRetryBackoffResponse(fixture retryBackoffResponse) *http.Response {
	header := http.Header{}
	for key, value := range fixture.headers {
		header.Set(key, value)
	}
	return &http.Response{
		StatusCode: fixture.status,
		Header:     header,
		Body:       io.NopCloser(strings.NewReader(fixture.body)),
	}
}

var _ = Describe("RoundTripperRetryWithBackoff", func() {
	DescribeTable(
		"rate-limit handling",
		func(c retryBackoffCase) {
			req, err := http.NewRequest(http.MethodGet, c.url, nil)
			Expect(err).To(BeNil())
			for key, value := range c.requestHeader {
				req.Header.Set(key, value)
			}

			roundTripper := &mocks.HttpRoundTripper{}
			if c.transportErr != nil {
				roundTripper.RoundTripReturns(nil, c.transportErr)
			} else {
				for i := 0; i < c.wantCalls; i++ {
					fixture := c.responses[len(c.responses)-1]
					if i < len(c.responses) {
						fixture = c.responses[i]
					}
					roundTripper.RoundTripReturnsOnCall(i, buildRetryBackoffResponse(fixture), nil)
				}
			}

			var sleeps []time.Duration
			var logs []string
			roundTripperRetry := libhttp.NewRoundTripperRetryWithBackoff(
				roundTripper,
				libhttp.RetryBackoffOptions{
					MaxAttempts: c.maxAttempts,
					MaxWait:     c.maxWait,
					Sleep: func(ctx context.Context, d time.Duration) error {
						sleeps = append(sleeps, d)
						return c.sleepErr
					},
					Now: func() time.Time {
						return retryBackoffNow
					},
					Logf: func(format string, args ...any) {
						logs = append(logs, fmt.Sprintf(format, args...))
					},
				},
			)

			resp, err := roundTripperRetry.RoundTrip(req)

			Expect(roundTripper.RoundTripCallCount()).To(Equal(c.wantCalls))
			Expect(logs).To(Equal(c.wantLogs))
			Expect(sleeps).To(Equal(c.wantWait))
			if c.wantErr {
				Expect(err).NotTo(BeNil())
				if c.wantErrIs != nil {
					Expect(stderrors.Is(err, c.wantErrIs)).To(BeTrue())
				}
				return
			}
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())
			Expect(resp.StatusCode).To(Equal(c.wantStatus))
			if c.wantBody != "" {
				body, readErr := io.ReadAll(resp.Body)
				Expect(readErr).To(BeNil())
				Expect(string(body)).To(Equal(c.wantBody))
			}
		},
		Entry("429 with Retry-After then 200", retryBackoffCase{
			url: "https://api.github.com/repos/x/y",
			responses: []retryBackoffResponse{
				{
					status:  http.StatusTooManyRequests,
					headers: map[string]string{"Retry-After": "2"},
				},
				{status: http.StatusOK},
			},
			wantCalls: 2,
			wantLogs: []string{
				"retry attempt=1 status=429 class=secondary wait=2s",
			},
			wantWait:   []time.Duration{2 * time.Second},
			wantStatus: http.StatusOK,
		}),
		Entry("429 with Retry-After as HTTP date", retryBackoffCase{
			url: "https://api.github.com/repos/x/y",
			responses: []retryBackoffResponse{
				{
					status: http.StatusTooManyRequests,
					headers: map[string]string{
						"Retry-After": retryBackoffNow.Add(30 * time.Second).
							Format(http.TimeFormat),
					},
				},
				{status: http.StatusOK},
			},
			wantCalls: 2,
			wantLogs: []string{
				"retry attempt=1 status=429 class=secondary wait=30s",
			},
			wantWait:   []time.Duration{30 * time.Second},
			wantStatus: http.StatusOK,
		}),
		Entry("403 primary limit uses reset time", retryBackoffCase{
			url: "https://api.github.com/repos/x/y",
			responses: []retryBackoffResponse{
				{
					status: http.StatusForbidden,
					headers: map[string]string{
						"X-RateLimit-Remaining": "0",
						"X-RateLimit-Reset": strconv.FormatInt(
							retryBackoffNow.Add(60*time.Second).Unix(),
							10,
						),
					},
				},
				{status: http.StatusOK},
			},
			wantCalls: 2,
			wantLogs: []string{
				"retry attempt=1 status=403 class=primary wait=1m0s",
			},
			wantWait:   []time.Duration{time.Minute},
			wantStatus: http.StatusOK,
		}),
		Entry("403 with Retry-After beats exhausted remaining", retryBackoffCase{
			url: "https://api.github.com/repos/x/y",
			responses: []retryBackoffResponse{
				{
					status: http.StatusForbidden,
					headers: map[string]string{
						"X-RateLimit-Remaining": "0",
						"Retry-After":           "5",
					},
				},
				{status: http.StatusOK},
			},
			wantCalls: 2,
			wantLogs: []string{
				"retry attempt=1 status=403 class=secondary wait=5s",
			},
			wantWait:   []time.Duration{5 * time.Second},
			wantStatus: http.StatusOK,
		}),
		Entry("429 with remaining 10 is secondary", retryBackoffCase{
			url: "https://api.github.com/repos/x/y",
			responses: []retryBackoffResponse{
				{
					status:  http.StatusTooManyRequests,
					headers: map[string]string{"X-RateLimit-Remaining": "10"},
				},
				{status: http.StatusOK},
			},
			wantCalls: 2,
			wantLogs: []string{
				"retry attempt=1 status=429 class=secondary wait=1s",
			},
			wantWait:   []time.Duration{time.Second},
			wantStatus: http.StatusOK,
		}),
		Entry("429 with negative remaining is unclassified and retried", retryBackoffCase{
			url: "https://api.github.com/repos/x/y",
			responses: []retryBackoffResponse{
				{
					status:  http.StatusTooManyRequests,
					headers: map[string]string{"X-RateLimit-Remaining": "-1"},
				},
				{status: http.StatusOK},
			},
			wantCalls: 2,
			wantLogs: []string{
				"retry attempt=1 status=429 class=unclassified wait=1s",
			},
			wantWait:   []time.Duration{time.Second},
			wantStatus: http.StatusOK,
		}),
		Entry("429 body mentioning secondary rate limit", retryBackoffCase{
			url: "https://api.github.com/repos/x/y",
			responses: []retryBackoffResponse{
				{
					status: http.StatusTooManyRequests,
					body:   "You have exceeded a Secondary Rate Limit",
				},
				{status: http.StatusOK},
			},
			wantCalls: 2,
			wantLogs: []string{
				"retry attempt=1 status=429 class=secondary wait=1s",
			},
			wantWait:   []time.Duration{time.Second},
			wantStatus: http.StatusOK,
		}),
		Entry("plain 403 is not a rate limit", retryBackoffCase{
			url:        "https://api.github.com/repos/x/y",
			responses:  []retryBackoffResponse{{status: http.StatusForbidden}},
			wantCalls:  1,
			wantStatus: http.StatusForbidden,
		}),
		Entry("403 with malformed remaining is not a rate limit", retryBackoffCase{
			url: "https://api.github.com/repos/x/y",
			responses: []retryBackoffResponse{
				{
					status:  http.StatusForbidden,
					headers: map[string]string{"X-RateLimit-Remaining": "abc"},
				},
			},
			wantCalls:  1,
			wantStatus: http.StatusForbidden,
		}),
		Entry("403 with reset in the past waits zero", retryBackoffCase{
			url: "https://api.github.com/repos/x/y",
			responses: []retryBackoffResponse{
				{
					status: http.StatusForbidden,
					headers: map[string]string{
						"X-RateLimit-Remaining": "0",
						"X-RateLimit-Reset": strconv.FormatInt(
							retryBackoffNow.Add(-60*time.Second).Unix(),
							10,
						),
					},
				},
				{status: http.StatusOK},
			},
			wantCalls: 2,
			wantLogs: []string{
				"retry attempt=1 status=403 class=primary wait=0s",
			},
			wantWait:   []time.Duration{0},
			wantStatus: http.StatusOK,
		}),
		Entry("429 with malformed Retry-After falls back to one second", retryBackoffCase{
			url: "https://api.github.com/repos/x/y",
			responses: []retryBackoffResponse{
				{
					status:  http.StatusTooManyRequests,
					headers: map[string]string{"Retry-After": "garbage"},
				},
				{status: http.StatusOK},
			},
			wantCalls: 2,
			wantLogs: []string{
				"retry attempt=1 status=429 class=secondary wait=1s",
			},
			wantWait:   []time.Duration{time.Second},
			wantStatus: http.StatusOK,
		}),
		Entry("200 on first call is not retried", retryBackoffCase{
			url:        "https://api.github.com/repos/x/y",
			responses:  []retryBackoffResponse{{status: http.StatusOK}},
			wantCalls:  1,
			wantStatus: http.StatusOK,
		}),
		Entry("200 body mentioning secondary rate limit is not retried", retryBackoffCase{
			url: "https://api.github.com/repos/x/y",
			responses: []retryBackoffResponse{
				{status: http.StatusOK, body: "secondary rate limit"},
			},
			wantCalls:  1,
			wantStatus: http.StatusOK,
		}),
		Entry("429 from a non-GitHub host is not retried", retryBackoffCase{
			url:        "https://example.com/x",
			responses:  []retryBackoffResponse{{status: http.StatusTooManyRequests}},
			wantCalls:  1,
			wantStatus: http.StatusTooManyRequests,
		}),
		Entry("prevent-retry header disables retries", retryBackoffCase{
			url:           "https://api.github.com/repos/x/y",
			requestHeader: map[string]string{libhttp.PreventRetryHeaderName: "1"},
			responses:     []retryBackoffResponse{{status: http.StatusTooManyRequests}},
			wantCalls:     1,
			wantStatus:    http.StatusTooManyRequests,
		}),
		Entry("persistent 429 exhausts attempts and returns the last response", retryBackoffCase{
			url: "https://api.github.com/repos/x/y",
			responses: []retryBackoffResponse{
				{
					status:  http.StatusTooManyRequests,
					headers: map[string]string{"Retry-After": "1"},
					body:    "original-body",
				},
			},
			maxAttempts: 2,
			wantCalls:   3,
			wantLogs: []string{
				"retry attempt=1 status=429 class=secondary wait=1s",
				"retry attempt=2 status=429 class=secondary wait=1s",
				"retry exhausted status=429 retries=2",
			},
			wantWait:   []time.Duration{time.Second, time.Second},
			wantStatus: http.StatusTooManyRequests,
			wantBody:   "original-body",
		}),
		Entry("primary limit an hour away is deferred", retryBackoffCase{
			url: "https://api.github.com/repos/x/y",
			responses: []retryBackoffResponse{
				{
					status: http.StatusForbidden,
					headers: map[string]string{
						"X-RateLimit-Remaining": "0",
						"X-RateLimit-Reset": strconv.FormatInt(
							retryBackoffNow.Add(time.Hour).Unix(),
							10,
						),
					},
					body: "original-body",
				},
			},
			wantCalls: 1,
			wantLogs: []string{
				"retry deferred wait=1h0m0s",
			},
			wantWait:   nil,
			wantStatus: http.StatusForbidden,
			wantBody:   "original-body",
		}),
		Entry("wait exactly at the cap is retried", retryBackoffCase{
			url: "https://api.github.com/repos/x/y",
			responses: []retryBackoffResponse{
				{
					status: http.StatusForbidden,
					headers: map[string]string{
						"X-RateLimit-Remaining": "0",
						"X-RateLimit-Reset": strconv.FormatInt(
							retryBackoffNow.Add(time.Minute).Unix(),
							10,
						),
					},
				},
				{status: http.StatusOK},
			},
			maxWait:   time.Minute,
			wantCalls: 2,
			wantLogs: []string{
				"retry attempt=1 status=403 class=primary wait=1m0s",
			},
			wantWait:   []time.Duration{time.Minute},
			wantStatus: http.StatusOK,
		}),
		Entry("wait one second over the cap is deferred", retryBackoffCase{
			url: "https://api.github.com/repos/x/y",
			responses: []retryBackoffResponse{
				{
					status: http.StatusForbidden,
					headers: map[string]string{
						"X-RateLimit-Remaining": "0",
						"X-RateLimit-Reset": strconv.FormatInt(
							retryBackoffNow.Add(time.Minute+time.Second).Unix(),
							10,
						),
					},
				},
			},
			maxWait:   time.Minute,
			wantCalls: 1,
			wantLogs: []string{
				"retry deferred wait=1m1s",
			},
			wantWait:   nil,
			wantStatus: http.StatusForbidden,
		}),
		Entry(
			"zero maxWait defaults to five minutes and defers a six-minute reset",
			retryBackoffCase{
				url: "https://api.github.com/repos/x/y",
				responses: []retryBackoffResponse{
					{
						status: http.StatusForbidden,
						headers: map[string]string{
							"X-RateLimit-Remaining": "0",
							"X-RateLimit-Reset": strconv.FormatInt(
								retryBackoffNow.Add(6*time.Minute).Unix(),
								10,
							),
						},
					},
				},
				wantCalls: 1,
				wantLogs: []string{
					"retry deferred wait=6m0s",
				},
				wantWait:   nil,
				wantStatus: http.StatusForbidden,
			},
		),
		Entry(
			"zero maxWait defaults to five minutes and retries a four-minute reset",
			retryBackoffCase{
				url: "https://api.github.com/repos/x/y",
				responses: []retryBackoffResponse{
					{
						status: http.StatusForbidden,
						headers: map[string]string{
							"X-RateLimit-Remaining": "0",
							"X-RateLimit-Reset": strconv.FormatInt(
								retryBackoffNow.Add(4*time.Minute).Unix(),
								10,
							),
						},
					},
					{status: http.StatusOK},
				},
				wantCalls: 2,
				wantLogs: []string{
					"retry attempt=1 status=403 class=primary wait=4m0s",
				},
				wantWait:   []time.Duration{4 * time.Minute},
				wantStatus: http.StatusOK,
			},
		),
		Entry(
			"negative maxWait defaults to five minutes and defers a six-minute reset",
			retryBackoffCase{
				url: "https://api.github.com/repos/x/y",
				responses: []retryBackoffResponse{
					{
						status: http.StatusForbidden,
						headers: map[string]string{
							"X-RateLimit-Remaining": "0",
							"X-RateLimit-Reset": strconv.FormatInt(
								retryBackoffNow.Add(6*time.Minute).Unix(),
								10,
							),
						},
					},
				},
				maxWait:   -time.Minute,
				wantCalls: 1,
				wantLogs: []string{
					"retry deferred wait=6m0s",
				},
				wantWait:   nil,
				wantStatus: http.StatusForbidden,
			},
		),
		Entry("transport error is returned immediately", retryBackoffCase{
			url:          "https://api.github.com/repos/x/y",
			transportErr: io.EOF,
			wantCalls:    1,
			wantErr:      true,
		}),
		Entry("context cancelled during the wait returns the wrapped error", retryBackoffCase{
			url: "https://api.github.com/repos/x/y",
			responses: []retryBackoffResponse{
				{
					status:  http.StatusTooManyRequests,
					headers: map[string]string{"Retry-After": "2"},
				},
			},
			sleepErr:  context.Canceled,
			wantCalls: 1,
			wantLogs: []string{
				"retry attempt=1 status=429 class=secondary wait=2s",
			},
			wantWait:  []time.Duration{2 * time.Second},
			wantErr:   true,
			wantErrIs: context.Canceled,
		}),
	)

	It("replays the request body on retry", func() {
		req, err := http.NewRequest(
			http.MethodPost,
			"https://api.github.com/repos/x/y",
			strings.NewReader("payload"),
		)
		Expect(err).To(BeNil())

		var bodies []string
		roundTripper := &mocks.HttpRoundTripper{}
		roundTripper.RoundTripCalls(func(req *http.Request) (*http.Response, error) {
			body := ""
			if req.Body != nil {
				read, readErr := io.ReadAll(req.Body)
				Expect(readErr).To(BeNil())
				body = string(read)
			}
			bodies = append(bodies, body)
			if len(bodies) == 1 {
				return buildRetryBackoffResponse(retryBackoffResponse{
					status:  http.StatusTooManyRequests,
					headers: map[string]string{"Retry-After": "1"},
				}), nil
			}
			return buildRetryBackoffResponse(retryBackoffResponse{status: http.StatusOK}), nil
		})

		roundTripperRetry := libhttp.NewRoundTripperRetryWithBackoff(
			roundTripper,
			libhttp.RetryBackoffOptions{
				Sleep: func(ctx context.Context, d time.Duration) error {
					return nil
				},
				Now: func() time.Time {
					return retryBackoffNow
				},
				Logf: func(format string, args ...any) {},
			},
		)

		resp, err := roundTripperRetry.RoundTrip(req)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		Expect(bodies).To(Equal([]string{"payload", "payload"}))
	})

	DescribeTable("RateLimitClass String round-trips",
		func(class libhttp.RateLimitClass) {
			Expect(libhttp.RateLimitClass(class.String())).To(Equal(class))
		},
		Entry("none", libhttp.RateLimitClassNone),
		Entry("primary", libhttp.RateLimitClassPrimary),
		Entry("secondary", libhttp.RateLimitClassSecondary),
		Entry("unclassified", libhttp.RateLimitClassUnclassified),
	)
})
