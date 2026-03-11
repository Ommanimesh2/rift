# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-03-11)

**Core value:** Be the "git diff" for container images. Beautiful, instant clarity on what changed between two images — file-level, security-aware, and pipeline-ready.
**Current focus:** Phase 6 — Security Analysis

## Current Position

Phase: 5 of 9 complete (Terminal Output)
Plan: 3/3 in Phase 5 complete
Status: Phase 5 complete, ready for Phase 6
Last activity: 2026-03-11 — Phase 5 complete (05-01 TDD, 05-02 lipgloss+CLI, 05-03 TDD+wiring)

Progress: ██████░░░░ 56%

## Performance Metrics

**Velocity:**
- Total plans completed: 5
- Average duration: ~10 min
- Total execution time: ~0.7 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 1. Project Foundation & CLI | 1 | ~15 min | ~15 min |
| 2. Image Source Access | 2 | ~15 min | ~8 min |
| 3. File Tree Engine | 3 | ~30 min | ~10 min |
| 4. Diff Engine | 2 | ~5 min | ~3 min |
| 5. Terminal Output | 3 | ~30 min | ~10 min |

**Recent Trend:**
- Last 5 plans: 03-01 (15 min), 03-02 (5 min), 03-03 (10 min), 04-01 (3 min), 04-02 (1 min)
- Trend: Stable ~7 min/plan

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- **cobra ExactArgs(2) for argument validation** — defer image reference format validation to Phase 2 when source types are known; cobra handles arity
- **`internal/` directory created but empty** — later phases populate it; keeping skeleton minimal per plan instruction
- **ldflags variables for version/commit/date** — defaults are "dev"/"none"/"unknown" so dev builds are self-describing without a build step
- **Use v1.Image directly — no custom wrapper** — keeps the abstraction thin and go-containerregistry idiomatic
- **parsePlatform supports os/arch only** — variant/os.version not needed for current scope
- **Images are local variables in RunE** — will be passed as arguments to diff functions in later phases, not stored globally
- **FileNode stores sha256 digest at parse time** — avoids re-reading content; digest is ready for Phase 4 diff engine
- **Path normalization strips ./, / prefix and trailing slash** — consistent keys across all Docker/OCI tar conventions
- **fakeLayer test helper pattern** — implement v1.Layer backed by bytes.Buffer for unit testing without real images
- **formatBytes lives in tree.go** — called by FileTree.String(), keeps formatting co-located with the type
- **BuildTree co-located with ParseLayer in 03-01** — single file package, all tree logic in internal/tree/tree.go
- **Diff package in internal/diff/** — separate from tree package; clean boundary between tree construction and comparison
- **DiffEntry has per-attribute change flags** — ContentChanged, ModeChanged, UIDChanged, GIDChanged, LinkTargetChanged, TypeChanged for granular diff reporting
- **formatBytes duplicated in diff package** — tree.formatBytes is unexported; diff package has its own copy (same logic)
- **output.FormatBytes is the single exported version** — replaces the three duplicated unexported copies; lipgloss renderer and layer formatter both reuse it
- **RenderTerminal delegates to RenderTerminalWithLayers(nil)** — backward compatible; nil layerSummary skips the layer section cleanly
- **CompareLayers errors are non-fatal in CLI** — logged to stderr as warning; diff output never blocked by layer inspection failure

### Deferred Issues

None yet.

### Blockers/Concerns

None yet.

## Session Continuity

Last session: 2026-03-11
Stopped at: Phase 5 complete — output package with FormatBytes/FormatEntry/FormatSummary/Render, lipgloss terminal renderer, layer comparison with CompareLayers/FormatLayerSummary, full CLI integration
Resume file: .planning/phases/05-terminal-output/05-03-SUMMARY.md
