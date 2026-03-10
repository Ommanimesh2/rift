# Roadmap: imgdiff

## Overview

Build a beautiful, file-level, security-aware container image diff tool in Go. Starting from project scaffolding and CLI, through image pulling and tree construction, to a polished diff engine with security analysis, multiple output formats, and CI/CD integration. The end result is a single static binary that fills the gap left by Google's archived container-diff.

## Domain Expertise

None

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

- [ ] **Phase 1: Project Foundation & CLI** - Go module, cobra CLI skeleton, command parsing, project layout
- [ ] **Phase 2: Image Source Access** - go-containerregistry integration for daemon, registry, and tarball sources
- [ ] **Phase 3: File Tree Engine** - Parse tar layers into in-memory file trees, handle OCI whiteouts, squash layers
- [ ] **Phase 4: Diff Engine** - Compare two file trees, compute added/removed/modified with file metadata
- [ ] **Phase 5: Terminal Output** - Color-coded diff with lipgloss, size impact summary, per-layer breakdown
- [ ] **Phase 6: Security Analysis** - SUID/SGID detection, permission changes, new executables, world-writable files
- [ ] **Phase 7: Output Formats** - JSON and Markdown formatters for CI/CD pipelines and PR comments
- [ ] **Phase 8: Performance Optimization** - Shared layer skip, streaming comparison, --quick manifest-only mode
- [ ] **Phase 9: CI/CD & Distribution** - Exit codes, multi-arch --platform flag, registry auth support

## Phase Details

### Phase 1: Project Foundation & CLI
**Goal**: Working Go binary with cobra CLI that parses two image references and validates arguments
**Depends on**: Nothing (first phase)
**Research**: Unlikely (Go project setup, cobra — established patterns)
**Plans**: TBD

### Phase 2: Image Source Access
**Goal**: Pull and open container images from local Docker daemon, remote registries, and OCI tarball archives using go-containerregistry
**Depends on**: Phase 1
**Research**: Likely (go-containerregistry API for different image sources)
**Research topics**: go-containerregistry v1.Image interface, daemon vs remote vs tarball access patterns, authentication for private registries
**Plans**: TBD

### Phase 3: File Tree Engine
**Goal**: Build complete squashed file trees from image layers — parse tar entries, track file metadata (permissions, size, uid/gid), handle OCI whiteout files for layer deletions
**Depends on**: Phase 2
**Research**: Unlikely (tar parsing is standard Go, OCI whiteout spec is simple)
**Plans**: TBD

### Phase 4: Diff Engine
**Goal**: Compare two file trees and produce a structured diff result with added/removed/modified files including size deltas, permission changes, and content hash comparison
**Depends on**: Phase 3
**Research**: Unlikely (internal algorithm, pure Go data structures)
**Plans**: TBD

### Phase 5: Terminal Output
**Goal**: Beautiful color-coded terminal diff output using lipgloss — green for added, red for removed, yellow for modified, size impact summary, per-layer breakdown view
**Depends on**: Phase 4
**Research**: Unlikely (lipgloss is well-documented, charmbracelet ecosystem)
**Plans**: TBD

### Phase 6: Security Analysis
**Goal**: Detect and highlight security-relevant changes — new SUID/SGID binaries, permission escalations, new executables, world-writable files. Add --security-only flag for focused security view
**Depends on**: Phase 4
**Research**: Unlikely (file permission bit analysis — standard Go)
**Plans**: TBD

### Phase 7: Output Formats
**Goal**: JSON output for CI/CD pipeline consumption and Markdown output for PR comments/documentation. --format flag to select output mode
**Depends on**: Phase 5, Phase 6
**Research**: Unlikely (JSON marshaling, Markdown generation — standard patterns)
**Plans**: TBD

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

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8 → 9

| Phase | Plans Complete | Status | Completed |
|-------|---------------|--------|-----------|
| 1. Project Foundation & CLI | 0/TBD | Not started | - |
| 2. Image Source Access | 0/TBD | Not started | - |
| 3. File Tree Engine | 0/TBD | Not started | - |
| 4. Diff Engine | 0/TBD | Not started | - |
| 5. Terminal Output | 0/TBD | Not started | - |
| 6. Security Analysis | 0/TBD | Not started | - |
| 7. Output Formats | 0/TBD | Not started | - |
| 8. Performance Optimization | 0/TBD | Not started | - |
| 9. CI/CD & Distribution | 0/TBD | Not started | - |
