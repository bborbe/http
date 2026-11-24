---
id: 2026-11-24-retry-backoff-roundtripper
title: Retry-After-aware GitHub Retry RoundTripper
status: Implemented
created: 2026-11-24
authors: ["@bborbe"]
tags: [prd, roundtripper, retry, github, rate-limit]
related_issues: []
implemented_pr: ""
superseded_by: ""
---

# PRD: Retry-After-aware GitHub Retry RoundTripper

**Status**: Implemented ✅
**Created**: 2026-11-24
**Target**: github.com/bborbe/http library

## Summary

Add an opt-in RoundTripper that retries rate-limited GitHub API responses after
the delay the response asks for, instead of handing the rate-limit response back
to the caller immediately.

## Problem Statement

The existing `NewRoundTripperRetry` retries with a fixed delay and has no notion
of GitHub rate limits. A `429` or rate-limited `403` is retried blindly, and the
`Retry-After` and `X-RateLimit-Reset` headers are ignored. Changing that
behaviour in place would alter timing for every existing caller.

## Goals

- Retry only GitHub rate-limit responses (`403` / `429`).
- Honour `Retry-After` (seconds or HTTP date), then `X-RateLimit-Reset`, then a
  one-second default.
- Bound retries with `MaxAttempts` and return the last response when exhausted.
- Defer a retry whose required wait exceeds `MaxWait` instead of sleeping: the
  response is returned unchanged and the next poll cycle retries.
- Emit exactly one warning log line per retry, one on exhaustion, and one when a
  retry is deferred.
- Leave non-GitHub hosts, `X-Prevent-Retry` requests and successful responses
  untouched, streaming large bodies.

### Non-Goals

- Retrying transport errors (the existing retry RoundTripper already covers
  transient errors).
- Retrying any non-GitHub host by default.

## Technical Specification

New file `http_roundtripper-retry-backoff.go`, package `http`.

```go
type RetryBackoffOptions struct {
    MaxAttempts int
    RetryHosts  []string
    Sleep       func(ctx context.Context, d time.Duration) error
    Now         func() time.Time
    Logf        func(format string, args ...any)
    MaxWait     time.Duration
}

func NewRoundTripperRetryWithBackoff(
    roundTripper http.RoundTripper,
    options RetryBackoffOptions,
) http.RoundTripper
```

Zero-value defaults: `MaxAttempts` 5, `RetryHosts` `[]string{"api.github.com"}`,
an interruptible timer-based `Sleep`, `libtime.Now`, and `glog.Warningf`.

`MaxWait` caps the wait a single retry will take. A zero or negative value falls
back to the default of five minutes. The bound is checked with a strict greater
than: a wait exactly equal to `MaxWait` is still retried, and only a wait beyond
it is deferred. On the deferred path the response is returned unchanged with a
nil error after one `retry deferred wait=<d>` log line; `logRetry` is not called,
`Sleep` is not called, and the attempt counter is not incremented, so a deferred
response does not consume an attempt. This ends the caller's poll cycle promptly
and lets the next cycle retry instead of blocking for hours on a reset far in the
future.

The cap is caller-tunable rather than a package constant because a five-minute
policy does not suit every poller: a background job may tolerate an hour-long
sleep while a foreground request cycle must return promptly.

Responses are classified by `ClassifyRateLimitResponse`:

```go
type RateLimitClass string

const (
    RateLimitClassNone         RateLimitClass = ""
    RateLimitClassPrimary      RateLimitClass = "primary"
    RateLimitClassSecondary    RateLimitClass = "secondary"
    RateLimitClassUnclassified RateLimitClass = "unclassified"
)

func ClassifyRateLimitResponse(resp *http.Response, body []byte) RateLimitClass
```

Only `403` and `429` are considered. Rules are evaluated in order and the first
match wins: body contains `secondary rate limit` → `secondary`; `Retry-After`
present → `secondary`; `X-RateLimit-Remaining` parses to `0` → `primary`;
status `429` → `unclassified`; otherwise `none`.

The response body is buffered only for `403`/`429` and restored before return,
so the caller reads the original bytes and successful downloads keep streaming.
The read/clone helpers are shared with `http_roundtripper-retry.go`.

## Testing Strategy

`http_roundtripper-retry-backoff_test.go` (Ginkgo/Gomega) covers both rate-limit
classes, the unclassified case, non-GitHub hosts, the attempt bound, context
cancellation, request-body replay and response-body restoration.

## Open Questions

None.
