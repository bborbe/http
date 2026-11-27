---
status: completed
summary: Hoisted the X-RateLimit-Remaining parse to a single call in ClassifyRateLimitResponse so the primary and secondary branches share one parse.
execution_id: http-classify-429-exec-008-hoist-remaining-parse-and-update-prd
dark-factory-version: v0.199.2
created: "2026-11-26T23:49:34Z"
queued: "2026-11-26T23:49:34Z"
started: "2026-11-26T23:49:53Z"
completed: "2026-11-26T23:51:08Z"
---

<summary>
- The rate-limit-remaining header is parsed once per call instead of twice, so the two branches cannot drift apart
- Behaviour is unchanged for every classification and retry decision
- The two branches that read the header now share one parse, so a future change cannot apply to only one of them
</summary>

<objective>
Parse the rate-limit-remaining header once instead of twice so the two branches cannot drift.
</objective>

<context>
Read CLAUDE.md for project conventions.

Read `http_roundtripper-retry-backoff.go` — `ClassifyRateLimitResponse` (around line 80). It currently parses `X-RateLimit-Remaining` with `strconv.Atoi` twice: once in the primary branch (`remaining == 0`) and again in the 429 branch (`remaining > 0`).
</context>

<requirements>
1. In `ClassifyRateLimitResponse`, hoist the parse to a single call placed immediately before the primary branch — after the `Retry-After` branch, so an earlier match still returns without parsing. Name the two results `remaining` and `remainingErr`. Both branches then use those variables instead of calling `strconv.Atoi` again.

   Target shape of the tail of the function (note the `Retry-After` branch above it is unchanged):

   ```go
   remaining, remainingErr := strconv.Atoi(resp.Header.Get("X-RateLimit-Remaining"))
   if remainingErr == nil && remaining == 0 {
       return RateLimitClassPrimary
   }
   if resp.StatusCode == http.StatusTooManyRequests {
       if remainingErr == nil && remaining > 0 {
           return RateLimitClassSecondary
       }
       return RateLimitClassUnclassified
   }
   return RateLimitClassNone
   ```

   Behaviour must be identical for every input, including a missing header and an unparseable one. The existing `http_roundtripper-retry-backoff_test.go` cases must still cover the 429-with-positive-remaining → secondary path; if that case is absent, add it before finishing.

2. Do not change any test expectation, any exported signature, `NewRoundTripperRetryWithBackoff`, `computeRetryWait`, or the `MaxWait` handling.

3. Do not touch `docs/` — the PRD's rule list is corrected separately and must not be edited here.

4. Before finishing, re-run `<verification>` and confirm it passes; then walk each numbered requirement against the change and report any that is unmet.
</requirements>

<constraints>
- Do NOT commit — dark-factory handles git
- Existing tests must pass unchanged — this is a pure refactor, so a test that needs editing means behaviour changed and the change is wrong
- Do not reorder any branch: the evaluation order is part of the contract
- Add no dependency
</constraints>

<verification>
Run `ROOTDIR=/workspace make precommit` -- must pass.
</verification>
