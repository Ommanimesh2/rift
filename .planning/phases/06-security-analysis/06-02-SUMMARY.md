---
phase: 06-security-analysis
plan: 02
status: complete
---

## What was done

Wired security analysis output and CLI integration for the `container-image-diff` project.

### Files modified

**`internal/output/format.go`**
- Added `securityKindLabel(kind SecurityEventKind) string` — internal helper that maps the seven `SecurityEventKind` constants to their display labels (SUID, SGID, SUID ADDED, SGID ADDED, NEW EXEC, WORLD-WRITABLE, PERM ESCALATION).
- Added `FormatSecurityEvent(event security.SecurityEvent) string` — formats a single security event as a one-line string. Added events (Before == 0) omit mode arrows; Modified events show `(0NNN → 0MMM)` in octal.

**`internal/output/terminal.go`**
- Added three new lipgloss styles: `securityHeaderStyle` (bold amber), `securityRedStyle` (red for SUID/SGID), `securityYellowStyle` (yellow for all other kinds).
- Added `RenderTerminalWithSecurity(result, image1Name, image2Name, layerSummary, events)` — extends the terminal renderer with an optional security findings section. If events is nil/empty the output is byte-for-byte identical to `RenderTerminalWithLayers`.
- Added `renderSecuritySection(events)` — internal helper that builds the styled security block with a header (`Security Findings (N)`), a separator line of `━` characters, and one event line per finding.
- Refactored `RenderTerminalWithLayers` to be a thin wrapper that delegates to `RenderTerminalWithSecurity` with nil events (fully backward compatible).

**`cmd/root.go`**
- Added import for `github.com/ommmishra/imgdiff/internal/security`.
- After `diff.Diff(tree1, tree2)`, calls `security.Analyze(result)` to obtain events (pure function, no error path needed).
- Added `--security-only` flag wiring: if no events found, prints "No security findings." and exits; otherwise filters `result.Entries` to only paths present in the events set before rendering.
- Both terminal rendering paths now call `output.RenderTerminalWithSecurity` instead of `output.RenderTerminalWithLayers`.

**`internal/output/terminal_test.go`** (new file)
- 11 new tests covering `FormatSecurityEvent` for all relevant kinds (AddedSUID, AddedSGID, ModifiedPermEscalation, WorldWritable, SUIDAdded, NewExecutable) and `RenderTerminalWithSecurity` (nil events, empty events, single event, multiple events, ordering of security section before summary).

### Verification results

- `go build ./...` — passes
- `go vet ./...` — no issues
- `go test ./internal/output/... -v` — 35 tests, all pass
- `go test ./internal/security/... -v` — 17 tests, all pass (cached, no regressions)
