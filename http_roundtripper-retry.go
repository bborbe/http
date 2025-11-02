// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/bborbe/errors"
	"github.com/golang/glog"
)

// PreventRetryHeaderName is the HTTP header name used to disable retry logic.
// When this header is present in a request, the retry RoundTripper will not attempt retries.
const PreventRetryHeaderName = "X-Prevent-Retry"

// defaultSkipStatusCodes contains HTTP status codes that should not trigger retries.
// These are typically client errors that won't be resolved by retrying.
var defaultSkipStatusCodes = []int{
	400, // Bad Request
	401, // Unauthorized
	404, // Not Found
}

// NewRoundTripperRetry wraps a RoundTripper with retry logic using default skip status codes.
// It will retry failed requests up to retryLimit times with retryDelay between attempts.
// Requests with 400, 401, and 404 status codes are not retried as they indicate client errors.
func NewRoundTripperRetry(
	roundTripper http.RoundTripper,
	retryLimit int,
	retryDelay time.Duration,
) http.RoundTripper {
	return NewRoundTripperRetryWithSkipStatus(
		roundTripper,
		retryLimit,
		retryDelay,
		defaultSkipStatusCodes,
	)
}

// NewRoundTripperRetryWithSkipStatus wraps a RoundTripper with retry logic and custom skip status codes.
// It allows specifying which HTTP status codes should not trigger retries.
// This is useful when you want to customize which responses are considered permanent failures.
func NewRoundTripperRetryWithSkipStatus(
	roundTripper http.RoundTripper,
	retryLimit int,
	retryDelay time.Duration,
	skipStatusCodes []int,
) http.RoundTripper {
	return &retryRoundTripper{
		roundTripper:       roundTripper,
		retryLimit:         retryLimit,
		retryDelay:         retryDelay,
		skipStatusCodesMap: toStatusCodeMap(skipStatusCodes),
	}
}

// retryRoundTripper implements http.RoundTripper with retry logic.
type retryRoundTripper struct {
	roundTripper       http.RoundTripper
	retryLimit         int
	retryDelay         time.Duration
	skipStatusCodesMap map[int]bool
}

func (r *retryRoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if req.Header.Get(PreventRetryHeaderName) != "" {
		glog.V(4).Infof("found prevent retry header")
		return r.roundTripper.RoundTrip(req)
	}

	ctx := req.Context()
	body, err := r.readRequestBody(req)
	if err != nil {
		return nil, err
	}

	retryCounter := 0
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		resp, err = r.attemptRequest(ctx, req, body, &retryCounter)
		if err != nil {
			return nil, err
		}
		if resp != nil {
			return resp, nil
		}
	}
}

func (r *retryRoundTripper) attemptRequest(
	ctx context.Context,
	req *http.Request,
	body []byte,
	retryCounter *int,
) (*http.Response, error) {
	resp, err := r.executeRequest(ctx, req, body)
	if err != nil {
		return r.handleRequestError(ctx, req, err, retryCounter)
	}
	return r.handleRequestResponse(ctx, req, resp, retryCounter)
}

func (r *retryRoundTripper) handleRequestError(
	ctx context.Context,
	req *http.Request,
	err error,
	retryCounter *int,
) (*http.Response, error) {
	if !r.shouldRetryError(err, *retryCounter) {
		return nil, errors.Wrap(ctx, err, "roundtrip failed")
	}
	if err := r.delayAndIncrement(ctx, retryCounter, req, err); err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *retryRoundTripper) handleRequestResponse(
	ctx context.Context,
	req *http.Request,
	resp *http.Response,
	retryCounter *int,
) (*http.Response, error) {
	if !r.shouldRetryStatusCode(resp.StatusCode, *retryCounter) {
		return resp, nil
	}

	glog.V(1).Infof(
		"%s request to %s failed with status code %d => retry",
		req.Method,
		removeSensibleArgs(req.URL.String()),
		resp.StatusCode,
	)

	if err := r.delayAndIncrement(ctx, retryCounter, req, nil); err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *retryRoundTripper) readRequestBody(req *http.Request) ([]byte, error) {
	// TODO: implement me
	// limit body reader to x mb
	if req.Body == nil {
		return nil, nil
	}
	return io.ReadAll(req.Body)
}

func (r *retryRoundTripper) executeRequest(
	ctx context.Context,
	req *http.Request,
	body []byte,
) (*http.Response, error) {
	reqCloned := req.Clone(ctx)
	if req.Body != nil {
		reqCloned.Body = io.NopCloser(bytes.NewBuffer(body))
	}
	return r.roundTripper.RoundTrip(reqCloned.WithContext(ctx))
}

func (r *retryRoundTripper) shouldRetryError(err error, retryCounter int) bool {
	return IsRetryError(err) && retryCounter < r.retryLimit
}

func (r *retryRoundTripper) shouldRetryStatusCode(
	statusCode int,
	retryCounter int,
) bool {
	if statusCode < 400 {
		return false
	}
	if r.skipStatusCodesMap[statusCode] {
		return false
	}
	if r.retryLimit == retryCounter {
		return statusCode == 502 || statusCode == 503 || statusCode == 504
	}
	return true
}

func (r *retryRoundTripper) delayAndIncrement(
	ctx context.Context,
	retryCounter *int,
	req *http.Request,
	err error,
) error {
	if err != nil {
		glog.V(1).Infof(
			"%s request to %s failed with error: %v => retry",
			req.Method,
			removeSensibleArgs(req.URL.String()),
			err,
		)
	}
	if delayErr := r.delay(ctx); delayErr != nil {
		return errors.Wrap(ctx, delayErr, "delay failed")
	}
	*retryCounter++
	return nil
}

func (r *retryRoundTripper) delay(ctx context.Context) error {
	if r.retryDelay > 0 {
		glog.V(3).Infof("sleep for %v", r.retryDelay)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(r.retryDelay):
		}
	}
	return nil
}

var removeSensibleArgsRegex = regexp.MustCompile(`hapikey=[^&]+`)

func removeSensibleArgs(value string) string {
	return removeSensibleArgsRegex.ReplaceAllString(value, "hapikey=***")
}

func toStatusCodeMap(skipStatusCodes []int) map[int]bool {
	skipStatusCodesMap := map[int]bool{}
	for _, code := range skipStatusCodes {
		skipStatusCodesMap[code] = true
	}
	return skipStatusCodesMap
}
