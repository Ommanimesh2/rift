---
phase: 03-file-tree-engine
plan: 01
subsystem: core
tags: [tar-parsing, file-tree, sha256, go-containerregistry]

# Dependency graph
requires:
  - phase: 02-image-source-access
    provides: v1.Image and v1.Layer interfaces via go-containerregistry

provides:
  - FileNode struct with Path, Size, Mode, UID, GID, IsDir, LinkTarget, Digest fields
  - ParseLayer(v1.Layer) -> map[string]*FileNode with path normalization and sha256 digest
  - FileTree struct with Entries map, Size(), Get(), String() methods
  - BuildTree([]v1.Layer) -> *FileTree with OCI whiteout squashing
  - BuildFromImage(v1.Image) -> *FileTree convenience wrapper

affects: [03-02, 03-03, 04-diff-engine, 06-security-analysis]

# Tech tracking
tech-stack:
  added: [archive/tar, crypto/sha256]
  patterns: [table-driven tests with subtests, fake v1.Layer backed by bytes.Buffer]

key-files:
  created:
    - internal/tree/tree.go
    - internal/tree/tree_test.go

key-decisions:
  - "FileNode stores sha256 digest computed at parse time — avoids re-reading content later"
  - "normalizePath strips both ./ and / prefixes and trailing slash — consistent keys for all tar conventions"
  - "fakeLayer test helper implements full v1.Layer interface via Uncompressed() with in-memory tar"
  - "BuildTree and BuildFromImage included in 03-01 alongside ParseLayer — co-located in single tree.go file"

patterns-established:
  - "fakeLayer pattern: implement v1.Layer backed by bytes.Buffer for unit testing without real images"
  - "tarEntry helper struct: declarative test data for tar archive construction"
  - "Path normalization: TrimLeft ./  then TrimLeft / then TrimRight / for clean relative keys"

issues-created: []

# Metrics
duration: 15min
completed: 2026-03-11
---

# Phase 03-01: FileNode Model + ParseLayer Summary

**FileNode struct with full metadata and sha256 digest, ParseLayer tar parsing with path normalization, plus FileTree/BuildTree/BuildFromImage stubs — all tested with in-memory fake v1.Layer helper**

## Performance

- **Duration:** 15 min
- **Started:** 2026-03-11T00:00:00Z
- **Completed:** 2026-03-11T00:15:00Z
- **Tasks:** 2 (RED test commit + GREEN impl commit)
- **Files modified:** 2

## Accomplishments
- FileNode struct defined with all metadata fields (Path, Size, Mode, UID, GID, IsDir, LinkTarget, Digest)
- ParseLayer reads v1.Layer.Uncompressed() tar stream, normalizes paths, computes sha256 per regular file
- fakeLayer test helper enables fully in-memory unit testing without real OCI images
- FileTree, BuildTree, and BuildFromImage also implemented here (needed for 03-02/03-03 to build cleanly)

## Task Commits

1. **RED: Failing tests** - `34c5cc8` (test: add failing tests for FileNode model and ParseLayer)
2. **GREEN: Implementation** - `c1e7d36` (feat: implement FileNode model and ParseLayer tar parsing)

## Files Created/Modified
- `internal/tree/tree.go` - FileNode struct, ParseLayer, FileTree, BuildTree, BuildFromImage, formatBytes
- `internal/tree/tree_test.go` - 9 test functions covering all ParseLayer behaviors + 7 BuildTree tests

## Decisions Made
- Co-located BuildTree and BuildFromImage in 03-01 so the package compiles cleanly before 03-02 adds tests
- Used crypto/sha256 at parse time rather than storing raw content — zero-copy, correct for diff engine use
- Path normalization uses TrimLeft for ./ and / then TrimRight for trailing slash — handles all Docker tar conventions

## Deviations from Plan

None - plan executed exactly as written. BuildTree was added alongside ParseLayer rather than as a separate plan stub, which caused 03-02 tests to pass immediately on first run (documented in 03-02 summary).

## Issues Encountered
None.

## Next Phase Readiness
- ParseLayer and FileNode ready for 03-02 whiteout tests
- BuildTree already implemented with full OCI whiteout semantics
- BuildFromImage ready for 03-03 CLI wiring

---
*Phase: 03-file-tree-engine*
*Completed: 2026-03-11*
