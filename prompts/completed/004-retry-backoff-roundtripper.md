---
status: completed
summary: Added a rate-limit-aware retry RoundTripper that honours Retry-After and X-RateLimit-Reset for GitHub hosts, with shared body helpers, tests, PRD, README and CHANGELOG entries
execution_id: http-exec-004-retry-backoff-roundtripper
dark-factory-version: v0.198.17
created: "2026-11-24T20:06:31Z"
queued: "2026-11-24T20:06:31Z"
started: "2026-11-24T20:07:22Z"
completed: "2026-11-24T20:10:58Z"
---

# Add a Retry-After-aware GitHub retry round-tripper

<summary>
- A new opt-in HTTP round-tripper retries a rate-limited GitHub response instead of handing it straight back to the caller
- It waits for the delay the response asks for, and falls back to the primary-quota reset time when the response carries no delay
- It stops after a bounded number of attempts and reports the failure in one line
- Every retry logs one warning line carrying the attempt number, status, rate-limit class and wait
- Only GitHub API hosts are retried; every other host passes through untouched
- Only a 403 or 429 response is buffered, so large downloads keep streaming and a successful response is never retried
- The response body survives every attempt, so the caller still reads the original bytes
- The existing retry round-tripper keeps its current behaviour for its existing callers
- Tests cover both rate-limit classes, the unclassified case, the non-GitHub host, the attempt bound, context cancellation and the restored body
</summary>

<objective>
Give callers a retry round-tripper that understands GitHub rate-limit responses, so a 429 becomes a wait-and-retry that honours the server's delay (bounded by the attempt limit and the caller's context) instead of a failed request — and make the retry visible in one log line per attempt. The existing `NewRoundTripperRetry` cannot serve this: it retries with a fixed delay, ignores `Retry-After` and `X-RateLimit-Reset`, and cannot be changed without altering timing for its existing callers.
</objective>

<context>
Read `CLAUDE.md` for project conventions.

Read these files before writing anything — they are the patterns the new code must follow:
- `http_roundtripper-retry.go` — the existing retry round-tripper: request cloning and body replay (`readRequestBody`, `executeRequest`), the ctx check at the top of each loop iteration, the `glog` calls, the `PreventRetryHeaderName` constant, and the error wrapping via `github.com/bborbe/errors`.
- `http_roundtripper-ratelimit.go` — how a second round-tripper is structured and constructed.
- `http_roundtripper-options.go` — the existing options struct this one sits beside.
- `http_roundtripper-retry_test.go` — the Ginkgo/Gomega test style and the `mocks.HttpRoundTripper` usage.
- `docs/dod.md` — the Definition of Done this prompt is judged against.
- `docs/prd/README.md` and one existing file under `docs/prd/` — the PRD shape you must follow.

Coding-plugin docs available in the container under `/home/node/.claude/plugins/marketplaces/coding/docs/` — read the ones your change touches: `go-enum-type-pattern.md` (the class type), `go-glog-guide.md` (log levels), `go-context-cancellation-in-loops.md` (the retry loop), `go-time-injection.md` (`libtime.Now`, never `time.Now`). The `Now func() time.Time` field in requirement 2 is a transport-level test seam, not an injected service dependency — the guide's `CurrentDateTimeGetter` anti-pattern targets service clocks; keep `func() time.Time` and do not convert it.
</context>

<requirements>
1. Add a new file `http_roundtripper-retry-backoff.go`, package `http`.

2. Declare an options struct and a constructor:
   ```go
   type RetryBackoffOptions struct {
       MaxAttempts int
       RetryHosts  []string
       Sleep       func(ctx context.Context, d time.Duration) error
       Now         func() time.Time
       Logf        func(format string, args ...any)
   }

   func NewRoundTripperRetryWithBackoff(roundTripper http.RoundTripper, options RetryBackoffOptions) http.RoundTripper
   ```
   Zero-value fields take these defaults: `MaxAttempts` 5 (a value below 1 is treated as 1), `RetryHosts` `[]string{"api.github.com"}`, `Sleep` a default that selects on `ctx.Done()` and `time.NewTimer(d)` (stop the timer on the ctx branch; never bare `time.Sleep`), `Now` `libtime.Now` (import `libtime "github.com/bborbe/time"` — never `time.Now`), `Logf` `glog.Warningf`.

