# imgdiff

## What This Is

A beautiful, file-level, security-aware container image diff tool. Compare two container images and see exactly what changed — files added/removed/modified, size impact, and security-relevant mutations — with color-coded terminal output inspired by `git diff`. Built in Go, designed for both human consumption and CI/CD pipelines.

## Core Value

**Be the "git diff" for container images.** Beautiful, instant clarity on what changed between two images — file-level, security-aware, and pipeline-ready.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] Core diff engine: pull two images (daemon + registry + tarball), build squashed file trees, compute file-level diff (added/removed/modified), handle OCI whiteout files
- [ ] Beautiful terminal output: color-coded diff with lipgloss (green=added, red=removed, yellow=modified), size impact summary, per-layer breakdown
- [ ] Security highlights as core feature: detect new SUID/SGID binaries, permission changes, new executables, world-writable files; `--security-only` flag for focused view
- [ ] Multiple output formats: terminal (default), JSON (CI/CD), Markdown (PR comments/docs)
- [ ] Multi-arch support: default to host platform, `--platform` flag for specific selection
- [ ] `--quick` manifest-only mode: compare layer digests without downloading content (instant for registry images)
- [ ] Shared layer optimization: skip layers with identical digests entirely
- [ ] Streaming comparison: stream tar entries and build file trees in memory, don't download entire layers to disk
- [ ] CI/CD exit codes: non-zero exit on size threshold exceeded or security-relevant changes detected
- [ ] Image source flexibility: support local Docker daemon, remote registries (with auth), and OCI tarball archives

### Out of Scope

- Web UI or dashboard — CLI-only tool, visual appeal comes from terminal output
- Vulnerability scanning — use trivy/grype/syft for that, don't reinvent
- Image optimization/slimming — SlimToolkit does that
- Single-image exploration — dive does that, we do two-image comparison
- Interactive TUI mode — post-MVP feature (bubbletea), not v1
- Package-level diff (apt/apk/pip detection) — post-MVP, v1 is raw file-level diff
- Docker plugin (`docker imgdiff`) — post-MVP distribution channel
- GitHub Action — post-MVP, ship CLI first

## Context

- **Market gap**: dive (53K stars) does single-image exploration. Google's container-diff (3.8K stars) did two-image comparison but was archived March 2024. diffoci (561 stars) is active but focused on reproducible builds with poor UX. Nobody does beautiful, security-aware, file-level two-image diff.
- **Demand signal**: Users have requested image comparison on dive's HN threads since 2018. Google validated the use case by building container-diff. Docker Scout compare exists but is CVE-only, not file-level.
- **Key insight**: The README GIF is the marketing. If the terminal output is visually stunning, it spreads through screenshots and "TIL" moments — same pattern that gave dive 53K stars.
- **Usage pattern**: Infrequent but high-value — debugging image bloat, validating releases, checking base image updates, CI/CD gating, supply chain security, incident response.
- **Target audience**: Platform/DevOps engineers, security-conscious teams, anyone debugging container image size or content changes.

## Constraints

- **Language**: Go — fast compilation, single binary distribution, excellent stdlib, matches ecosystem (dive, crane, skopeo all Go)
- **Core library**: google/go-containerregistry for image access (the standard). Custom lightweight layer-to-filetree built on top of raw layer tar readers (not stereoscope — too heavy, ties to Anchore ecosystem)
- **TUI styling**: charmbracelet/lipgloss for terminal output styling
- **CLI framework**: cobra for CLI argument parsing
- **Distribution**: Single static binary, Homebrew formula, goreleaser for cross-platform releases

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Custom layer-tree over stereoscope | More control, lighter dependency, avoids Anchore ecosystem lock-in | — Pending |
| Security highlights as core v1 feature | Key differentiator vs competitors, unlocks motivated security audience | — Pending |
| All three output formats in v1 (terminal + JSON + markdown) | Covers human, CI/CD, and PR comment use cases from launch | — Pending |
| No interactive TUI in v1 | Focus on perfecting static diff output first, TUI is post-MVP | — Pending |

---
*Last updated: 2026-03-11 after initialization*
