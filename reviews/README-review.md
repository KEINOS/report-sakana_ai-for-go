# README Report Review

- Review date: 2026-06-23
- Target: [`README.md`](../README.md)
- Method: Two-pass read-only review using Copilot, Claude, Agy, Hermes, and Codex
- Final result: All five active reviewers agreed that no blocking findings remained

## Reviewer status

| Reviewer | First pass | Second pass |
| :--- | :--- | :--- |
| Copilot | Reported a Sakura metric reproducibility gap | Agreed |
| Claude | Reported an incorrect API-response count and a completion-limit caveat | Agreed |
| Agy | Reported overgeneralized use-case recommendations and a Modern metric limitation | Agreed |
| Hermes | Agreed without blocking findings | Agreed |
| Codex | Reported the `go fix`/race evaluation order and an ambiguous Pareto candidate set | Agreed |

Agy's first invocation failed because its wrapper could not read a process-substitution path. It was rerun once with a regular prompt file and returned a usable review. Hermes, which had failed during an earlier review session, returned usable results in both passes of this review.

## Accepted findings

- Corrected the stored API-response count from 30 to 15 and distinguished the 15 extracted source files.
- Narrowed the recommended use cases to code generation similar to the tested tasks.
- Removed unsupported recommendations about code-review capability and documented that review capability was not tested.
- Clarified that the three task-level race tests run after `go fix`, while the 20-run ParallelMapOrdered repetition uses unmodified generated code.
- Documented that Modern Syntax is a `go fix`-based proxy with false-positive and false-negative limitations.
- Defined the aggregate Pareto candidate set as models with average Functionality of at least 80%.
- Removed `gpt-oss-120b` from that aggregate candidate set because its runs 2–5 average Functionality is 75%.
- Added direct links to the Sakura runs 2–5 summary files used for the aggregate comparison.
- Documented the observed `completion_tokens` value exceeding the submitted `max_completion_tokens` value.
- Restored the task prompt files to match the Sakura reference prompts exactly after Markdown formatting had altered their whitespace and one `*ParseError` token.

## Rejected or limited findings

- The suggestion to regenerate every race result on unmodified code was not applied. The benchmark intentionally follows the Sakura reference order, and the report now states that task-level race tests use the `go fix` result. The unmodified ParallelMapOrdered code also passed the race-enabled 20-run repetition in every run.
- Style-only wording suggestions were not applied unless they corrected an unsupported conclusion or factual ambiguity.

## Validation

- Recalculated Fugu token totals, response medians, usability, Modern Syntax, and Pareto inputs from the stored JSON files.
- Recalculated Sakura runs 2–5 aggregate values from the sibling repository summary files.
- Confirmed all three task prompts match the Sakura reference files.
- Ran `markdownlint-cli2 README.md API_SPEC.md reviews/README-review.md --fix`.
- Ran a second read-only review after applying accepted findings.