3. Declare a rate-limit classification with named values, a collection, a `String` method, and one function:
   ```go
   type RateLimitClass string

   const (
       RateLimitClassNone         RateLimitClass = ""
       RateLimitClassPrimary      RateLimitClass = "primary"
       RateLimitClassSecondary    RateLimitClass = "secondary"
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
   func (c RateLimitClass) String() string

   func ClassifyRateLimitResponse(resp *http.Response, body []byte) RateLimitClass
   ```
   **Only status 403 and 429 are rate-limit responses** — return `RateLimitClassNone` for every other status before evaluating the rules below, so a 200 or 404 body can never trigger a retry.
   Then evaluate **in this order and stop at the first match** — the order is the contract, because GitHub sends `Retry-After` on some primary-limit responses:
   1. `body` contains `secondary rate limit` (case-insensitive) → `RateLimitClassSecondary`
   2. `Retry-After` header present → `RateLimitClassSecondary`
   3. `X-RateLimit-Remaining` parses to `0` → `RateLimitClassPrimary`
   4. status is 429 (any 429 that reached this rule, including one with a negative `X-RateLimit-Remaining`) → `RateLimitClassUnclassified`
   5. otherwise → `RateLimitClassNone`

   A plain 403 (no exhausted remaining, no `Retry-After`) must classify as `RateLimitClassNone` — it is a permissions failure, not a rate limit. A malformed `X-RateLimit-Remaining` is treated as absent.

4. Retry only when `req.URL.Hostname()` is in `RetryHosts`, `req.Header.Get(PreventRetryHeaderName)` is empty, and `ClassifyRateLimitResponse` returns a non-empty class. Every other request is passed straight through on the first attempt.

5. Compute the wait for a retry:
   - `Retry-After` present → parse it inline as an integer number of seconds (`strconv.Atoi`), or as an HTTP date (`http.ParseTime`), and use the resulting duration.
   - Otherwise, `X-RateLimit-Reset` present → wait until that unix timestamp, using the injected `Now`. A reset in the past yields a zero wait.
   - Otherwise a 1-second default.
   A malformed `Retry-After` or `X-RateLimit-Reset` falls through to the next source, ending at the 1-second default.
   Wait the **full** computed duration — do not cap it and do not substitute a fixed delay. The wait must be interruptible: call the injected `Sleep` with the request context and return whatever error `Sleep` returned, wrapped with `errors.Wrap(ctx, err, "retry wait cancelled")`. Also check `ctx.Err()` at the top of each loop iteration before issuing the next attempt, returning `errors.Wrap(ctx, ctx.Err(), "retry cancelled")` when set.

6. On every retry, before waiting, call `Logf` with exactly this shape so downstream log greps match:
   ```
   retry attempt=%d status=%d class=%s wait=%s
   ```
   The first retry is `attempt=1`. `wait` renders as a Go duration string (for example `2s`).

7. When `MaxAttempts` retries have been used without success, call `Logf` once with:
   ```
   retry exhausted status=%d retries=%d
   ```
   and return the **last response with a nil error**, so the caller's existing non-2xx handling sees the rate-limited response and ends its cycle rather than the process exiting or the item being dropped.

8. Extract the request read/clone/buffer logic to package-level helpers shared with `http_roundtripper-retry.go` — e.g. `func readRequestBody(req *http.Request) ([]byte, error)` and `func cloneRequestWithBody(ctx context.Context, req *http.Request, body []byte) *http.Request` — and call the clone helper on each attempt (do not copy the code). Also replace the bodies of `retryRoundTripper.readRequestBody` and `retryRoundTripper.executeRequest` in `http_roundtripper-retry.go` with calls to the new package-level helpers — requirement 13 forbids changing the constructors or their behaviour, not this internal delegation. A transport error from the inner round-tripper is wrapped with `github.com/bborbe/errors` in the style of the existing file and returned immediately — it is not retried (only a classified rate-limit response is).

9. Buffer and restore the response body **only for a 403 or 429 response** — every other response is returned with its body untouched so large GitHub downloads keep streaming. For a buffered response, read `resp.Body` to EOF into a byte slice (treat a nil `resp.Body` as empty), close it, and pass that slice to `ClassifyRateLimitResponse`. Before returning any buffered response — an intermediate retry or the final one — set `resp.Body = io.NopCloser(bytes.NewReader(body))` so the caller still reads the original bytes. Close the body of every response discarded on a retry (`bodyclose` is enabled).

