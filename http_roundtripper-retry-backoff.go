// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bborbe/errors"
	libtime "github.com/bborbe/time"
	"github.com/golang/glog"
)

// RetryBackoffOptions configures NewRoundTripperRetryWithBackoff.
// Every zero-value field falls back to a sensible default.
type RetryBackoffOptions struct {
	// MaxAttempts is the number of retries allowed after the first attempt.
	// A value of zero uses the default of 5; a negative value is treated as 1.
	MaxAttempts int

	// RetryHosts lists the hosts whose rate-limited responses are retried.
	// An empty list uses the default of []string{"api.github.com"}.
	RetryHosts []string

	// Sleep waits for the requested duration and is interruptible via ctx.
	Sleep func(ctx context.Context, d time.Duration) error

	// Now returns the current time and is used to resolve X-RateLimit-Reset.
	Now func() time.Time

	// Logf emits one line per retry attempt.
	Logf func(format string, args ...any)
}

// RateLimitClass classifies why a GitHub response was rate limited.
type RateLimitClass string

const (
	// RateLimitClassNone means the response is not a rate-limit response.
	RateLimitClassNone RateLimitClass = ""
	// RateLimitClassPrimary means the primary quota is exhausted.
	RateLimitClassPrimary RateLimitClass = "primary"
	// RateLimitClassSecondary means GitHub asks for a secondary-limit backoff.
	RateLimitClassSecondary RateLimitClass = "secondary"
	// RateLimitClassUnclassified means a 429 without a recognizable limit signal.
	RateLimitClassUnclassified RateLimitClass = "unclassified"
)

// RateLimitClasses is the collection of every RateLimitClass value.
type RateLimitClasses []RateLimitClass

// AvailableRateLimitClasses lists every RateLimitClass value.
var AvailableRateLimitClasses = RateLimitClasses{
	RateLimitClassNone,
	RateLimitClassPrimary,
	RateLimitClassSecondary,
	RateLimitClassUnclassified,
}

// String returns the class as its underlying string value.
func (c RateLimitClass) String() string {
	return string(c)
}

// ClassifyRateLimitResponse classifies a GitHub rate-limit response.
// Only 403 and 429 responses are considered; every other status returns
// RateLimitClassNone. The rules are evaluated in order and the first match wins.
func ClassifyRateLimitResponse(resp *http.Response, body []byte) RateLimitClass {
	if resp == nil {
		return RateLimitClassNone
	}
	if resp.StatusCode != http.StatusForbidden &&
		resp.StatusCode != http.StatusTooManyRequests {
		return RateLimitClassNone
	}
	if strings.Contains(strings.ToLower(string(body)), "secondary rate limit") {
		return RateLimitClassSecondary
	}
	if resp.Header.Get("Retry-After") != "" {
		return RateLimitClassSecondary
	}
	if remaining, err := strconv.Atoi(resp.Header.Get("X-RateLimit-Remaining")); err == nil &&
		remaining == 0 {
		return RateLimitClassPrimary
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return RateLimitClassUnclassified
	}
	return RateLimitClassNone
}

// NewRoundTripperRetryWithBackoff wraps a RoundTripper with rate-limit aware
// retry logic. A 403 or 429 response from one of RetryHosts that classifies as
// a rate limit is retried after the delay the response asks for.
func NewRoundTripperRetryWithBackoff(
	roundTripper http.RoundTripper,
	options RetryBackoffOptions,
) http.RoundTripper {
	if options.MaxAttempts == 0 {
		options.MaxAttempts = 5
	}
	if options.MaxAttempts < 1 {
		options.MaxAttempts = 1
	}
	if len(options.RetryHosts) == 0 {
		options.RetryHosts = []string{"api.github.com"}
	}
	if options.Sleep == nil {
		options.Sleep = defaultRetryBackoffSleep
	}
	if options.Now == nil {
		options.Now = libtime.Now
	}
	if options.Logf == nil {
		options.Logf = glog.Warningf
	}
	return &retryBackoffRoundTripper{
		roundTripper: roundTripper,
		maxAttempts:  options.MaxAttempts,
		retryHosts:   options.RetryHosts,
		sleep:        options.Sleep,
		now:          options.Now,
		logf:         options.Logf,
	}
}

