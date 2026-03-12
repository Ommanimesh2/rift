# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-03-12)

**Core value:** Be the "git diff" for container images. Beautiful, instant clarity on what changed between two images — file-level, security-aware, and pipeline-ready.
**Current focus:** v1.1 Demo — demo GIF + README polish

## Current Position

Milestone: v1.1 Demo — IN PROGRESS
Phase: 11 of 11 (README Polish)
Plan: Not started
Status: Ready to plan
Last activity: 2026-03-12 — Phase 10 complete, demo.gif produced and committed

Progress: ░░░░░░░░░░ 0% (v1.1)

## Performance Metrics

**Velocity:**
- Total plans completed: 7
- Average duration: ~10 min
- Total execution time: ~0.9 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 1. Project Foundation & CLI | 1 | ~15 min | ~15 min |
| 2. Image Source Access | 2 | ~15 min | ~8 min |
| 3. File Tree Engine | 3 | ~30 min | ~10 min |
| 4. Diff Engine | 2 | ~5 min | ~3 min |
| 5. Terminal Output | 3 | ~30 min | ~10 min |
| 6. Security Analysis | 2 | ~10 min | ~5 min |

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
- **perm_escalation uses 0o7777 mask not 0o777** — SUID/SGID bits are above 0o777; wider mask needed to detect bit additions as escalation
- **RenderTerminalWithLayers delegates to RenderTerminalWithSecurity(nil events)** — nil/empty events skips security section cleanly
- **security.Analyze always returns non-nil slice** — callers can len-check safely without nil guard

### Deferred Issues

None yet.

### Blockers/Concerns

None yet.

### Roadmap Evolution

- Milestone v1.0 MVP shipped: full-featured container image diff tool, 9 phases (Phase 1–9)
- Milestone v1.1 Demo created: demo GIF + README polish, 2 phases (Phase 10–11)

## Session Continuity

Last session: 2026-03-12
Stopped at: Milestone v1.1 Demo initialization
Resume file: None
