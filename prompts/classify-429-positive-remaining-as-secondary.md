---
status: draft
---

<summary>
- A 429 response carrying positive rate-limit-remaining quota and no Retry-After header is classified as a secondary rate limit
- Previously that shape was labelled "unclassified", so a rate-limit investigation read the wrong bucket
- The classification order now matches the five-branch evaluation order in the classifier
- A 429 with no remaining header, or a non-positive one, keeps its existing classification
- A plain 403 with positive remaining quota is still not treated as a rate limit
- Retry behaviour is unchanged: both classes were already retried with the same wait
</summary>

<objective>
Make `ClassifyRateLimitResponse` add the fourth branch of its existing evaluation order, so a 429 that carries positive remaining quota and no `Retry-After` header classifies as `secondary` instead of `unclassified`.
</objective>

<context>
Read CLAUDE.md for project conventions.

Read `http_roundtripper-retry-backoff.go` — `ClassifyRateLimitResponse` (around line 75) and the `RateLimitClass` constants above it. The function currently evaluates, in order: body containing "secondary rate limit" → secondary; `Retry-After` present → secondary; `X-RateLimit-Remaining` equal to 0 → primary; any remaining 429 → unclassified; everything else → none.

Read `http_roundtripper-retry-backoff_test.go` — the `retryBackoffCase` table. Two entries currently assert `class=unclassified`: "429 with remaining 10 is unclassified" (around line 200) and "429 with negative remaining is unclassified and retried" (around line 216). The file is a Ginkgo v2 / Gomega suite in package `http_test`, importing the library as `libhttp`.
</context>

<requirements>
1. In `ClassifyRateLimitResponse` in `http_roundtripper-retry-backoff.go`, replace the final 429 branch so a 429 carrying a positive `X-RateLimit-Remaining` returns `RateLimitClassSecondary`, while any other 429 still returns `RateLimitClassUnclassified`. Parse the header exactly as the existing primary branch does (`strconv.Atoi`); a missing or unparseable header must not satisfy the new condition.

   Target shape of the final branches:

   ```go
   if remaining, err := strconv.Atoi(resp.Header.Get("X-RateLimit-Remaining")); err == nil &&
       remaining == 0 {
       return RateLimitClassPrimary
   }
   if resp.StatusCode == http.StatusTooManyRequests {
       if remaining, err := strconv.Atoi(resp.Header.Get("X-RateLimit-Remaining")); err == nil &&
           remaining > 0 {
           return RateLimitClassSecondary
       }
       return RateLimitClassUnclassified
   }
   return RateLimitClassNone
   ```

   Match the existing line wrapping: the current primary branch is already split after `err == nil &&`, and this repository's `make precommit` runs a line-length linter, so keep both new conditions wrapped the same way rather than on one long line.

2. Update the existing table entry "429 with remaining 10 is unclassified" in `http_roundtripper-retry-backoff_test.go`: rename it to state the new expectation and change its asserted log line from `retry attempt=1 status=429 class=unclassified wait=1s` to `retry attempt=1 status=429 class=secondary wait=1s`. Leave its `wantCalls`, `wantWait` and `wantStatus` unchanged — the retry count and the wait are deliberately identical.

3. Leave the "429 with negative remaining is unclassified and retried" entry unchanged. It is the boundary control: `-1` is not greater than zero, so it must keep classifying as `unclassified`.

4. Add a classifier-level test that calls `ClassifyRateLimitResponse` directly rather than going through the round-tripper. Put it in a new file `http_roundtripper-retry-backoff_classify_test.go` in package `http_test`, using the same Ginkgo v2 / Gomega style, the same license header as its neighbours, and a `DescribeTable`. Cover at minimum:
   - 429, `X-RateLimit-Remaining: 12`, no `Retry-After` → `libhttp.RateLimitClassSecondary`
   - 429, `X-RateLimit-Remaining: 0` → `libhttp.RateLimitClassPrimary`
   - 429, no `X-RateLimit-Remaining`, no `Retry-After` → `libhttp.RateLimitClassUnclassified`
   - 403, `X-RateLimit-Remaining: 12`, no `Retry-After` → `libhttp.RateLimitClassNone`
   - 429, `X-RateLimit-Remaining: abc`, no `Retry-After` → `libhttp.RateLimitClassUnclassified` (the malformed-header boundary)
   - 429, `Retry-After: 30` → `libhttp.RateLimitClassSecondary` (the `Retry-After` control, proving the header set is read)

   Build each case as an `*http.Response` with `StatusCode` and a populated `Header`, passing a body slice where the case needs one.

5. Update the `ClassifyRateLimitResponse` doc comment if it enumerates the rules, so it states the new case.

6. Update `CHANGELOG.md` with one `## Unreleased` bullet describing the classification change. If no `## Unreleased` section exists, create one directly above the newest released `## vX.Y.Z` heading. Do not rename or renumber any released heading.

7. Before finishing, re-run `<verification>` and confirm it passes; then walk each numbered requirement against the change and report any that is unmet.
</requirements>

<constraints>
- Do NOT commit — dark-factory handles git
- Existing tests must still pass
- Do not change `NewRoundTripperRetryWithBackoff`, `computeRetryWait`, `MaxWait` handling, or any exported signature
- Do not change the behaviour of `RateLimitClassNone` for statuses other than 403 and 429
- A plain 403 with positive remaining quota must remain `RateLimitClassNone` — the existing "plain 403 is not a rate limit" entry must stay green
- Keep the existing `github.com/bborbe/errors` and `glog` conventions; add no dependency
</constraints>

<verification>
Run `ROOTDIR=/workspace make precommit` -- must pass.
</verification>
