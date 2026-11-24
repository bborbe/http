---
status: completed
summary: Added NewRoundTripperLimitConcurrency with bounded in-flight requests, Ginkgo tests, CHANGELOG/README entries, and a PRD
execution_id: http-exec-006-concurrency-limiter-roundtripper
dark-factory-version: v0.198.17
created: "2026-11-24T21:08:39Z"
queued: "2026-11-24T21:08:39Z"
started: "2026-11-24T21:11:34Z"
completed: "2026-11-24T21:14:41Z"
---

# Add a concurrency-limiting round-tripper

<summary>
- A new opt-in HTTP round-tripper bounds how many requests are in flight at once
- Callers pass the bound explicitly; the only correction is that a non-positive bound is treated as 1
- The effective bound is logged once at construction so a running pod can be read back
- A request that cannot get a slot waits for one, and returns the context's error if the context ends first
- The limiter is host-agnostic — the caller scopes it by wrapping only the client it wants bounded
- Every other round-tripper keeps its current behaviour; this one is additive
- Tests assert the observed peak never exceeds the bound, and that a cancelled context neither consumes nor inflates it
</summary>

<objective>
Give callers a way to bound concurrent GitHub requests, so a secondary rate limit is answered by pacing rather than by firing more requests. GitHub's secondary limiter is a concurrency heuristic — retrying immediately makes it worse — so the fix is to keep the in-flight count under a named bound.
</objective>

<context>
Read `CLAUDE.md` for project conventions, `docs/dod.md` for the Definition of Done this prompt is judged against, and `docs/prd/README.md` for the PRD convention a new public API must follow.

Read these files before writing anything:
- `http_roundtripper-ratelimit.go` — an existing round-tripper that limits request rate; this one limits concurrency instead, and should follow its constructor and doc-comment shape.
- `http_roundtripper-retry-backoff.go` — the ctx-cancellation and `logf` conventions to match.
- `http_roundtripper-retry_test.go` — the Ginkgo/Gomega style and `mocks.HttpRoundTripper` usage.

Coding-plugin docs under `/home/node/.claude/plugins/marketplaces/coding/docs/`: `go-concurrency-patterns.md` (no raw goroutines in production code), `go-library-guide.md` (new exported constructor), `go-glog-guide.md` (log levels).
</context>

<requirements>
1. Add a new file `http_roundtripper-limit-concurrency.go`, package `http`.

2. Declare one constructor:
   ```go
   func NewRoundTripperLimitConcurrency(
       roundTripper http.RoundTripper,
       maxConcurrent int,
       logf func(format string, args ...any),
   ) http.RoundTripper
   ```
   A `maxConcurrent` below 1 is raised to 1 (an unbuffered channel would deadlock). A nil `logf` falls back to `glog.Warningf`.

3. On construction, call `logf` once with exactly
   ```
   inflight_bound=%d
   ```
   carrying the effective bound. Log nothing per request at Info level.

4. `RoundTrip` acquires a slot before delegating and releases it when it returns:
   - Acquire with a select on the slot channel and `req.Context().Done()`; on cancellation return `errors.Wrap(ctx, ctx.Err(), "acquire inflight slot")`.
   - Register the release with `defer` **only after** the slot is acquired, so a cancelled acquire never releases a slot it never took and every post-acquire return path frees the slot.
   - The slot covers the call to the inner round-tripper, not the response body — release when `RoundTrip` returns.
   - Wrap a transport error with `github.com/bborbe/errors` in the style of `http_roundtripper-retry-backoff.go`.

5. Add `http_roundtripper-limit-concurrency_test.go` (Ginkgo/Gomega, `package http_test`, `mocks.HttpRoundTripper`) with:
   - a peak-concurrency assertion: drive 20 concurrent `RoundTrip` calls through a stub that records the current in-flight count, and assert the observed peak never exceeded the bound and that all 20 completed
   - a `maxConcurrent` of 0 is raised to 1, asserted through the `inflight_bound=1` log line
   - a nil `logf` does not panic and construction still succeeds
   - the `inflight_bound=<n>` line is emitted exactly once, at construction, and not per request
   - a context cancelled while waiting for a slot returns a wrapped context error, does not consume a slot (a later call still succeeds), and does not inflate the bound (drive `bound` concurrent calls plus one cancelled waiter, then assert the observed in-flight peak is still `bound`)
   - a transport error is wrapped and the slot is released (a subsequent call still succeeds)

6. Add a `## Unreleased` entry to `CHANGELOG.md` (prefix the line `feat:`, matching the existing entries) and one line to the `HTTP Client & RoundTrippers` list in `README.md`.

7. Write `docs/prd/2026-11-24-concurrency-limiter-roundtripper.md` using the shape in `docs/prd/README.md` (Summary, Problem Statement, Goals, Non-Goals, Technical Specification, Testing Strategy) and add it to the `Current PRDs` table with status `In Progress`. A new public API requires a PRD per the project's own convention.

8. Do not modify any existing round-tripper or its behaviour. This constructor is additive.

9. Run `make precommit` and confirm it passes; walk each numbered requirement against the change.
</requirements>

<constraints>
- Do NOT commit — dark-factory handles git.
- Existing tests must still pass unchanged.
- All paths are repo-relative; no absolute or home-relative paths.
- Use Ginkgo/Gomega for tests; use Counterfeiter mocks from `mocks/`.
- Wrap errors with `github.com/bborbe/errors`.
- Do not add a `//nolint` without an explanation on the same line.
- Do not run `go mod vendor`.
</constraints>

<verification>
Run `make precommit` — must pass.
</verification>
