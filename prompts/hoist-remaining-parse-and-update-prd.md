---
status: draft
---

<summary>
- The rate-limit-remaining header is parsed once per call instead of twice, so the two branches cannot drift apart
- The PRD's rule list gains the branch that classifies a 429 with positive remaining quota as secondary
- Behaviour is unchanged for every classification and retry decision
- The header-parse and PRD-rule changes land together, so the code and its spec cannot disagree
- The evaluation order stated in the code and in the PRD now match the shipped classifier
</summary>

<objective>
Parse the rate-limit-remaining header once instead of twice so the two branches cannot drift, and correct the PRD's rule list so it describes the classifier that actually ships.
</objective>

<context>
Read CLAUDE.md for project conventions.

Read `http_roundtripper-retry-backoff.go` — `ClassifyRateLimitResponse` (around line 80). It currently parses `X-RateLimit-Remaining` with `strconv.Atoi` twice: once in the primary branch (`remaining == 0`) and again in the 429 branch (`remaining > 0`).

Read `docs/prd/2026-11-24-retry-backoff-roundtripper.md` — the paragraph under the `ClassifyRateLimitResponse` signature (around line 103) lists the rules and ends with "status `429` → `unclassified`", which no longer matches the code.
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

2. In `docs/prd/2026-11-24-retry-backoff-roundtripper.md`, replace the whole rule-list paragraph (the sentence beginning "Rules are evaluated in order" through "otherwise `none`.") so it reads, in evaluation order: body contains `secondary rate limit` → `secondary`; `Retry-After` present → `secondary`; `X-RateLimit-Remaining` parses to `0` → `primary`; status `429` with `X-RateLimit-Remaining` greater than `0` → `secondary`; any other `429` → `unclassified`; otherwise `none`.

3. Do not change any test expectation, any exported signature, `NewRoundTripperRetryWithBackoff`, `computeRetryWait`, or the `MaxWait` handling.

4. Before finishing, re-run `<verification>` and confirm it passes; then walk each numbered requirement against the change and report any that is unmet.
</requirements>

<constraints>
- Do NOT commit — dark-factory handles git
- Existing tests must pass unchanged — this is a refactor plus a documentation correction, so a test that needs editing means behaviour changed and the change is wrong
- Do not reorder any branch: the evaluation order is part of the contract
- Add no dependency
</constraints>

<verification>
Run `ROOTDIR=/workspace make precommit` -- must pass.
</verification>