type retryBackoffRoundTripper struct {
	roundTripper http.RoundTripper
	maxAttempts  int
	retryHosts   []string
	sleep        func(ctx context.Context, d time.Duration) error
	now          func() time.Time
	logf         func(format string, args ...any)
}

func (r *retryBackoffRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if !r.isRetryHost(req) {
		return r.roundTripper.RoundTrip(req)
	}

	ctx := req.Context()
	body, err := readRequestBody(req)
	if err != nil {
		return nil, err
	}

	retries := 0
	for {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, errors.Wrap(ctx, ctxErr, "retry cancelled")
		}

		resp, err := r.roundTripper.RoundTrip(cloneRequestWithBody(ctx, req, body))
		if err != nil {
			return nil, errors.Wrap(ctx, err, "roundtrip failed")
		}

		if !isRateLimitStatus(resp.StatusCode) {
			return resp, nil
		}

		respBody, err := bufferResponseBody(resp)
		if err != nil {
			return nil, errors.Wrap(ctx, err, "read response body failed")
		}

		class := ClassifyRateLimitResponse(resp, respBody)
		if class == RateLimitClassNone {
			return resp, nil
		}
		if retries >= r.maxAttempts {
			r.logf("retry exhausted status=%d retries=%d", resp.StatusCode, retries)
			return resp, nil
		}

		retries++
		wait := computeRetryWait(resp, r.now())
		r.logRetry(retries, resp.StatusCode, class, wait)
		if err := r.sleep(ctx, wait); err != nil {
			return nil, errors.Wrap(ctx, err, "retry wait cancelled")
		}
	}
}

func (r *retryBackoffRoundTripper) isRetryHost(req *http.Request) bool {
	if req.Header.Get(PreventRetryHeaderName) != "" {
		return false
	}
	host := req.URL.Hostname()
	for _, retryHost := range r.retryHosts {
		if retryHost == host {
			return true
		}
	}
	return false
}

func (r *retryBackoffRoundTripper) logRetry(
	attempt int,
	status int,
	class RateLimitClass,
	wait time.Duration,
) {
	r.logf("retry attempt=%d status=%d class=%s wait=%s", attempt, status, class, wait)
}

func isRateLimitStatus(statusCode int) bool {
	return statusCode == http.StatusForbidden || statusCode == http.StatusTooManyRequests
}

// bufferResponseBody reads the response body to EOF, closes it and restores it
// so the caller still reads the original bytes. A nil body is treated as empty.
func bufferResponseBody(resp *http.Response) ([]byte, error) {
	if resp.Body == nil {
		resp.Body = io.NopCloser(bytes.NewReader(nil))
		return nil, nil
	}
	body, err := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	resp.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

// computeRetryWait returns the delay requested by the response, falling back to
// the primary-quota reset time and then to one second.
func computeRetryWait(resp *http.Response, now time.Time) time.Duration {
	if value := resp.Header.Get("Retry-After"); value != "" {
		if wait, ok := parseRetryAfter(value, now); ok {
			return wait
		}
	}
	if value := resp.Header.Get("X-RateLimit-Reset"); value != "" {
		if reset, err := strconv.ParseInt(value, 10, 64); err == nil {
			wait := time.Unix(reset, 0).Sub(now)
			if wait < 0 {
				return 0
			}
			return wait
		}
	}
	return time.Second
}

func parseRetryAfter(value string, now time.Time) (time.Duration, bool) {
	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds < 0 {
			seconds = 0
		}
		return time.Duration(seconds) * time.Second, true
	}
	if at, err := http.ParseTime(value); err == nil {
		wait := at.Sub(now)
		if wait < 0 {
			return 0, true
		}
		return wait, true
	}
	return 0, false
}

func defaultRetryBackoffSleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
