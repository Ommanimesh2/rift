# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-03-11)

**Core value:** Be the "git diff" for container images. Beautiful, instant clarity on what changed between two images — file-level, security-aware, and pipeline-ready.
**Current focus:** Phase 3 — File Tree Engine

## Current Position

Phase: 4 of 9 (Diff Engine)
Plan: 0 of TBD in current phase
Status: Phase 3 complete, ready for Phase 4
Last activity: 2026-03-11 — Phase 3 complete (03-01, 03-02, 03-03 all done)

Progress: ████░░░░░░ 33%

## Performance Metrics

**Velocity:**
- Total plans completed: 3
- Average duration: ~10 min
- Total execution time: ~0.5 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 1. Project Foundation & CLI | 1 | ~15 min | ~15 min |
| 2. Image Source Access | 2 | ~15 min | ~8 min |
| 3. File Tree Engine | 3 | ~30 min | ~10 min |

**Recent Trend:**
- Last 5 plans: 02-01 (10 min), 02-02 (5 min), 03-01 (15 min), 03-02 (5 min), 03-03 (10 min)
- Trend: Stable ~10 min/plan

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

### Deferred Issues

None yet.

### Blockers/Concerns

None yet.

## Session Continuity

Last session: 2026-03-11
Stopped at: Phase 3 complete — FileNode, ParseLayer, BuildTree, BuildFromImage implemented and tested; CLI prints tree summary
Resume file: .planning/phases/03-file-tree-engine/03-03-SUMMARY.md
