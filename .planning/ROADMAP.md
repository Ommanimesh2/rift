# Roadmap: imgdiff

## Overview

Build a beautiful, file-level, security-aware container image diff tool in Go. Starting from project scaffolding and CLI, through image pulling and tree construction, to a polished diff engine with security analysis, multiple output formats, and CI/CD integration. The end result is a single static binary that fills the gap left by Google's archived container-diff.

## Milestones

- ✅ **[v1.0 MVP](milestones/v1.0-ROADMAP.md)** — Phases 1–9 (shipped 2026-03-12)
- ✅ **v1.1 Demo** — Phases 10–11 (shipped 2026-03-12)

## Domain Expertise

None

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

- [x] **Phase 1: Project Foundation & CLI** - Go module, cobra CLI skeleton, command parsing, project layout
- [x] **Phase 2: Image Source Access** - go-containerregistry integration for daemon, registry, and tarball sources
- [x] **Phase 3: File Tree Engine** - Parse tar layers into in-memory file trees, handle OCI whiteouts, squash layers
- [x] **Phase 4: Diff Engine** - Compare two file trees, compute added/removed/modified with file metadata
- [x] **Phase 5: Terminal Output** - Color-coded diff with lipgloss, size impact summary, per-layer breakdown (3 plans)
- [x] **Phase 6: Security Analysis** - SUID/SGID detection, permission changes, new executables, world-writable files
- [x] **Phase 7: Output Formats** - JSON and Markdown formatters for CI/CD pipelines and PR comments
- [x] **Phase 8: Performance Optimization** - Shared layer skip, streaming comparison, --quick manifest-only mode
- [x] **Phase 9: CI/CD & Distribution** - Exit codes, multi-arch --platform flag, registry auth support
- [x] **Phase 10: vhs Demo GIF** - Write vhs tape script, produce demo.gif covering file diff, security, and JSON output scenes
- [x] **Phase 11: README Polish** - Embed demo.gif, add badges, tighten copy for first-impression impact

## Phase Details

### Phase 1: Project Foundation & CLI
**Goal**: Working Go binary with cobra CLI that parses two image references and validates arguments
**Depends on**: Nothing (first phase)
**Research**: Unlikely (Go project setup, cobra — established patterns)
**Plans**: 1 plan

Plans:
- [x] 01-01: Go module setup, cobra CLI with flags and argument validation

### Phase 2: Image Source Access
**Goal**: Pull and open container images from local Docker daemon, remote registries, and OCI tarball archives using go-containerregistry
**Depends on**: Phase 1
**Research**: Likely (go-containerregistry API for different image sources)
**Research topics**: go-containerregistry v1.Image interface, daemon vs remote vs tarball access patterns, authentication for private registries
**Plans**: 2 plans

Plans:
- [x] 02-01: Source abstraction and providers (remote, daemon, tarball)
- [x] 02-02: CLI integration — wire source.Open() into root command

### Phase 3: File Tree Engine
**Goal**: Build complete squashed file trees from image layers — parse tar entries, track file metadata (permissions, size, uid/gid), handle OCI whiteout files for layer deletions
**Depends on**: Phase 2
**Research**: Unlikely (tar parsing is standard Go, OCI whiteout spec is simple)
**Plans**: 3 plans

Plans:
- [x] 03-01: FileNode model + single layer tar parsing (TDD)
- [x] 03-02: Whiteout handling + multi-layer squashing (TDD)
- [x] 03-03: BuildFromImage + CLI integration

### Phase 4: Diff Engine
**Goal**: Compare two file trees and produce a structured diff result with added/removed/modified files including size deltas, permission changes, and content hash comparison
**Depends on**: Phase 3
**Research**: Unlikely (internal algorithm, pure Go data structures)
**Plans**: 2 plans

Plans:
- [x] 04-01: DiffEntry/DiffResult types + Diff function (TDD)
- [x] 04-02: CLI integration — wire Diff into root command with summary output

### Phase 5: Terminal Output
**Goal**: Beautiful color-coded terminal diff output using lipgloss — green for added, red for removed, yellow for modified, size impact summary, per-layer breakdown view
**Depends on**: Phase 4
**Research**: Unlikely (lipgloss is well-documented, charmbracelet ecosystem)
**Plans**: 3 plans

