---
id: 2026-11-24-concurrency-limiter-roundtripper
title: Concurrency-Limiting RoundTripper
status: In Progress
created: 2026-11-24
authors: ["@bborbe"]
tags: [prd, roundtripper, concurrency, github, rate-limit]
related_issues: []
implemented_pr: ""
superseded_by: ""
---

# PRD: Concurrency-Limiting RoundTripper

**Status**: In Progress 🚧
**Created**: 2026-11-24
**Target**: github.com/bborbe/http library

## Summary

Add an opt-in RoundTripper that bounds how many requests are in flight at once,
so a caller answers a secondary rate limit by pacing rather than by firing more
requests.

## Problem Statement

GitHub's secondary rate limiter is a concurrency heuristic: it reacts to how many
requests are in flight at the same time. The existing `NewRoundTripperRateLimit`
paces requests over time but does not bound concurrency, and the retry
RoundTripper reacts after the limit has already been hit. Retrying immediately
against a secondary limit makes the situation worse. Callers need a way to keep
the in-flight count under a named bound.

## Goals

- Bound the number of simultaneous in-flight requests to a caller-supplied
  `maxConcurrent`.
- Treat a non-positive bound as 1, because an unbuffered slot channel would
  deadlock.
- Wait for a free slot instead of rejecting the request, and return the context's
  error if the context ends first.
- Log the effective bound once at construction so a running pod can be read back.
- Stay host-agnostic: the caller scopes the limiter by wrapping only the client
  that should be bounded.
- Leave every existing round-tripper untouched; the constructor is additive.

### Non-Goals

- Replacing or changing `NewRoundTripperRateLimit`, `NewRoundTripperRetry` or
  `NewRoundTripperRetryWithBackoff`.
- Per-host bounds or dynamic limits driven by response headers.
- Holding a slot for the lifetime of the response body; the slot covers the
  `RoundTrip` call only.

## Technical Specification

New file `http_roundtripper-limit-concurrency.go`, package `http`.

```go
func NewRoundTripperLimitConcurrency(
    roundTripper http.RoundTripper,
    maxConcurrent int,
    logf func(format string, args ...any),
) http.RoundTripper
```

A `maxConcurrent` below 1 is raised to 1. A nil `logf` falls back to
`glog.Warningf`. On construction, `logf` is called once with exactly
`inflight_bound=%d` carrying the effective bound; nothing is logged per request
at Info level.

`RoundTrip` acquires a slot from a buffered channel of size `maxConcurrent` with
a select on the slot channel and `req.Context().Done()`. On cancellation it
returns `errors.Wrap(ctx, ctx.Err(), "acquire inflight slot")`. The release is
registered with `defer` only after the slot is acquired, so a cancelled acquire
never releases a slot it never took and every post-acquire return path frees the
slot. A transport error is wrapped with `github.com/bborbe/errors` as
`roundtrip failed`.

## Testing Strategy

`http_roundtripper-limit-concurrency_test.go` (Ginkgo/Gomega) covers:

- a peak-concurrency assertion: 20 concurrent calls through a recording stub
  never exceed the bound and all complete;
- `maxConcurrent` of 0 raised to 1, asserted through the `inflight_bound=1` log
  line;
- a nil `logf` not panicking;
- the `inflight_bound=<n>` line emitted exactly once, not per request;
- a context cancelled while waiting returning a wrapped context error without
  consuming a slot or inflating the observed peak;
- a transport error being wrapped and its slot released.

## Open Questions

None.
