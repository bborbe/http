// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libhttp "github.com/bborbe/http"
	"github.com/bborbe/http/mocks"
)

// concurrencyRecorder records the peak number of simultaneous RoundTrip calls.
type concurrencyRecorder struct {
	mux       sync.Mutex
	current   int
	peak      int
	completed int
	release   chan struct{}
	block     bool
}

func (r *concurrencyRecorder) RoundTrip(_ *http.Request) (*http.Response, error) {
	r.mux.Lock()
	r.current++
	if r.current > r.peak {
		r.peak = r.current
	}
	r.mux.Unlock()

	if r.block {
		<-r.release
	} else {
		time.Sleep(5 * time.Millisecond)
	}

	r.mux.Lock()
	r.current--
	r.completed++
	r.mux.Unlock()
	return &http.Response{StatusCode: http.StatusOK}, nil
}

func (r *concurrencyRecorder) snapshot() (current int, peak int, completed int) {
	r.mux.Lock()
	defer r.mux.Unlock()
	return r.current, r.peak, r.completed
}

// logRecorder captures logf calls so tests can assert on them.
type logRecorder struct {
	mux   sync.Mutex
	lines []string
}

func (r *logRecorder) logf(format string, args ...any) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.lines = append(r.lines, fmt.Sprintf(format, args...))
}

func (r *logRecorder) recorded() []string {
	r.mux.Lock()
	defer r.mux.Unlock()
	return append([]string(nil), r.lines...)
}

var _ = Describe("RoundTripperLimitConcurrency", func() {
	var req *http.Request
	var err error
	BeforeEach(func() {
		req, err = http.NewRequest(http.MethodGet, "http://example.com", &bytes.Buffer{})
		Expect(err).To(BeNil())
	})

	Context("peak concurrency", func() {
		It("never exceeds the bound and completes all requests", func() {
			recorder := &concurrencyRecorder{}
			roundTripper := libhttp.NewRoundTripperLimitConcurrency(
				recorder,
				3,
				func(string, ...any) {},
			)

			const total = 20
			var wg sync.WaitGroup
			wg.Add(total)
			for i := 0; i < total; i++ {
				go func() {
					defer GinkgoRecover()
					defer wg.Done()
					resp, err := roundTripper.RoundTrip(req)
					Expect(err).To(BeNil())
					Expect(resp).NotTo(BeNil())
				}()
			}
			wg.Wait()

			_, peak, completed := recorder.snapshot()
			Expect(peak).To(BeNumerically("<=", 3))
			Expect(completed).To(Equal(total))
		})
	})

	Context("logf", func() {
		It("logs inflight_bound=1 when maxConcurrent is zero", func() {
			recorder := &logRecorder{}
			libhttp.NewRoundTripperLimitConcurrency(&mocks.HttpRoundTripper{}, 0, recorder.logf)
			Expect(recorder.recorded()).To(Equal([]string{"inflight_bound=1"}))
		})

		It("does not panic when logf is nil", func() {
			roundTripper := libhttp.NewRoundTripperLimitConcurrency(
				&mocks.HttpRoundTripper{},
				2,
				nil,
			)
			Expect(roundTripper).NotTo(BeNil())
		})

		It("logs inflight_bound exactly once, at construction and not per request", func() {
			recorder := &logRecorder{}
			roundTripper := libhttp.NewRoundTripperLimitConcurrency(
				okRoundTripper(),
				2,
				recorder.logf,
			)
			Expect(recorder.recorded()).To(Equal([]string{"inflight_bound=2"}))

			_, err := roundTripper.RoundTrip(req)
			Expect(err).To(BeNil())
			_, err = roundTripper.RoundTrip(req)
			Expect(err).To(BeNil())
			Expect(recorder.recorded()).To(Equal([]string{"inflight_bound=2"}))
		})
	})

	Context("context cancelled while waiting", func() {
		It("returns a wrapped context error, consumes no slot and keeps the bound", func() {
			recorder := &concurrencyRecorder{block: true, release: make(chan struct{})}
			roundTripper := libhttp.NewRoundTripperLimitConcurrency(
				recorder,
				2,
				func(string, ...any) {},
			)

			var wg sync.WaitGroup
			wg.Add(2)
			for i := 0; i < 2; i++ {
				go func() {
					defer GinkgoRecover()
					defer wg.Done()
					_, err := roundTripper.RoundTrip(req)
					Expect(err).To(BeNil())
				}()
			}
			Eventually(func() int {
				current, _, _ := recorder.snapshot()
				return current
			}).Should(Equal(2))

			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			waiterReq := req.Clone(ctx)
			_, err := roundTripper.RoundTrip(waiterReq)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("acquire inflight slot"))

			_, peak, _ := recorder.snapshot()
			Expect(peak).To(Equal(2))

			close(recorder.release)
			wg.Wait()

			// The cancelled waiter never took a slot, so a later call still succeeds.
			resp, err := roundTripper.RoundTrip(req)
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())
		})
	})

	Context("transport error", func() {
		It("wraps the error and releases the slot", func() {
			roundTripperMock := &mocks.HttpRoundTripper{}
			roundTripperMock.RoundTripReturnsOnCall(0, nil, io.EOF)
			roundTripperMock.RoundTripReturnsOnCall(
				1,
				&http.Response{StatusCode: http.StatusOK},
				nil,
			)
			roundTripper := libhttp.NewRoundTripperLimitConcurrency(
				roundTripperMock,
				1,
				func(string, ...any) {},
			)

			_, err := roundTripper.RoundTrip(req)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("roundtrip failed"))

			resp, err := roundTripper.RoundTrip(req)
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())
			Expect(roundTripperMock.RoundTripCallCount()).To(Equal(2))
		})
	})
})

func okRoundTripper() *mocks.HttpRoundTripper {
	roundTripperMock := &mocks.HttpRoundTripper{}
	roundTripperMock.RoundTripReturns(&http.Response{StatusCode: http.StatusOK}, nil)
	return roundTripperMock
}