Plans:
- [x] 05-01: Diff output text formatter (TDD)
- [x] 05-02: lipgloss terminal styling + CLI integration
- [x] 05-03: Layer comparison and breakdown (TDD)

### Phase 6: Security Analysis
**Goal**: Detect and highlight security-relevant changes — new SUID/SGID binaries, permission escalations, new executables, world-writable files. Add --security-only flag for focused security view
**Depends on**: Phase 4
**Research**: Unlikely (file permission bit analysis — standard Go)
**Plans**: 2 plans

Plans:
- [x] 06-01: SecurityEvent types + Analyze function (TDD)
- [x] 06-02: Terminal output security section + --security-only CLI flag

### Phase 7: Output Formats
**Goal**: JSON output for CI/CD pipeline consumption and Markdown output for PR comments/documentation. --format flag to select output mode
**Depends on**: Phase 5, Phase 6
**Research**: Unlikely (JSON marshaling, Markdown generation — standard patterns)
**Plans**: 3 plans

Plans:
- [x] 07-01: FormatJSON TDD (10 tests)
- [x] 07-02: FormatMarkdown TDD (11 tests)
- [x] 07-03: CLI wiring + integration tests (4 tests)

### Phase 8: Performance Optimization
**Goal**: Skip shared layers with identical digests, stream tar entries without full download to disk, --quick manifest-only mode for instant registry comparison
**Depends on**: Phase 4
**Research**: Likely (manifest-only comparison via go-containerregistry, streaming tar patterns)
**Research topics**: go-containerregistry manifest/descriptor API, comparing layer digests without pulling content, memory-efficient tar streaming
**Plans**: TBD

### Phase 9: CI/CD & Distribution
**Goal**: Non-zero exit codes on size threshold exceeded or security changes detected, --platform flag for multi-arch image selection, registry authentication support
**Depends on**: Phase 7, Phase 8
**Research**: Unlikely (exit codes, flag handling — standard Go/cobra patterns)
**Plans**: TBD

---

### ✅ v1.1 Demo (Shipped 2026-03-12)

**Milestone Goal:** Make the first impression land — produce a demo GIF and polish the README so the tool looks as good as it works.

#### Phase 10: vhs Demo GIF

**Goal**: Write a vhs tape script and produce a compelling demo.gif (~30–60s) covering three scenes: nginx:1.24→1.25 file diff with layer breakdown, ubuntu:22.04→24.04 security events, and a JSON output snippet for CI/CD value
**Depends on**: Phase 9 (v1.0 complete)
**Research**: Likely (vhs is new tooling — need to verify install, tape syntax, and GIF output options)
**Research topics**: charmbracelet/vhs tape syntax, install via Homebrew or Go, deterministic recording options, GIF size optimization, font/theme config for screenshots
**Plans**: TBD

Plans:
- [ ] 10-01: TBD (run /gsd:plan-phase 10 to break down)

#### Phase 11: README Polish

**Goal**: Embed demo.gif at the top of README.md, add badges (Go version, license), tighten tagline and install/CI-CD copy for maximum first-scroll impact
**Depends on**: Phase 10
**Research**: Unlikely (Markdown and badge patterns — established conventions)
**Plans**: TBD

Plans:
- [x] 11-01: GIF embed, badges, tagline, tighter opening paragraph

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8 → 9 → 10 → 11

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|---------------|--------|-----------|
| 1. Project Foundation & CLI | v1.0 | 1/1 | Complete | 2026-03-11 |
| 2. Image Source Access | v1.0 | 2/2 | Complete | 2026-03-11 |
| 3. File Tree Engine | v1.0 | 3/3 | Complete | 2026-03-11 |
| 4. Diff Engine | v1.0 | 2/2 | Complete | 2026-03-11 |
| 5. Terminal Output | v1.0 | 3/3 | Complete | 2026-03-11 |
| 6. Security Analysis | v1.0 | 2/2 | Complete | 2026-03-11 |
| 7. Output Formats | v1.0 | 3/3 | Complete | 2026-03-11 |
| 8. Performance Optimization | v1.0 | 3/3 | Complete | 2026-03-12 |
| 9. CI/CD & Distribution | v1.0 | 2/2 | Complete | 2026-03-12 |
| 10. vhs Demo GIF | v1.1 | 2/2 | Complete | 2026-03-12 |
| 11. README Polish | v1.1 | 1/1 | Complete | 2026-03-12 |
