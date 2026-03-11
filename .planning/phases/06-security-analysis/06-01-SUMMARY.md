---
phase: 06-security-analysis
plan: 01
status: complete
---

## What Was Built

Created `internal/security/` package with two files:

- `internal/security/security.go` — core detection engine
- `internal/security/security_test.go` — 17 test cases covering all event kinds and edge cases

## Types

**`SecurityEventKind`** (string enum):
- `"new_suid"` — Added file with SUID bit set
- `"new_sgid"` — Added file with SGID bit set
- `"suid_added"` — Modified file gained SUID bit
- `"sgid_added"` — Modified file gained SGID bit
- `"new_executable"` — Added non-directory file with any execute bit
- `"world_writable"` — Added or modified non-directory file with world-write bit
- `"perm_escalation"` — Modified file with strictly more permissive mode

**`SecurityEvent`** struct: `Kind`, `Path`, `Before os.FileMode`, `After os.FileMode`

**`Analyze(result *diff.DiffResult) []SecurityEvent`** — pure function, no I/O.

## Key Design Decision

The `perm_escalation` check uses mask `0o7777` (not `0o777` as stated in the plan spec) so that gaining SUID or SGID bits also triggers a `perm_escalation` event, which is consistent with the test expectations (`Before=0o755, After=0o4755` → both `suid_added` + `perm_escalation`).

## TDD Commits

1. `test(06-01): add failing tests for security analysis` — 17 test cases, RED state
2. `feat(06-01): implement security analysis engine` — all tests pass, GREEN state

## Verification

- `go test ./internal/security/... -v` — 17/17 PASS
- `go build ./...` — OK
- `go vet ./...` — OK
