# Phase 9 Plan 02: CLI Wiring Summary

**Wired --exit-code, --fail-on-security, --size-threshold, --username, --password flags into the CLI; os.Exit(2) fires after output when conditions are triggered.**

## Accomplishments

- Added `Username` and `Password` fields to `source.Options`; `openRemote()` uses explicit Basic auth when `Username` is non-empty, falling back to `DefaultKeychain` otherwise
- Registered 5 new flags on `rootCmd`: `--exit-code`, `--fail-on-security`, `--size-threshold`, `--username`, `--password`
- `exitcode.Evaluate()` is called after all output is written in RunE; `os.Exit(2)` fires when any condition is triggered (avoiding cobra's error-path exit 1)
- Invalid `--size-threshold` values return a helpful error via cobra's error return path
- Quick mode (`--quick`) is unaffected — it returns early before exit code evaluation
- 3 new tests in `cmd/root_test.go`: invalid threshold values, valid CLI formats, platform flag regression check
- All tests pass: `go build ./... && go test ./...`

## Task Commits

1. `feat(09-02): add Username/Password fields to source.Options and wire explicit auth in remote.go` — 98f7e95
2. `feat(09-02): wire exit code flags and registry auth into CLI` — eeff703

## Files Created/Modified

- `internal/source/source.go` — added `Username` and `Password` to `Options` struct
- `internal/source/remote.go` — branched on `opts.Username` presence; explicit Basic auth or DefaultKeychain
- `cmd/root.go` — 5 new flags registered in `init()`, `source.Options` wired with credentials, `exitcode.Evaluate()` + `os.Exit(2)` after output
- `cmd/root_test.go` — 3 new test functions covering invalid/valid threshold values and platform regression

## Decisions Made

- `os.Exit(2)` is called directly (not returned as error) so cobra does not print "Error: ..." to stderr — tools should distinguish "tool error" (exit 1) from "condition triggered" (exit 2)
- Threshold parsing error is returned as a cobra error (exit 1) since it is a user input validation failure, not a "differences found" condition
- Quick mode exits before exit code evaluation — manifest-only mode does not produce a full `DiffResult`, so CI/CD conditions cannot be evaluated
- `--username`/`--password` only apply to remote sources; daemon and tarball sources do not use registry credentials

## Deviations from Plan

None. Both tasks executed exactly as specified.

## Issues Encountered

None. All existing tests continued to pass without modification.

## Next Phase Readiness

Phase 9 complete — all planned features shipped: exit codes (--exit-code, --fail-on-security, --size-threshold), --platform (Phase 8), registry auth (DefaultKeychain + explicit --username/--password). Project is feature-complete across all 9 phases.
