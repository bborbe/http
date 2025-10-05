// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

const metricsHost = "host"

const metricsStatusCode = "status"

const metricsMethod = "method"

var (
	totalCounter = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "http",
		Subsystem: "request",
		Name:      "total_counter",
		Help:      "Counts http request",
	}, []string{metricsHost, metricsMethod})
	successCounter = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "http",
		Subsystem: "request",
		Name:      "success_counter",
		Help:      "Counts successful http request",
	}, []string{metricsHost, metricsMethod, metricsStatusCode})
	failureCounter = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "http",
		Subsystem: "request",
		Name:      "failure_counter",
		Help:      "Counts failed http request",
	}, []string{metricsHost, metricsMethod})
	durationMeasure = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "http",
		Subsystem: "request",
		Name:      "duration",
		Help:      "Duration of http request",
		Buckets:   prometheus.LinearBuckets(4000, 1, 1),
	}, []string{metricsHost, metricsMethod})
)

func init() {
	prometheus.DefaultRegisterer.MustRegister(
		totalCounter,
		successCounter,
		failureCounter,
		durationMeasure,
	)
}

//counterfeiter:generate -o mocks/http-roundtripper-metrics.go --fake-name HttpRoundTripperMetrics . RoundTripperMetrics

// RoundTripperMetrics defines the interface for collecting HTTP request metrics.
// It provides methods to record various metrics about HTTP requests including counts, status codes, and durations.
type RoundTripperMetrics interface {
	TotalCounterInc(host string, method string)
	SuccessCounterInc(host string, method string, statusCode int)
	FailureCounterInc(host string, method string)
	DurationMeasureObserve(host string, method string, duration time.Duration)
}

// NewRoundTripperMetrics creates a new RoundTripperMetrics implementation that uses Prometheus metrics.
// The returned instance will record metrics to the default Prometheus registry.
func NewRoundTripperMetrics() RoundTripperMetrics {
	return &roundTripperMetrics{}
}

type roundTripperMetrics struct {
}

func (r *roundTripperMetrics) TotalCounterInc(host string, method string) {
	totalCounter.With(prometheus.Labels{
		metricsHost:   host,
		metricsMethod: method,
	}).Inc()

}

func (r *roundTripperMetrics) SuccessCounterInc(host string, method string, statusCode int) {
	successCounter.With(prometheus.Labels{
		metricsHost:       host,
		metricsMethod:     method,
		metricsStatusCode: strconv.Itoa(statusCode),
	}).Inc()
}

func (r *roundTripperMetrics) FailureCounterInc(host string, method string) {
	failureCounter.With(prometheus.Labels{
		metricsHost:   host,
		metricsMethod: method,
	}).Inc()
}

func (r *roundTripperMetrics) DurationMeasureObserve(
	host string,
	method string,
	duration time.Duration,
) {
	durationMeasure.With(prometheus.Labels{
		metricsHost:   host,
		metricsMethod: method,
	}).Observe(duration.Seconds())
}

// NewMetricsRoundTripper wraps a given RoundTripper and adds Prometheus metrics.
func NewMetricsRoundTripper(
	roundTripper http.RoundTripper,
	metrics RoundTripperMetrics,
) http.RoundTripper {
	m := &metricsRoundTripper{
		roundTripper: roundTripper,
		metrics:      metrics,
	}
	return m
}

// RoundTripper for recording prometheus metrics
type metricsRoundTripper struct {
	roundTripper http.RoundTripper
	metrics      RoundTripperMetrics
}

// RoundTrip records the request duration of every received request to prometheus
func (h *metricsRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	h.metrics.TotalCounterInc(req.Host, req.Method)
	start := time.Now()
	resp, err := h.roundTripper.RoundTrip(req)
	duration := time.Since(start)
	h.metrics.DurationMeasureObserve(req.Host, req.Method, duration)
	if err != nil {
		h.metrics.FailureCounterInc(req.Host, req.Method)
		glog.V(3).
			Infof("failed %s request to %s in %d ms: %v", req.Method, req.URL.String(), duration.Milliseconds(), err)
		return nil, err
	}
	h.metrics.SuccessCounterInc(req.Host, req.Method, resp.StatusCode)
	glog.V(3).
		Infof("complete %s request to %s in %d ms with status %d", req.Method, req.URL.String(), duration.Milliseconds(), resp.StatusCode)
	return resp, nil
}
