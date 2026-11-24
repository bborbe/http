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
- Emit exactly one warning log line per retry and one on exhaustion.
- Leave non-GitHub hosts, `X-Prevent-Retry` requests and successful responses
  untouched, streaming large bodies.

### Non-Goals

- Retrying transport errors (the existing retry RoundTripper already covers
  transient errors).
- Retrying any non-GitHub host by default.
- Capping the server-provided wait.

## Technical Specification

New file `http_roundtripper-retry-backoff.go`, package `http`.

```go
type RetryBackoffOptions struct {
    MaxAttempts int
    RetryHosts  []string
    Sleep       func(ctx context.Context, d time.Duration) error
    Now         func() time.Time
    Logf        func(format string, args ...any)
}

func NewRoundTripperRetryWithBackoff(
    roundTripper http.RoundTripper,
    options RetryBackoffOptions,
) http.RoundTripper
```

Zero-value defaults: `MaxAttempts` 5, `RetryHosts` `[]string{"api.github.com"}`,
an interruptible timer-based `Sleep`, `libtime.Now`, and `glog.Warningf`.

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