10. Keep `RoundTrip` small enough for the repo's linters (`funlen` 80 lines, `nestif` 4): factor `bufferResponseBody`, `computeRetryWait` and `logRetry` into their own functions.

11. Add `http_roundtripper-retry-backoff_test.go` (Ginkgo/Gomega, `package http_test`, `mocks.HttpRoundTripper`) with a table-driven set of cases. Use `https://api.github.com/...` for every retrying case so the `RetryHosts` match is explicit, and a whole-second injected `Now` (for example `time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)`) wherever an exact wait is asserted:
   - 429 with `Retry-After: 2`, then 200 → two calls, one log line with `retry attempt=1 status=429 class=secondary wait=2s`
   - 429 with `Retry-After` set to an RFC 1123 HTTP date 30 seconds ahead of the injected `Now` → class `secondary`, wait `30s`
   - 403 with `X-RateLimit-Remaining: 0` and `X-RateLimit-Reset` 60 seconds ahead → class `primary`, wait `1m0s`
   - 403 with `X-RateLimit-Remaining: 0` **and** `Retry-After: 5` → class `secondary`, wait `5s` (rule 2 beats rule 3)
   - 429 with `X-RateLimit-Remaining: 10` and no `Retry-After` → class `unclassified`
   - 429 with `X-RateLimit-Remaining: -1` and no `Retry-After` → class `unclassified`, retried
   - 429 whose body contains `secondary rate limit` (case-insensitive) → class `secondary`
   - 403 with no `X-RateLimit-Remaining` and no `Retry-After` → `RateLimitClassNone`, one call
   - 403 with `X-RateLimit-Remaining: abc` (malformed) and no `Retry-After` → class `None`, one call
   - 403 with `X-RateLimit-Remaining: 0` and `X-RateLimit-Reset` in the past → class `primary`, wait `0s`
   - 429 with a malformed `Retry-After: garbage` and no `X-RateLimit-Reset` → class `secondary`, wait `1s`
   - 200 on the first call → one call, no log line
   - 200 whose body contains `secondary rate limit` → class `None`, one call, no retry
   - 429 from a non-GitHub host → one call, no retry, no log line
   - a request carrying `X-Prevent-Retry` → one call, no retry, no log line
   - persistent 429 with `MaxAttempts: 2` → three calls, a `retry exhausted status=429 retries=2` line, and the last response returned with a nil error
   - a retried response whose body reads back the original bytes after the round-tripper returns (non-nil `resp.Body` in the test fixture)
   - a request with a non-nil body retried once → the second attempt's body reads back the same bytes
   - a transport error from the inner round-tripper → wrapped error returned immediately, no retry, no retry log line
   - context cancelled during the wait → the wrapped error is returned (assert with `MatchError`/`errors.Is`, not `Equal`)
   - a table test over `AvailableRateLimitClasses` asserting each value's `String()` round-trips (update the table when a class is added)
   Assert the injected `Sleep` received the expected duration in each retrying case; do not let a test sleep for real.

12. Add `docs/prd/2026-11-24-retry-backoff-roundtripper.md` following the shape of the existing files under `docs/prd/` (a new public API warrants one per `docs/prd/README.md`), and add one line to the `HTTP Client & RoundTrippers` list in `README.md`.

13. Do not modify `NewRoundTripperRetry`, `NewRoundTripperRetryWithSkipStatus` or their behaviour. The new constructor is additive; only the internal helper delegation in requirement 8 touches the existing file.

14. Add a `## Unreleased` entry to `CHANGELOG.md` for the new retry-backoff round-tripper.

15. Run `make precommit` and confirm it passes; walk each numbered requirement against the change.
</requirements>

<constraints>
- Do NOT commit — dark-factory handles git.
- Existing tests must still pass unchanged.
- All paths are repo-relative; no absolute or home-relative paths.
- Use Ginkgo/Gomega for tests; use Counterfeiter mocks from `mocks/` rather than hand-written fakes.
- Wrap errors with `github.com/bborbe/errors`, matching `http_roundtripper-retry.go`.
- Do not add a `//nolint` without an explanation on the same line.
- Do not run `go mod vendor`; the repo does not commit `vendor/`.
</constraints>

<verification>
Run `make precommit` — must pass.
</verification>
