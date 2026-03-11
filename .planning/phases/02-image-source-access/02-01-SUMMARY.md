---
phase: 02-image-source-access
plan: 01
subsystem: source

tags: [go, go-containerregistry, v1-image, remote, daemon, tarball, authn]

# Dependency graph
requires:
  - phase: 01-project-foundation
    provides: Go module, cobra CLI skeleton, internal/ directory
provides:
  - internal/source/ package with Open(ref, opts) returning v1.Image
  - Auto-detection of source type from reference string (remote, daemon, tarball)
  - Remote registry provider with authn.DefaultKeychain and platform support
  - Docker daemon provider with daemon:// prefix stripping
  - OCI tarball provider with nil tag (first image)
  - parsePlatform() helper for "os/arch" format
affects: [02-02-cli-integration, 03-file-tree-engine, 04-diff-engine, 08-performance-optimization]

# Tech tracking
tech-stack:
  added:
    - github.com/google/go-containerregistry v0.21.2 — container image access (remote, daemon, tarball)
  patterns:
    - Unified v1.Image interface — all sources return same type, downstream never knows the source
    - Reference auto-detection — classify by prefix (daemon://), file existence, extension, or default to remote
    - Lazy loading — go-containerregistry downloads layers on access, not on Image() call

key-files:
  created:
    - internal/source/source.go — SourceType enum, DetectSourceType(), parsePlatform(), Open() dispatcher
    - internal/source/remote.go — openRemote() with authn.DefaultKeychain and optional platform
    - internal/source/daemon.go — openDaemon() with daemon:// prefix stripping
    - internal/source/tarball.go — openTarball() with nil tag for first image
    - internal/source/source_test.go — 18 test cases covering detection, string, and platform parsing
  modified:
    - go.mod — added go-containerregistry dependency
    - go.sum — updated checksums

key-decisions:
  - "Use v1.Image directly — no custom wrapper interface"
  - "parsePlatform supports os/arch only (2 parts) — variant not needed for current scope"
  - "Tarball detection checks os.Stat first, then falls back to extension matching"

patterns-established:
  - "Provider pattern: each source type in its own file (remote.go, daemon.go, tarball.go) with unexported openXxx() functions"
  - "Open() as single entry point — dispatches to providers based on DetectSourceType()"
  - "Error wrapping: fmt.Errorf('open %s image %q: %w', sourceType, ref, err) for context"

issues-created: []

# Metrics
duration: 10min
completed: 2026-03-11
---

# Phase 2 Plan 01: Source Abstraction Summary

**go-containerregistry source abstraction with auto-detecting Open() that resolves remote, daemon, and tarball references to v1.Image**

## Performance

- **Duration:** ~10 min
- **Started:** 2026-03-11
- **Completed:** 2026-03-11
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments
- Source package with Open(ref, opts) that auto-detects and delegates to correct provider
- Three providers: remote (with authn + platform), daemon (prefix stripping), tarball (nil tag)
- 18 table-driven test cases covering detection logic, string conversion, and platform parsing — all pass
- go build, go vet, go test all clean

## Task Commits

Each task was committed atomically:

1. **Task 1: Add go-containerregistry and create source types with reference detection** - `c1fe2ac` (feat)
2. **Task 2: Implement tests for detection and platform parsing** - `62f2bfd` (test)

**Plan metadata:** (docs commit below)

## Files Created/Modified
- `go.mod` — added go-containerregistry v0.21.2
- `go.sum` — updated checksums
- `internal/source/source.go` — SourceType enum, DetectSourceType(), parsePlatform(), Open()
- `internal/source/remote.go` — openRemote() with authn.DefaultKeychain and platform
- `internal/source/daemon.go` — openDaemon() with daemon:// stripping
- `internal/source/tarball.go` — openTarball() with nil tag
- `internal/source/source_test.go` — 18 test cases

## Decisions Made
- Used v1.Image directly without wrapping — keeps the abstraction thin and go-containerregistry idiomatic
- parsePlatform() only supports "os/arch" (2-part) format — variant/os.version not needed yet
- Detection priority: daemon:// prefix > os.Stat file check > extension match > remote default

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## Next Phase Readiness
- Plan 02-02 (CLI Integration) can begin — source.Open() is ready to wire into root command
- The v1.Image interface provides Layers(), Digest(), and all metadata needed for confirmation output

---
*Phase: 02-image-source-access*
*Completed: 2026-03-11*
