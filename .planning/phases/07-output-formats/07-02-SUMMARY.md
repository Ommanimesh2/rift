---
phase: 07-output-formats
plan: 02
status: complete
completed_date: 2026-03-11
---

# 07-02 Summary: Markdown Output Formatter

## Objective

Implement `FormatMarkdown` in `internal/output/markdown.go` — a GitHub-Flavored Markdown renderer for `DiffResult` suitable for PR comments and documentation.

## What Was Done

### RED (failing tests)
- Created `internal/output/markdown_test.go` with 11 test cases covering:
  - Empty result: summary zeros, `(No changes)` in changes section, no security section
  - One added entry: `+` status, backtick-wrapped path, `+120.6 KB` in summary
  - One removed entry: `-` status, negative size in summary
  - One modified entry: `~` status, `content, mode` in details column
  - Security event with `Before == 0`: empty mode column
  - Security event with mode transition: `` `00755 → 04755` `` in mode column
  - No security section when events nil or empty
  - Full mixed report: all sections present in order (Summary → Changes → Security)
  - Summary dash (`—`) when byte count is zero
  - Exact changes table header and separator rows
  - Single trailing newline

### GREEN (implementation)
- Created `internal/output/markdown.go` with `FormatMarkdown` function:
  - Header: `## Image Diff: \`{image1}\` → \`{image2}\``
  - Summary table with Added/Removed/Modified rows; size shown as `+X KB` / `-X KB` / `—`
  - Changes table with Status (`` `+` ``/`` `-` ``/`` `~` ``), Path (backtick), Size Delta, Details
  - Security Findings section only when events non-empty
  - Reuses `collectFlags`, `securityKindLabel`, `FormatBytes`, `FormatSizeDelta` helpers
- Fixed test assertion bug: `TestFormatMarkdown_SecurityEvent_AddedNoMode` checked for `→` in entire output, but the header always contains it; updated check to look for backtick-octal specifically.

### REFACTOR
No changes needed — implementation was clean on first pass.

## Files Modified

- `internal/output/markdown_test.go` — new test file (11 test cases)
- `internal/output/markdown.go` — new implementation file

## Verification

```
go test ./internal/output/... -run TestFormatMarkdown   # all 11 pass
go build ./...                                           # clean
go vet ./...                                             # clean
```

Pre-existing failure `TestFormatJSON_ModifiedNoFlags` is unrelated (from 07-01) and was failing before this plan.

## Commits

| Hash | Phase | Message |
|------|-------|---------|
| `46723e0` | RED | test(07-02): add failing tests for FormatMarkdown |
| `070acbc` | GREEN | feat(07-02): implement FormatMarkdown for GitHub-Flavored Markdown output |
