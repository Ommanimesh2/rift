---
phase: 05-terminal-output
plan: 01
status: complete
completed: 2026-03-11
---

# 05-01 Summary: Diff Output Text Formatter

## What Was Built

Created `internal/output/format.go` with four exported functions that transform a `*diff.DiffResult` into structured plain-text output lines.

## Functions Delivered

| Function | Description |
|----------|-------------|
| `FormatBytes(n int64) string` | Binary SI conversion: 0→"0 bytes", 1024→"1.0 KB", etc. |
| `FormatSizeDelta(delta int64) string` | Signed delta: 1024→"+1.0 KB", -2048→"-2.0 KB", 0→"0 bytes" |
| `FormatEntry(entry *diff.DiffEntry) string` | Per-entry line with indicator (+/-/~), path, flags, delta |
| `FormatSummary(result *diff.DiffResult) string` | Summary with counts and optional byte parentheticals |
| `Render(result *diff.DiffResult) string` | Full output: entries + blank line + summary |

## TDD Commits

1. **RED**: `test(05-01): add failing tests for output formatter` — 18 tests across 5 test functions
2. **GREEN**: `feat(05-01): implement output formatter` — all 18 tests pass

No REFACTOR commit needed — implementation was clean on first pass.

## Key Design Decisions

- `FormatBytes` is exported so lipgloss renderer (05-02) and layer formatter (05-03) can reuse it
- `FormatSummary` uses `result.AddedBytes` / `result.RemovedBytes` directly (already computed by `diff.Diff`)
- Modified entries show change flags in canonical order: content, mode, uid, gid, link, type
- Byte parentheticals are only emitted when the byte count is non-zero (cleaner output)

## Test Coverage

18 tests in `output_test` external package using table-driven subtests. Covers:
- All six `FormatBytes` boundary cases
- All five `FormatSizeDelta` cases
- Four `FormatEntry` cases (added, removed, modified with 2 flags, modified with all 6 flags)
- Four `FormatSummary` cases (empty, with bytes, no parenthetical, modified-only)
- Four `Render` cases (empty, all entries present, summary last line, entries precede summary)

## Verification

- `go test ./internal/output/... -v` — 18/18 pass
- `go build ./...` — clean
- `go vet ./...` — no issues
- `go test ./...` — all packages pass
