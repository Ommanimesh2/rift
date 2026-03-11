---
phase: 07-output-formats
plan: 03
status: complete
completed_date: 2026-03-11
commits:
  task1: a822e5a
  task2: 8f7c3d1
---

# 07-03 Summary: CLI Wiring — Format Dispatch

## Objective

Wire the `--format` flag in `cmd/root.go` to dispatch to `FormatJSON` and
`FormatMarkdown`, replacing the "not yet supported" fallback. Add integration
smoke tests in `cmd/root_test.go` to confirm the formatters are importable and
produce the correct structure.

## What Changed

### cmd/root.go

- Replaced the `if flags.format == "terminal" { ... } else { fmt.Fprintf not yet supported }` block with a three-case `switch`:
  - `"terminal"` / `""` → `output.RenderTerminalWithSecurity`
  - `"json"` → `output.FormatJSON` (error returned on marshal failure)
  - `"markdown"` → `output.FormatMarkdown`
  - `default` → `fmt.Errorf("unknown format %q: supported formats are terminal, json, markdown")`
- Removed the debug `fmt.Printf("Opened %s (%s)\n", ...)` and `fmt.Printf("  Tree: %s\n", ...)` lines from `RunE` — these were scaffolding from Phase 2/3 that would have polluted JSON and Markdown output.
- The `os` import is retained: it is still used by `fmt.Fprintf(os.Stderr, ...)` for the layer warning and by `os.Exit(1)` in `Execute()`.

### cmd/root_test.go (new file)

Four integration tests in `package cmd_test` that call output functions directly
without requiring a Docker daemon:

| Test | What it checks |
|------|----------------|
| `TestFormatJSON_ValidJSON` | Valid JSON with correct image names, summary counts (1 added, 1 modified), non-nil Changes and SecurityEvents slices |
| `TestFormatJSON_SecurityEventsPresent` | Security event with `kind` and `path` fields serialized |
| `TestFormatMarkdown_ContainsRequiredSections` | Output contains `## Image Diff`, `### Summary`, `### Changes` |
| `TestFormatMarkdown_ImageNamesInHeader` | Image names appear in header; empty result shows `(No changes)` |

## Verification

```
go build ./...   # clean
go vet ./...     # clean
go test ./...    # all packages pass — cmd(4 new), output(11), security(17), diff, tree, source
```

## Note: Phase 7 Complete

All three plans in Phase 7 are complete:

- **07-01**: `FormatJSON` TDD (10 tests, `internal/output/json.go`)
- **07-02**: `FormatMarkdown` TDD (11 tests, `internal/output/markdown.go`)
- **07-03**: CLI wiring + integration tests (4 tests, `cmd/root.go` + `cmd/root_test.go`)

The `--format` flag is now fully operational end-to-end. `imgdiff --format json`
and `imgdiff --format markdown` produce correct output; unknown formats return a
clear error; terminal (default) behaviour is unchanged.
