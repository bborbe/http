---
status: completed
summary: Added a configurable MaxWait (default 5m) to the retry-backoff round-tripper so a retry whose wait exceeds it is deferred with one log line instead of slept, plus tests, PRD, README and changelog updates
execution_id: http-exec-005-retry-deferred-cap
dark-factory-version: v0.198.17
created: "2026-11-24T21:08:39Z"
queued: "2026-11-24T21:08:39Z"
started: "2026-11-24T21:09:04Z"
completed: "2026-11-24T21:11:32Z"
---

# Defer a retry whose wait would exceed five minutes

<summary>
- A retry whose required wait is longer than five minutes is deferred instead of slept
- The round-tripper logs one line naming the wait it declined to take, then returns the response unchanged
- The caller's poll cycle therefore ends promptly and the next cycle retries
- The cap is configurable and defaults to five minutes
- A wait exactly at the cap still retries; only a wait beyond it defers
- Every other behaviour of the retry round-tripper is unchanged
- The design record is updated, because it currently rejects this behaviour
</summary>

<objective>
Stop a primary-quota reset an hour away from blocking a poll cycle for hours. Today the round-tripper sleeps the full computed wait, so one request can hold a cycle far longer than the interval that would have retried it — the silent stall this work exists to remove.
</objective>

<context>
Read `CLAUDE.md` for project conventions and `docs/prd/2026-11-24-retry-backoff-roundtripper.md` — the design this prompt extends. **Its Non-Goals currently reject capping the server-provided wait; this change deliberately reverses that**, so update the PRD as part of the work.

Read `http_roundtripper-retry-backoff.go` before writing anything — it holds the options struct, the retry loop, `computeRetryWait`, and the ctx-aware sleep helper. Read `http_roundtripper-retry-backoff_test.go` for the Ginkgo/Gomega style, the `retryBackoffCase` table, the injected `Now`/`Sleep`/`Logf` seams, and the existing 429 fixtures.

Also read `docs/dod.md` (the project's own Definition of Done).

Coding-plugin docs under `/home/node/.claude/plugins/marketplaces/coding/docs/`: `go-glog-guide.md` (log-line style), `go-library-guide.md` and `go-functional-options-pattern.md` (public field naming and doc-comment style — but keep this file's existing `RetryBackoffOptions` struct and plain constructor; do not convert it to variadic functional options), and `changelog-guide.md` (the `## Unreleased` entry format and conventional prefix).
</context>

<requirements>
1. Add one field to `RetryBackoffOptions` in `http_roundtripper-retry-backoff.go`:
   ```go
   // MaxWait caps the wait a single retry will take. A wait longer than this is
   // deferred: the response is returned and the next poll cycle retries.
   // A zero or negative value uses the default of 5 minutes.
   MaxWait time.Duration
   ```

2. In `retryBackoffRoundTripper.RoundTrip`, compute `wait := computeRetryWait(resp, r.now())` **before** `retries++` and **before** `logRetry`. If `wait > r.maxWait`, call `r.logf("retry deferred wait=%s", wait)` and return `resp, nil` unchanged — do not call `logRetry`, do not sleep, and do not increment `retries`. The deferred case therefore emits exactly **one** log line, the `retry deferred` one. Otherwise increment `retries`, call `logRetry`, and sleep exactly as today.

3. Add a `maxWait time.Duration` field to `retryBackoffRoundTripper` and set it in the struct literal inside `NewRoundTripperRetryWithBackoff`. Resolve the effective bound there alongside the other defaults: `MaxWait` of zero becomes 5 minutes; a negative value is treated as 5 minutes.

4. A wait exactly equal to the bound still retries — the comparison is strictly greater than.

5. Add a `maxWait time.Duration` field to the `retryBackoffCase` struct, pass it into `libhttp.RetryBackoffOptions{...}` in the table runner, and add cases to `http_roundtripper-retry-backoff_test.go`:
   - 403 with `X-RateLimit-Remaining: 0` and `X-RateLimit-Reset` one hour ahead of the injected whole-second `Now` → exactly one call, exactly one `retry deferred wait=1h0m0s` line (no `retry attempt=` line), and the response returned with a nil error, its original 403 status, and its original body still readable
   - a wait exactly at `MaxWait` → retried (two calls), no `retry deferred` line
   - a wait one second over `MaxWait` → deferred
   - `MaxWait` left at zero → the default of 5 minutes applies (a 6-minute reset defers, a 4-minute reset retries)
   - a negative `MaxWait` → the 5-minute default applies (a 6-minute reset defers)
   Assert the injected `Sleep` was never called in the deferred cases.

6. Update `docs/prd/2026-11-24-retry-backoff-roundtripper.md` to match: remove the "Capping the server-provided wait" line from Non-Goals, add a Goals bullet for deferring a wait that exceeds the cap, add `MaxWait time.Duration` to the `RetryBackoffOptions` block in Technical Specification, and in Technical Specification document the five-minute default, the zero/negative fallback, the strict-greater comparison, and that a deferred response does not consume an attempt. Record why the cap is caller-tunable rather than a package constant — a five-minute policy does not suit every poller.

7. Add a `## Unreleased` entry to `CHANGELOG.md` (prefix the line `feat:`, matching the existing entries) that calls out the behaviour change — an existing caller now receives a deferred 403/429 response instead of a long blocking sleep. Also add one line to the `HTTP Client & RoundTrippers` list in `README.md`, directly under the existing `GitHub rate-limit retries that honour ...` bullet: a wait longer than five minutes is deferred instead of slept, so the caller's poll cycle returns promptly.

8. Do not change `NewRoundTripperRetry`, `NewRoundTripperRetryWithSkipStatus`, `ClassifyRateLimitResponse`, or the existing retry log lines.

9. Run `make precommit` and confirm it passes; walk each numbered requirement against the change.
</requirements>

<constraints>
- Do NOT commit — dark-factory handles git.
- Existing tests must still pass unchanged.
- All paths you create, edit or run are repo-relative; the container-absolute coding-plugin doc paths in `<context>` are read-only references and are valid.
- Use Ginkgo/Gomega for tests.
- Do not add a `//nolint` without an explanation on the same line.
- Do not run `go mod vendor`.
</constraints>

<verification>
Run `make precommit` — must pass.
</verification>
