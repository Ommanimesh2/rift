---
phase: 07-output-formats
plan: 01
status: complete
commits:
  red:    1329b2f
  green:  5f4534d
  refactor: ~
---

# 07-01 Summary: JSON Formatter

## What Was Built

`FormatJSON` in `internal/output/json.go` serializes a `*diff.DiffResult` plus
image names and `[]security.SecurityEvent` into 2-space indented JSON bytes.

## Struct Schema

| Struct | JSON key |
|--------|----------|
| `DiffReport` | top-level object |
| `ReportSummary` | `"summary"` |
| `ChangeEntry` | `"changes"[]` |
| `SecurityEntry` | `"security_events"[]` |

Key behaviours locked in by tests:
- `"changes"` array is always present (non-nil) at top level
- `ChangeEntry.Changes` (per-entry flags) is omitted via `omitempty` when empty
- `"security_events"` renders as `[]` not `null` when events is nil
- Modes stored as `uint32` decimal integers (e.g. `0o4755` → `2541`)
- Flag order is canonical: content, mode, uid, gid, link, type

## TDD Stages

### RED — `1329b2f`
Wrote `internal/output/json_test.go` with 10 tests covering:
- Empty result (all-zero summary, `[]` for both arrays)
- Nil events → `[]` (not null) verified via raw JSON string check
- Added entry (type, path, size_delta, no changes key)
- Removed entry (negative size_delta, no changes key)
- Modified entry with content+mode flags
- Modified entry with no flags (Changes field nil via omitempty)
- Security event `new_suid` (before_mode=0, after_mode=2541)
- Security event `perm_escalation` (before_mode=493, after_mode=2541)
- All six flags in canonical order
- 2-space indented output, no trailing newline

### GREEN — `5f4534d`
Implemented `internal/output/json.go`:
- Defined four struct types with snake_case JSON tags
- `FormatJSON` builds `DiffReport` from `DiffResult` + inputs
- `collectFlagSlice` mirrors `collectFlags` but returns `[]string` for JSON
- `make([]SecurityEntry, 0, ...)` guarantees non-nil slice
- Also fixed `TestFormatJSON_ModifiedNoFlags` — original test checked raw JSON
  for absence of `"changes"` substring but the top-level field always exists;
  corrected to unmarshal and check the struct field instead

### REFACTOR — skipped
Code was clean after GREEN. `collectFlagSlice` and `collectFlags` serve distinct
output formats (JSON array vs comma-string); duplication is intentional.

## Files

- `internal/output/json.go` — implementation
- `internal/output/json_test.go` — 10 tests, all passing
