# imgdiff

## What This Is

A beautiful, file-level, security-aware container image diff tool. Compare two container images and see exactly what changed — files added/removed/modified, size impact, and security-relevant mutations — with color-coded terminal output inspired by `git diff`. Built in Go, single static binary, designed for both human consumption and CI/CD pipelines.

## Core Value

**Be the "git diff" for container images.** Beautiful, instant clarity on what changed between two images — file-level, security-aware, and pipeline-ready.

## Requirements

### Validated

- ✓ Core diff engine: pull two images (daemon + registry + tarball), build squashed file trees, compute file-level diff (added/removed/modified), handle OCI whiteout files — v1.0
- ✓ Beautiful terminal output: color-coded diff with lipgloss (green=added, red=removed, yellow=modified), size impact summary, per-layer breakdown — v1.0
- ✓ Security highlights as core feature: detect new SUID/SGID binaries, permission changes, new executables, world-writable files; `--security-only` flag for focused view — v1.0
- ✓ Multiple output formats: terminal (default), JSON (CI/CD), Markdown (PR comments/docs) — v1.0
- ✓ Multi-arch support: default to host platform, `--platform` flag for specific selection — v1.0
- ✓ `--quick` manifest-only mode: compare layer digests without downloading content (instant for registry images) — v1.0
- ✓ Shared layer optimization: skip layers with identical digests entirely — v1.0
- ✓ Streaming comparison: stream tar entries and build file trees in memory, don't download entire layers to disk — v1.0
- ✓ CI/CD exit codes: non-zero exit on size threshold exceeded or security-relevant changes detected — v1.0
- ✓ Image source flexibility: support local Docker daemon, remote registries (with auth), and OCI tarball archives — v1.0

### Active (v1.1 targets)

- [ ] Distribution: goreleaser cross-platform releases, GitHub Releases, Homebrew formula
- [ ] GitHub Action: `ommmishra/imgdiff-action` for one-line CI/CD integration
- [ ] Demo GIF in README: the README GIF is the marketing — needs a stunning terminal recording
- [ ] `docker imgdiff` plugin packaging

### Out of Scope

- Web UI or dashboard — CLI-only tool, visual appeal comes from terminal output
- Vulnerability scanning — use trivy/grype/syft for that, don't reinvent
- Image optimization/slimming — SlimToolkit does that
- Single-image exploration — dive does that, we do two-image comparison
- Interactive TUI mode — post-v1.1 feature (bubbletea)
- Package-level diff (apt/apk/pip detection) — post-v1.1, v1 is raw file-level diff

## Context

- **Market gap**: dive (53K stars) does single-image exploration. Google's container-diff (3.8K stars) did two-image comparison but was archived March 2024. diffoci (561 stars) is active but focused on reproducible builds with poor UX. Nobody does beautiful, security-aware, file-level two-image diff.
- **Demand signal**: Users have requested image comparison on dive's HN threads since 2018. Google validated the use case by building container-diff. Docker Scout compare exists but is CVE-only, not file-level.
- **Key insight**: The README GIF is the marketing. If the terminal output is visually stunning, it spreads through screenshots and "TIL" moments — same pattern that gave dive 53K stars.
- **Usage pattern**: Infrequent but high-value — debugging image bloat, validating releases, checking base image updates, CI/CD gating, supply chain security, incident response.
- **Target audience**: Platform/DevOps engineers, security-conscious teams, anyone debugging container image size or content changes.
- **v1.0 state**: 6,333 lines of Go across 9 packages. All core features shipped. No ISSUES.md entries. Ready for distribution.

## Constraints

- **Language**: Go — fast compilation, single binary distribution, excellent stdlib, matches ecosystem (dive, crane, skopeo all Go)
- **Core library**: google/go-containerregistry for image access (the standard). Custom lightweight layer-to-filetree built on top of raw layer tar readers (not stereoscope — too heavy, ties to Anchore ecosystem)
- **TUI styling**: charmbracelet/lipgloss for terminal output styling
- **CLI framework**: cobra for CLI argument parsing
- **Distribution**: Single static binary, goreleaser for cross-platform releases

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Custom layer-tree over stereoscope | More control, lighter dependency, avoids Anchore ecosystem lock-in | ✓ Good — kept implementation simple and dependency-light |
| Security highlights as core v1 feature | Key differentiator vs competitors, unlocks motivated security audience | ✓ Good — security analysis is a clean pure function, easy to extend |
| All three output formats in v1 (terminal + JSON + markdown) | Covers human, CI/CD, and PR comment use cases from launch | ✓ Good — formatters are independent, easy to test |
| No interactive TUI in v1 | Focus on perfecting static diff output first, TUI is post-MVP | ✓ Good — kept scope manageable for v1.0 |
| Use v1.Image directly — no custom wrapper | Keeps the abstraction thin and go-containerregistry idiomatic | ✓ Good |
| FileNode stores sha256 digest at parse time | Avoids re-reading content; digest ready for diff engine immediately | ✓ Good |
| DiffEntry has per-attribute change flags | ContentChanged, ModeChanged, UIDChanged etc. for granular diff reporting | ✓ Good — enables downstream consumers to filter by change type |
| perm_escalation uses 0o7777 mask not 0o777 | SUID/SGID bits are above 0o777; wider mask needed to detect bit additions | ✓ Good — prevented false negatives for SUID/SGID escalation |
| Exit code 2 for condition triggers | Exit 1 reserved for tool errors (cobra convention); exit 2 = "differences found" like diff(1) | ✓ Good |
| os.Exit(2) directly after output (not cobra error return) | Cobra error return causes "Error: ..." stderr + exit 1, defeating purpose | ✓ Good |

---
*Last updated: 2026-03-12 after v1.0 milestone*
