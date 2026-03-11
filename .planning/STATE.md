# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-03-11)

**Core value:** Be the "git diff" for container images. Beautiful, instant clarity on what changed between two images — file-level, security-aware, and pipeline-ready.
**Current focus:** Phase 3 — File Tree Engine

## Current Position

Phase: 3 of 9 (File Tree Engine)
Plan: Not started
Status: Ready to plan
Last activity: 2026-03-11 — Phase 2 complete (02-01, 02-02)

Progress: ██░░░░░░░░ 22%

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

**Recent Trend:**
- Last 5 plans: 01-01 (15 min), 02-01 (10 min), 02-02 (5 min)
- Trend: Accelerating

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

### Deferred Issues

None yet.

### Blockers/Concerns

None yet.

## Session Continuity

Last session: 2026-03-11
Stopped at: Phase 2 complete — source.Open() wired into CLI, all three providers implemented (remote, daemon, tarball)
Resume file: .planning/phases/02-image-source-access/02-02-SUMMARY.md
